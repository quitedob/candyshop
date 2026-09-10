package middleware

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"candypro/api/internal/config"
	modelsAuth "candypro/api/internal/models/auth"
	modelsCommon "candypro/api/internal/models/common"
	"candypro/api/internal/pkg/authsession"
	commonRepo "candypro/api/internal/repository/common"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newIdempotencyTestRepository(t *testing.T) *commonRepo.IdempotencyKeyRepository {
	t.Helper()
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = connection.Close() })
	if err := database.AutoMigrate(&modelsCommon.IdempotencyKey{}); err != nil {
		t.Fatal(err)
	}
	return commonRepo.NewIdempotencyKeyRepository(database)
}

func serveIdempotencyRequest(router http.Handler, resourcePath, requestBody, token string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, resourcePath, strings.NewReader(requestBody))
	request.Header.Set(IdempotencyHeader, "test-operation-key")
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestIdempotency_ConcurrentReservationAndResourceScope(t *testing.T) {
	repository := newIdempotencyTestRepository(t)
	router := gin.New()
	router.Use(func(context *gin.Context) { context.Set("userID", "request-owner"); context.Next() })
	router.Use(Idempotency(repository, nil))
	entered := make(chan struct{})
	release := make(chan struct{})
	var handlerCalls atomic.Int32
	router.POST("/orders/:id", func(context *gin.Context) {
		if handlerCalls.Add(1) == 1 {
			close(entered)
			<-release
		}
		context.JSON(http.StatusCreated, gin.H{"order": context.Param("id")})
	})
	firstResponse := make(chan *httptest.ResponseRecorder, 1)
	go func() { firstResponse <- serveIdempotencyRequest(router, "/orders/first", "original", "") }()
	<-entered
	concurrent := serveIdempotencyRequest(router, "/orders/first", "original", "")
	close(release)
	first := <-firstResponse
	if concurrent.Code != http.StatusConflict || !strings.Contains(concurrent.Body.String(), "idempotency_in_progress") {
		t.Fatalf("concurrent response=%d %s", concurrent.Code, concurrent.Body.String())
	}
	if first.Code != http.StatusCreated {
		t.Fatalf("first response=%d", first.Code)
	}
	replay := serveIdempotencyRequest(router, "/orders/first", "original", "")
	if replay.Code != first.Code || !bytes.Equal(replay.Body.Bytes(), first.Body.Bytes()) || replay.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("replay response=%d %s", replay.Code, replay.Body.String())
	}
	conflict := serveIdempotencyRequest(router, "/orders/first", "changed", "")
	if conflict.Code != http.StatusConflict {
		t.Fatalf("changed body response=%d", conflict.Code)
	}
	queryConflict := serveIdempotencyRequest(router, "/orders/first?mode=changed", "original", "")
	if queryConflict.Code != http.StatusConflict {
		t.Fatalf("changed query response=%d", queryConflict.Code)
	}
	otherResource := serveIdempotencyRequest(router, "/orders/second", "original", "")
	if otherResource.Code != http.StatusCreated || !strings.Contains(otherResource.Body.String(), "second") || handlerCalls.Load() != 2 {
		t.Fatalf("different resource collided: response=%d calls=%d", otherResource.Code, handlerCalls.Load())
	}
}

func TestIdempotency_ReplayRequiresUnrevokedSession(t *testing.T) {
	sessions := &fakeSessionStore{usesRedis: true, sess: &authsession.AccessSession{UserID: "request-owner", Role: modelsAuth.User}}
	router := gin.New()
	router.Use(AuthMiddleware(&config.Config{JWT: config.JWTConfig{Secret: testJWTSecret}}, sessions))
	router.Use(RequireRole(modelsAuth.UserPortal()...))
	router.Use(Idempotency(newIdempotencyTestRepository(t), nil))
	handlerCalls := 0
	router.POST("/orders", func(context *gin.Context) { handlerCalls++; context.JSON(http.StatusCreated, gin.H{"created": true}) })
	token := buildAccessToken(t, "test-session", "request-owner", "test@example.invalid", modelsAuth.User)
	if response := serveIdempotencyRequest(router, "/orders", "original", token); response.Code != http.StatusCreated {
		t.Fatalf("first request=%d", response.Code)
	}
	sessions.sess = nil
	sessions.validateErr = authsession.ErrSessionNotFound
	revoked := serveIdempotencyRequest(router, "/orders", "original", token)
	if revoked.Code != http.StatusUnauthorized || revoked.Header().Get("Idempotency-Replayed") == "true" || handlerCalls != 1 {
		t.Fatalf("revoked replay=%d calls=%d", revoked.Code, handlerCalls)
	}
}

func TestIdempotency_ReservationFailureDoesNotRunHandler(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(func(context *gin.Context) { context.Set("userID", "request-owner"); context.Next() })
	router.Use(Idempotency(commonRepo.NewIdempotencyKeyRepository(database), nil))
	handlerCalled := false
	router.POST("/orders", func(context *gin.Context) { handlerCalled = true })
	response := serveIdempotencyRequest(router, "/orders", "original", "")
	if response.Code != http.StatusServiceUnavailable || handlerCalled {
		t.Fatalf("missing reservation table: status=%d ran=%v", response.Code, handlerCalled)
	}
}

func TestIdempotency_LegacyTemplateKeyDoesNotRepeatEffects(t *testing.T) {
	repository := newIdempotencyTestRepository(t)
	if err := repository.Save(context.Background(), &modelsCommon.IdempotencyKey{
		Key: "test-operation-key", UserID: "request-owner", Method: http.MethodPost,
		Path: "/orders/:id", RequestHash: "legacy-body-fingerprint", StatusCode: http.StatusCreated,
		ExpiresAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.Use(func(requestContext *gin.Context) {
		requestContext.Set("userID", "request-owner")
		requestContext.Next()
	})
	router.Use(Idempotency(repository, nil))
	handlerCalled := false
	router.POST("/orders/:id", func(requestContext *gin.Context) { handlerCalled = true })
	response := serveIdempotencyRequest(router, "/orders/previously-created", "original", "")
	if response.Code != http.StatusConflict || handlerCalled {
		t.Fatalf("legacy rollout request status=%d effects=%v", response.Code, handlerCalled)
	}
}
