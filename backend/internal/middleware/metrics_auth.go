package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// MetricsAuth protects /metrics with optional METRICS_TOKEN bearer auth.
func MetricsAuth() gin.HandlerFunc {
	token := strings.TrimSpace(os.Getenv("METRICS_TOKEN"))
	return func(c *gin.Context) {
		if token == "" {
			c.Next()
			return
		}
		auth := c.GetHeader("Authorization")
		if auth == "Bearer "+token {
			c.Next()
			return
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
