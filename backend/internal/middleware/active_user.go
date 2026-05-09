package middleware

import (
	"net/http"

	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/roles"
	"candypro/api/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequireActiveUser checks that the authenticated user has "active" status.
// This must be placed after AuthMiddleware.
func RequireActiveUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawUserID, exists := c.Get("userID")
		if !exists {
			utils.ErrorResp(c, http.StatusUnauthorized, "auth_required")
			c.Abort()
			return
		}
		userID, ok := rawUserID.(string)
		if !ok || userID == "" {
			utils.ErrorResp(c, http.StatusUnauthorized, "invalid_user_identity")
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
			utils.ErrorResp(c, http.StatusUnauthorized, "user_not_found")
			c.Abort()
			return
		}

		if user.Status != "active" {
			utils.ErrorResp(c, http.StatusForbidden, "account_not_active")
			c.Abort()
			return
		}

		c.Next()
	}
}
