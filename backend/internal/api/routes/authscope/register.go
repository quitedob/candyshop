package authscoperoutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"

	"github.com/gin-gonic/gin"
)

// Register wires auth API routes under /api/v1/auth.
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config) {
	group.POST("/register", h.AuthScope.Register)
	group.POST("/login", h.AuthScope.Login)
	group.POST("/refresh", h.AuthScope.RefreshToken)
	group.POST("/forgot-password", h.AuthScope.ForgotPassword)
	group.POST("/reset-password", h.AuthScope.ResetPassword)
	group.POST("/verify-email", h.AuthScope.VerifyEmail)

	authProtected := group.Group("")
	authProtected.Use(middleware.AuthMiddleware(cfg))
	{
		authProtected.POST("/logout", h.AuthScope.Logout)
		authProtected.GET("/me", h.AuthScope.Me)
		authProtected.PUT("/profile", h.AuthScope.UpdateProfile)
		authProtected.PUT("/change-password", h.AuthScope.ChangePassword)
		authProtected.POST("/resend-verification", h.AuthScope.ResendVerificationEmail)
	}
}
