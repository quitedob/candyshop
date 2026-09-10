package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"candypro/api/internal/config"
	modelsCommon "candypro/api/internal/models/common"
	"candypro/api/internal/pkg/authcookie"
	"candypro/api/internal/pkg/jwtutil"
	commonrepo "candypro/api/internal/repository/common"

	"github.com/gin-gonic/gin"
)

// IdempotencyHeader is the canonical request header name (Stripe-compatible).
const IdempotencyHeader = "Idempotency-Key"

// idempotencyTTL bounds replay-detection memory; tuned to match typical
// duplicate-submission windows (form double-click, network retry).
const idempotencyTTL = 24 * time.Hour

// idempotencyPersistTimeout bounds recording the result after a client disconnect.
const idempotencyPersistTimeout = 5 * time.Second

// IdempotencyKeyResolver optionally refines the authenticated key scope.
// A verified context["userID"] is always required, even when a resolver is supplied.
type IdempotencyKeyResolver func(c *gin.Context) string

// JWTUserIDResolver returns an IdempotencyKeyResolver that derives the user id from the
// JWT directly. This helper does not authorize a request or check session revocation.
// Idempotency must still run after AuthMiddleware and all route authorization.
func JWTUserIDResolver(cfg *config.Config) IdempotencyKeyResolver {
	return func(c *gin.Context) string {
		if uid := c.GetString("userID"); uid != "" {
			return uid
		}
		token := authcookie.TokenFromRequest(c)
		if token == "" {
			return ""
		}
		claims, err := jwtutil.ValidateJWT(token, cfg.JWT.Secret)
		if err != nil {
			return ""
		}
		sub, _ := claims["sub"].(string)
		return sub
	}
}

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
//
//	no header         → bypass middleware, handler runs as usual
//	first call w/ key → reserve key, run handler, store completed response
//	in-flight replay → 409 Conflict, handler skipped
//	replay same request → cached response served, handler skipped
//	replay changed body/query → 409 Conflict, handler skipped
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
		userID := c.GetString("userID")
		if userID == "" {
			// Require a verified authentication context before looking up any key.
			c.Next()
			return
		}
		if resolver != nil {
			if resolvedUserID := resolver(c); resolvedUserID != "" {
				userID = resolvedUserID
			}
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
		// Query arguments can change a mutation even when its body is identical.
		requestHash := sha256.Sum256([]byte(c.Request.URL.RawQuery + "\n" + hashHex))
		hashHex = hex.EncodeToString(requestHash[:])

		method := c.Request.Method
		// Use the concrete resource path: /orders/one and /orders/two must not
		// share a cached result just because Gin maps both to /orders/:id.
		path := c.Request.URL.EscapedPath()
		if routeTemplate := c.FullPath(); routeTemplate != "" && routeTemplate != path {
			// Pre-fix records collapsed concrete resources into the route template.
			// Their resource identity cannot be recovered, so reject this old key
			// during rollout rather than replaying it or repeating its effects.
			legacyRecord, err := repo.Find(c.Request.Context(), userID, method, routeTemplate, key)
			if err != nil {
				log.Printf("idempotency legacy lookup failed: %v", err)
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "idempotency_unavailable"})
				return
			}
			if legacyRecord != nil {
				c.Header("Idempotency-Conflict", "true")
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "idempotency_conflict"})
				return
			}
		}
		rec := &modelsCommon.IdempotencyKey{
			Key: key, UserID: userID, Method: method, Path: path,
			RequestHash: hashHex, StatusCode: modelsCommon.IdempotencyStatusPending,
			CreatedAt: time.Now(), ExpiresAt: time.Now().Add(idempotencyTTL),
		}
		existing, err := repo.Reserve(c.Request.Context(), rec)
		if err != nil {
			c.Header("Idempotency-Replayed", "false")
			switch {
			case errors.Is(err, commonrepo.ErrIdempotencyKeyConflict):
				c.Header("Idempotency-Conflict", "true")
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "idempotency_conflict"})
			case errors.Is(err, commonrepo.ErrIdempotencyInProgress):
				c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": "idempotency_in_progress"})
			default:
				log.Printf("idempotency reservation failed: %v", err)
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "idempotency_unavailable"})
			}
			return
		}
		if existing != nil {
			c.Header("Idempotency-Replayed", "true")
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

		// Cache every completed outcome: a handler may commit an effect before
		// returning an error, so rerunning failures is not necessarily safe.
		rec.StatusCode = writer.Status()
		rec.ResponseBody = writer.body.String()
		rec.ResponseCType = writer.Header().Get("Content-Type")
		// Client disconnects must not discard an already committed result.
		persistContext, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), idempotencyPersistTimeout)
		defer cancel()
		if err := repo.Save(persistContext, rec); err != nil {
			// Retain the pending claim on persistence failure, preventing repeats.
			log.Printf("idempotency response persistence failed: %v", err)
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
