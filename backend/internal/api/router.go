package api

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	adminportalroutes "candypro/api/internal/api/routes/adminportal"
	authscoperoutes "candypro/api/internal/api/routes/authscope"
	publicroutes "candypro/api/internal/api/routes/public"
	systemroutes "candypro/api/internal/api/routes/system"
	userportalroutes "candypro/api/internal/api/routes/userportal"
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"
	"candypro/api/internal/pkg/response"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// RouterWithShutdown contains the router and rate limiter for proper shutdown
type RouterWithShutdown struct {
	Router               *gin.Engine
	Limiter              *middleware.RateLimiter
	AIPublicLimiter      *middleware.RateLimiter
	InquiryPublicLimiter *middleware.RateLimiter
}

// SetupRouter configures and returns the Gin router with all routes and middleware
func SetupRouter(h *handlers.Handlers, cfg *config.Config, db *gorm.DB) *RouterWithShutdown {
	// Set Gin mode based on environment
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// R4-07: Configure trusted proxies to prevent X-Forwarded-For spoofing
	// Only trust localhost/loopback by default; override via TRUSTED_PROXIES env var
	trustedProxies := []string{"127.0.0.1", "::1"}
	if tp := os.Getenv("TRUSTED_PROXIES"); tp != "" {
		trustedProxies = strings.Split(tp, ",")
		for i, p := range trustedProxies {
			trustedProxies[i] = strings.TrimSpace(p)
		}
	}
	if err := router.SetTrustedProxies(trustedProxies); err != nil {
		// Non-fatal: log and continue with default behavior
		log.Printf("Warning: SetTrustedProxies failed, using Gin defaults: %v", err)
	}

	// Custom recovery middleware
	router.Use(middleware.Recovery())

	// Request ID middleware for tracing
	router.Use(middleware.RequestID())

	// Security headers middleware
	router.Use(middleware.SecurityHeaders(cfg.Server.Environment))

	// CORS middleware
	corsConfig := cors.Config{
		AllowOrigins:     cfg.Security.CORSAllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept-Language"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           12 * 60 * 60,
	}
	router.Use(cors.New(corsConfig))

	// Locale middleware — sets "locale" in gin.Context for i18n
	router.Use(middleware.LocaleMiddleware())

	// Rate limiting middleware
	rateLimitHandler, limiter := middleware.RateLimit(&cfg.Security)
	router.Use(rateLimitHandler)

	// Request logging
	router.Use(gin.Logger())

	// 上传文件：敏感子目录禁止直链（见 registerUploadRoutes）
	registerUploadRoutes(router, cfg)

	aiPublicLimiter := middleware.NewRateLimiter(cfg.Security.PublicAIRateLimitPerMinute, time.Minute)
	publicAIMiddlewares := []gin.HandlerFunc{
		middleware.DisablePublicAIRoutes(cfg),
		middleware.PublicAIRateLimit(aiPublicLimiter, cfg.Security.PublicAIRateLimitPerMinute),
	}

	inquiryPublicLimiter := middleware.NewRateLimiter(cfg.Security.PublicInquiryRateLimitPerMinute, time.Minute)
	publicInquiryMiddlewares := []gin.HandlerFunc{
		middleware.PublicInquiryRateLimit(inquiryPublicLimiter, cfg.Security.PublicInquiryRateLimitPerMinute),
	}

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (preferred explicit namespace)
		public := api.Group("/public")
		publicroutes.Register(public, h, publicInquiryMiddlewares...)

		// Auth routes
		auth := api.Group("/auth")
		authscoperoutes.Register(auth, h, cfg)

		// User routes (Protected)
		user := api.Group("/user")
		userportalroutes.Register(user, h, cfg, db)

		// Admin routes (Protected)
		admin := api.Group("/admin")
		adminportalroutes.Register(admin, h, cfg)

		// System routes (preferred explicit namespace)
		system := api.Group("/system")
		systemroutes.Register(system, h, cfg, publicAIMiddlewares...)
	}

	// Swagger documentation (if enabled)
	if cfg.Security.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		aiReady := false
		if h.System != nil {
			aiReady = h.System.IsAgentReady()
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"service":  "candypro-api",
			"version":  "1.0.0",
			"ai_agent": aiReady,
		})
	})

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
	})

	return &RouterWithShutdown{
		Router:               router,
		Limiter:              limiter,
		AIPublicLimiter:      aiPublicLimiter,
		InquiryPublicLimiter: inquiryPublicLimiter,
	}
}
