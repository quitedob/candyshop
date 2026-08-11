package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/authsession"
	"candypro/api/internal/pkg/jwtutil"

	modelsAuth "candypro/api/internal/models/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-jwt-secret-that-is-definitely-longer-than-32-chars"

// fakeSessionStore implements authsession.Store so AuthMiddleware can be tested
// without a live Redis. ValidateAccessSession is controlled per-test.
type fakeSessionStore struct {
	usesRedis   bool
	sess        *authsession.AccessSession
	validateErr error
	validateCnt int
}

func (f *fakeSessionStore) UsesRedis() bool { return f.usesRedis }
func (f *fakeSessionStore) SaveRefreshToken(context.Context, *modelsAuth.RefreshToken) error {
	return nil
}
func (f *fakeSessionStore) GetRefreshToken(context.Context, string) (*modelsAuth.RefreshToken, error) {
	return nil, nil
}
func (f *fakeSessionStore) RevokeRefreshToken(context.Context, string) error { return nil }
func (f *fakeSessionStore) DeleteRefreshToken(context.Context, string) error { return nil }
func (f *fakeSessionStore) RevokeAllUserRefreshTokens(context.Context, string) error {
	return nil
}
func (f *fakeSessionStore) SaveAccessSession(context.Context, string, authsession.AccessSession, time.Duration) error {
	return nil
}
func (f *fakeSessionStore) ValidateAccessSession(context.Context, string) (*authsession.AccessSession, error) {
	f.validateCnt++
	return f.sess, f.validateErr
}
func (f *fakeSessionStore) RevokeAccessSession(context.Context, string) error { return nil }
func (f *fakeSessionStore) RevokeAllUserAccessSessions(context.Context, string) error {
	return nil
}

func buildAccessToken(t *testing.T, jti, sub, email, role string) string {
	t.Helper()
	claims := jwt.MapClaims{
		"jti":   jti,
		"sub":   sub,
		"email": email,
		"role":  role,
		"iat":   time.Now().Add(-time.Minute).Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}
	tok, err := jwtutil.GenerateJWT(claims, testJWTSecret)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func authRouter(t *testing.T, sessions authsession.Store) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{JWT: config.JWTConfig{Secret: testJWTSecret}}
	r := gin.New()
	r.Use(AuthMiddleware(cfg, sessions))
	r.GET("/api/v1/user/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"userID": c.GetString("userID"),
			"email":  c.GetString("userEmail"),
			"role":   c.GetString("userRole"),
		})
	})
	return r
}

func serveBearer(r http.Handler, tok string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

// G18 regression: a revoked/missing access session must DENY the request even
// when the JWT signature and claims are fully valid. Pre-fix this returned 200.
func TestAuthMiddleware_RevokedSessionDenied(t *testing.T) {
	store := &fakeSessionStore{
		usesRedis:   true,
		validateErr: authsession.ErrSessionNotFound,
	}
	r := authRouter(t, store)
	rec := serveBearer(r, buildAccessToken(t, "revoked-jti", "u1", "a@b.com", "customer"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for revoked session, got %d (body=%s)", rec.Code, rec.Body.String())
	}
	if store.validateCnt == 0 {
		t.Fatal("expected ValidateAccessSession to be called")
	}
}

// G18: a genuine Redis outage must fail closed (503), never fall back to trusting
// the JWT claims. Pre-fix this returned 200.
func TestAuthMiddleware_StoreUnavailableFailsClosed(t *testing.T) {
	store := &fakeSessionStore{
		usesRedis:   true,
		validateErr: authsession.ErrStoreUnavailable,
	}
	r := authRouter(t, store)
	rec := serveBearer(r, buildAccessToken(t, "jti-1", "u1", "a@b.com", "customer"))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on Redis outage, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// Wrapped infrastructure error (real Redis client wraps the sentinel) also fails closed.
func TestAuthMiddleware_StoreUnavailableWrappedFailsClosed(t *testing.T) {
	store := &fakeSessionStore{
		usesRedis:   true,
		validateErr: fmt.Errorf("%w: connection refused", authsession.ErrStoreUnavailable),
	}
	r := authRouter(t, store)
	rec := serveBearer(r, buildAccessToken(t, "jti-2", "u1", "a@b.com", "customer"))

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 on wrapped Redis error, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// Valid session: token passes, email/role filled from the session when absent.
func TestAuthMiddleware_ValidSessionPasses(t *testing.T) {
	store := &fakeSessionStore{
		usesRedis: true,
		sess:      &authsession.AccessSession{UserID: "u1", Email: "a@b.com", Role: "customer"},
	}
	r := authRouter(t, store)
	rec := serveBearer(r, buildAccessToken(t, "jti-3", "u1", "a@b.com", "customer"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid session, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// Session belonging to another user is rejected.
func TestAuthMiddleware_SessionUserIDMismatchDenied(t *testing.T) {
	store := &fakeSessionStore{
		usesRedis: true,
		sess:      &authsession.AccessSession{UserID: "other-user", Email: "x@y.com", Role: "customer"},
	}
	r := authRouter(t, store)
	rec := serveBearer(r, buildAccessToken(t, "jti-4", "u1", "a@b.com", "customer"))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on session/user mismatch, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// Redis disabled (Postgres fallback): access sessions are not tracked, JWT claims pass.
func TestAuthMiddleware_NoRedisStorePasses(t *testing.T) {
	r := authRouter(t, &fakeSessionStore{usesRedis: false})
	rec := serveBearer(r, buildAccessToken(t, "jti-5", "u1", "a@b.com", "customer"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 without Redis store, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}

// No session store at all: JWT claims pass (unchanged legacy behavior).
func TestAuthMiddleware_NilStorePasses(t *testing.T) {
	r := authRouter(t, nil)
	rec := serveBearer(r, buildAccessToken(t, "jti-6", "u1", "a@b.com", "customer"))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with nil store, got %d (body=%s)", rec.Code, rec.Body.String())
	}
}
