package authscoperoutes

import (
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// Register wires auth API routes under /api/v1/auth.
func Register(group *gin.RouterGroup, h *handlers.Handlers, cfg *config.Config) {
	group.POST("/register", middleware.AuthEndpointRateLimit(3, time.Minute, cfg.Security.RedisURL), h.AuthScope.Register)
	group.POST("/login", middleware.AuthEndpointRateLimit(5, time.Minute, cfg.Security.RedisURL), h.AuthScope.Login)
	group.POST("/refresh", middleware.AuthEndpointRateLimit(10, time.Minute, cfg.Security.RedisURL), h.AuthScope.RefreshToken)
	group.POST("/forgot-password", middleware.AuthEndpointRateLimit(2, time.Minute, cfg.Security.RedisURL), h.AuthScope.ForgotPassword)
	// Reset-password and verify-email accept opaque tokens — without per-IP throttling
	// they're vulnerable to token brute-force and verification-link spam (R2 C-1).
	// 5/min matches the login limit and is conservative enough for legitimate retries
	// (typo in token, slow form submit) without enabling enumeration-grade scanning.
	group.POST("/reset-password", middleware.AuthEndpointRateLimit(5, time.Minute, cfg.Security.RedisURL), h.AuthScope.ResetPassword)
	group.POST("/verify-email", middleware.AuthEndpointRateLimit(5, time.Minute, cfg.Security.RedisURL), h.AuthScope.VerifyEmail)
	group.POST("/resend-verification-public", middleware.AuthEndpointRateLimit(2, time.Minute, cfg.Security.RedisURL), h.AuthScope.ResendVerificationEmailPublic)

	authProtected := group.Group("")
	authProtected.Use(middleware.AuthMiddleware(cfg, h.AuthScope.SessionStore()))
	{
		authProtected.POST("/logout", h.AuthScope.Logout)
		authProtected.GET("/me", h.AuthScope.Me)
		authProtected.PUT("/profile", h.AuthScope.UpdateProfile)
		authProtected.PUT("/change-password", h.AuthScope.ChangePassword)
		authProtected.POST("/resend-verification", h.AuthScope.ResendVerificationEmail)
	}
}
