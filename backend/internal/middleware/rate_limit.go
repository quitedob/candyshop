package middleware

import (
	modelsProduct "candypro/api/internal/models/product"
)

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"candypro/api/internal/config"

	"github.com/gin-gonic/gin"
)

// maxRateLimitEntries is the upper bound on tracked IPs to limit memory growth
// under high traffic. Once this cap is hit, old entries are evicted eagerly.
const maxRateLimitEntries = 100_000

// RateLimiter stores rate limit information per IP
type RateLimiter struct {
	requests map[string]*clientInfo
	mu       sync.RWMutex
	limit    int
	window   time.Duration
	stopCh   chan struct{}
}

type clientInfo struct {
	count     int
	firstSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*clientInfo),
		limit:    limit,
		window:   window,
		stopCh:   make(chan struct{}),
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Stop stops the cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

// Allow checks if a request from the given IP is allowed
func (rl *RateLimiter) Allow(ip string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[ip]

	if !exists || now.Sub(info.firstSeen) > rl.window {
		// If we've hit the max entries cap, do an inline eviction pass
		if len(rl.requests) >= maxRateLimitEntries {
			rl.evictExpiredLocked(now)
		}
		rl.requests[ip] = &clientInfo{
			count:     1,
			firstSeen: now,
		}
		return true, rl.limit - 1
	}

	if info.count >= rl.limit {
		return false, 0
	}

	info.count++
	return true, rl.limit - info.count
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			rl.evictExpiredLocked(time.Now())
			rl.mu.Unlock()
		case <-rl.stopCh:
			return
		}
	}
}

// evictExpiredLocked removes entries older than the window. Must be called with mu held.
func (rl *RateLimiter) evictExpiredLocked(now time.Time) {
	for ip, info := range rl.requests {
		if now.Sub(info.firstSeen) > rl.window {
			delete(rl.requests, ip)
		}
	}
}

// RateLimit returns a rate limiting middleware
func RateLimit(cfg *config.SecurityConfig) (gin.HandlerFunc, *RateLimiter) {
	limiter := NewRateLimiter(cfg.RateLimitPerMinute, time.Minute)

	return func(c *gin.Context) {
		ip := c.ClientIP()

		allowed, remaining := limiter.Allow(ip)

		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RateLimitPerMinute))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))

		if !allowed {
			c.JSON(http.StatusTooManyRequests, modelsProduct.ErrorResponse{
				Error:   "rate_limit_exceeded",
				Message: "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}, limiter
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		c.Header("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		// SEC-11: Content-Security-Policy — allow Google Fonts for frontend styling
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self'; frame-ancestors 'none'")
		c.Next()
	}
}

// RequestID ensures every request has a unique correlation ID for tracing.
// If the client sends X-Request-ID, it is reused; otherwise a new one is generated.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID produces a 16-char hex random ID (8 bytes of entropy).
func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; fall back to timestamp
		return time.Now().Format("20060102150405.000000")
	}
	return hex.EncodeToString(b)
}

// Recovery returns a panic recovery middleware
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// L9: Log stack trace for debugging
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				log.Printf("PANIC recovered: %v\n%s", err, buf[:n])
				c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
					Error:   "internal_error",
					Message: "An unexpected error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
