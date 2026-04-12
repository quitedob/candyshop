package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"candypro/api/internal/config"
	"candypro/api/internal/utils"
)

// AuthMiddleware verifies the JWT token in the Authorization header
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateJWT(tokenString, cfg.JWT.Secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// R4-09: Reject tokens that carry a purpose claim (e.g. verify_email tokens)
		if purpose, hasPurpose := claims["purpose"]; hasPurpose && purpose != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token type not accepted as access token"})
			c.Abort()
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || strings.TrimSpace(sub) == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token subject"})
			c.Abort()
			return
		}

		email, _ := claims["email"].(string)
		role, _ := claims["role"].(string)

		// Set normalized identity to context
		c.Set("userID", sub)
		c.Set("userEmail", email)
		c.Set("userRole", role)

		c.Next()
	}
}
