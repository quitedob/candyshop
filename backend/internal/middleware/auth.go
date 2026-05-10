package middleware

import (
	"net/http"
	"strings"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/jwtutil"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies the JWT token in the Authorization header
func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.ErrorResp(c, http.StatusUnauthorized, "auth_header_required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.ErrorResp(c, http.StatusUnauthorized, "auth_header_invalid")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtutil.ValidateJWT(tokenString, cfg.JWT.Secret)
		if err != nil {
			response.ErrorResp(c, http.StatusUnauthorized, "token_invalid")
			c.Abort()
			return
		}

		// R4-09: Reject tokens that carry a purpose claim (e.g. verify_email tokens)
		if purpose, hasPurpose := claims["purpose"]; hasPurpose && purpose != "" {
			response.ErrorResp(c, http.StatusUnauthorized, "token_type_mismatch")
			c.Abort()
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || strings.TrimSpace(sub) == "" {
			response.ErrorResp(c, http.StatusUnauthorized, "token_subject_invalid")
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
