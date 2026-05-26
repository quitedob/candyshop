package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"candypro/api/internal/config"

	"github.com/gin-gonic/gin"
)

func TestCookieCSRFGuard_AllowsMatchingOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.Server.Environment = "production"
	cfg.Security.CORSAllowedOrigins = []string{"https://app.example.com"}

	r := gin.New()
	r.Use(CookieCSRFGuard(cfg))
	r.POST("/api/v1/user/profile", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/profile", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestCookieCSRFGuard_BlocksBadOriginInProduction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.Server.Environment = "production"
	cfg.Security.CORSAllowedOrigins = []string{"https://app.example.com"}

	r := gin.New()
	r.Use(CookieCSRFGuard(cfg))
	r.POST("/api/v1/user/profile", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/profile", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestCookieCSRFGuard_SkipsBearerAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{}
	cfg.Server.Environment = "production"
	cfg.Security.CORSAllowedOrigins = []string{"https://app.example.com"}

	r := gin.New()
	r.Use(CookieCSRFGuard(cfg))
	r.POST("/api/v1/user/profile", func(c *gin.Context) { c.Status(http.StatusOK) })

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/profile", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
