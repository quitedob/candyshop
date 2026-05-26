package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	modelsCommon "candypro/api/internal/models/common"
	commonrepo "candypro/api/internal/repository/common"

	"github.com/gin-gonic/gin"
)

// IdempotencyHeader is the canonical request header name (Stripe-compatible).
const IdempotencyHeader = "Idempotency-Key"

// idempotencyTTL bounds replay-detection memory; tuned to match typical
// duplicate-submission windows (form double-click, network retry).
const idempotencyTTL = 24 * time.Hour

// IdempotencyKeyResolver returns the user identifier to scope the key under.
// We default to context["userID"] when not provided.
type IdempotencyKeyResolver func(c *gin.Context) string

// Idempotency returns a middleware that captures the response of mutating
// requests carrying an Idempotency-Key header. Replays return the cached
// status + body without re-running the handler (M-17).
//
// Scope (per row): (user_id, method, path, key). Different users / routes /
// methods do NOT collide — a key replayed across endpoints triggers the
// handler again, which is safe because the row is keyed on the route too.
//
// Headers:
//   - Idempotency-Replayed: true|false  (added on every request that carried a key)
//   - Idempotency-Conflict: true        (added on 409 conflict response)
//
// Behaviour matrix:
//   no header         → bypass middleware, handler runs as usual
//   first call w/ key → handler runs; response captured and stored
//   replay same body  → cached response served, handler skipped
//   replay diff body  → 409 Conflict, handler skipped
func Idempotency(repo *commonrepo.IdempotencyKeyRepository, resolver IdempotencyKeyResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		if repo == nil {
			c.Next()
			return
		}
		key := strings.TrimSpace(c.GetHeader(IdempotencyHeader))
		if key == "" {
			c.Next()
			return
		}
		// Only mutating verbs are eligible.
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			c.Next()
			return
		}
		userID := ""
		if resolver != nil {
			userID = resolver(c)
		}
		if userID == "" {
			userID = c.GetString("userID")
		}
		if userID == "" {
			// Anonymous requests are not eligible — we can't safely scope the row.
			c.Next()
			return
		}

		// Read body fully so we can hash it and replay on cache hit.
		var bodyBytes []byte
		if c.Request.Body != nil {
			b, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid_body"})
				return
			}
			bodyBytes = b
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		hash := sha256.Sum256(bodyBytes)
		hashHex := hex.EncodeToString(hash[:])

		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Cache hit?
		existing, err := repo.Find(c.Request.Context(), userID, method, path, key)
		if err == nil && existing != nil {
			if existing.RequestHash != hashHex {
				c.Header("Idempotency-Replayed", "false")
				c.Header("Idempotency-Conflict", "true")
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{
					"error":   "idempotency_conflict",
					"message": "Idempotency-Key has already been used with a different request body",
				})
				return
			}
			c.Header("Idempotency-Replayed", "true")
			if existing.ResponseCType != "" {
				c.Header("Content-Type", existing.ResponseCType)
			}
			c.Data(existing.StatusCode, existing.ResponseCType, []byte(existing.ResponseBody))
			c.Abort()
			return
		}

		// Capture response
		writer := &capturingResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Header("Idempotency-Replayed", "false")
		c.Next()

		// Only cache successful responses (2xx). Errors leave the row absent so
		// the client can retry safely.
		status := writer.Status()
		if status < 200 || status >= 300 {
			return
		}
		rec := &modelsCommon.IdempotencyKey{
			Key:           key,
			UserID:        userID,
			Method:        method,
			Path:          path,
			RequestHash:   hashHex,
			StatusCode:    status,
			ResponseBody:  writer.body.String(),
			ResponseCType: writer.Header().Get("Content-Type"),
			CreatedAt:     time.Now(),
			ExpiresAt:     time.Now().Add(idempotencyTTL),
		}
		if err := repo.Save(c.Request.Context(), rec); err != nil && !errors.Is(err, commonrepo.ErrIdempotencyKeyConflict) {
			// Non-fatal: the user already received their successful response.
			// Logging avoids silent loss of a duplicate-prevention record.
			c.Header("Idempotency-Cache-Error", err.Error())
		}
	}
}

// capturingResponseWriter mirrors gin.ResponseWriter while keeping a copy of
// the body for the idempotency cache.
type capturingResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *capturingResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *capturingResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}
