package middleware

import (
	"candypro/api/internal/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole restricts access to specific roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("userRole")
		if !exists {
			response.ErrorResp(c, http.StatusUnauthorized, "role_not_found")
			c.Abort()
			return
		}

		roleStr, ok := userRole.(string)
		if !ok {
			response.ErrorResp(c, http.StatusUnauthorized, "invalid_role_format")
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		response.ErrorResp(c, http.StatusForbidden, "forbidden_insufficient_role")
		c.Abort()
	}
}
