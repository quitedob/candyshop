package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequirePermission restricts access to users with specific permissions
// Note: Requires loading user permissions into context beforehand
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPermissionsInter, exists := c.Get("userPermissions")
		if !exists {
			// If not using precise permissions, we can fallback to role checks or just block
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: no permissions loaded"})
			c.Abort()
			return
		}

		userPerms, ok := userPermissionsInter.([]interface{}) // Assume parsed from JSON/JWT or DB
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: invalid permissions format"})
			c.Abort()
			return
		}

		// Convert to string slice
		perms := make([]string, len(userPerms))
		for i, p := range userPerms {
			perms[i] = p.(string)
		}

		// Check if user has AT LEAST ONE of the required permissions
		for _, reqPerm := range permissions {
			for _, userPerm := range perms {
				if userPerm == reqPerm || userPerm == "*" {
					c.Next()
					return
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
		c.Abort()
	}
}
