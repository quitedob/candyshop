package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		allowed, _ := limiter.Allow("192.168.1.1")
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be blocked
	allowed, _ := limiter.Allow("192.168.1.1")
	if allowed {
		t.Error("4th request should be blocked")
	}

	// Different IP should be allowed
	allowed2, _ := limiter.Allow("192.168.1.2")
	if !allowed2 {
		t.Error("Request from different IP should be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := &config.SecurityConfig{
		RateLimitPerMinute: 2,
	}

	router := gin.New()
	rateLimitHandler, limiter := RateLimit(cfg)
	defer limiter.Stop()
	router.Use(rateLimitHandler)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// First 2 requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	// 3rd request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	router := gin.New()
	router.Use(SecurityHeaders("development"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expectedHeaders := map[string]string{
		"X-Frame-Options":        "DENY",
		"X-Content-Type-Options": "nosniff",
		"X-Xss-Protection":       "1; mode=block",
	}

	for header, expected := range expectedHeaders {
		if got := w.Header().Get(header); got != expected {
			t.Errorf("Header %s = %s, want %s", header, got, expected)
		}
	}

	// HSTS should NOT be set in non-production
	if hsts := w.Header().Get("Strict-Transport-Security"); hsts != "" {
		t.Errorf("HSTS should not be set in development, got %s", hsts)
	}
}

func TestSecurityHeadersProduction(t *testing.T) {
	router := gin.New()
	router.Use(SecurityHeaders("production"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := "max-age=63072000; includeSubDomains; preload"
	if got := w.Header().Get("Strict-Transport-Security"); got != expected {
		t.Errorf("HSTS = %s, want %s", got, expected)
	}
}

func TestRequestID(t *testing.T) {
	router := gin.New()
	router.Use(RequestID())
	router.GET("/test", func(c *gin.Context) {
		requestID, _ := c.Get("RequestID")
		c.JSON(200, gin.H{"requestId": requestID})
	})

	// Test auto-generated request ID
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if rid := w.Header().Get("X-Request-ID"); rid == "" {
		t.Error("Expected X-Request-ID header to be set")
	}

	// Test client-provided request ID
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Request-ID", "my-custom-id")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if got := w2.Header().Get("X-Request-ID"); got != "my-custom-id" {
		t.Errorf("Expected X-Request-ID = 'my-custom-id', got %s", got)
	}
}

func TestRecovery(t *testing.T) {
	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
