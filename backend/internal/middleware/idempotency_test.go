package middleware

import (
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/jwtutil"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func testResolverConfig() *config.Config {
	return &config.Config{JWT: config.JWTConfig{Secret: "test-secret-for-idempotency-resolver"}}
}

func TestJWTUserIDResolver_ReadsSubFromBearerToken(t *testing.T) {
	cfg := testResolverConfig()
	token, err := jwtutil.GenerateJWT(jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
	}, cfg.JWT.Secret)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/user/orders", nil)
	c.Request.Header.Set("Authorization", "Bearer "+token)

	got := JWTUserIDResolver(cfg)(c)
	if got != "user-123" {
		t.Fatalf("expected user-123, got %q", got)
	}
}

func TestJWTUserIDResolver_PrefersContextWhenSet(t *testing.T) {
	cfg := testResolverConfig()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/user/orders", nil)
	c.Set("userID", "from-context")

	if got := JWTUserIDResolver(cfg)(c); got != "from-context" {
		t.Fatalf("expected from-context, got %q", got)
	}
}

func TestJWTUserIDResolver_EmptyForMissingToken(t *testing.T) {
	cfg := testResolverConfig()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/user/orders", nil)

	if got := JWTUserIDResolver(cfg)(c); got != "" {
		t.Fatalf("expected empty for missing token, got %q", got)
	}
}

func TestJWTUserIDResolver_EmptyForInvalidToken(t *testing.T) {
	cfg := testResolverConfig()
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/user/orders", nil)
	c.Request.Header.Set("Authorization", "Bearer not-a-real-token")

	if got := JWTUserIDResolver(cfg)(c); got != "" {
		t.Fatalf("expected empty for invalid token, got %q", got)
	}
}
