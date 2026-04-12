package middleware

import (
	"net/http"

	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/roles"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireActiveUser checks that the authenticated user has "active" status.
// This must be placed after AuthMiddleware.
func RequireActiveUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawUserID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}
		userID, ok := rawUserID.(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user identity"})
			c.Abort()
			return
		}

		// Admin and superadmin bypass the active check
		role, _ := c.Get("userRole")
		if roleStr, ok := role.(string); ok {
			for _, adminRole := range roles.AdminPortal() {
				if roleStr == adminRole {
					c.Next()
					return
				}
			}
		}

		var user modelsUser.User
		if err := db.WithContext(c.Request.Context()).
			Where("id = ?", userID).
			Select("id, status").
			First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		if user.Status != "active" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "account_not_active",
				"message": "Your account is pending approval. Please wait for verification.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
