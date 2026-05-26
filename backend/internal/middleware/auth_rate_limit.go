package middleware

import (
	"net/http"
	"time"

	"candypro/api/internal/pkg/ratelimit"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthEndpointRateLimit applies per-IP limits on sensitive auth endpoints.
func AuthEndpointRateLimit(limit int, window time.Duration, redisURL string) gin.HandlerFunc {
	store := ratelimit.NewFromEnv(redisURL)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !store.Allow(ip, limit, window) {
			response.ErrorResp(c, http.StatusTooManyRequests, "rate_limited")
			c.Abort()
			return
		}
		c.Next()
	}
}
