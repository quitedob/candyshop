package api

import (
	"net/http"
	"os"
	"strings"

	adminportalroutes "candypro/api/internal/api/routes/adminportal"
	authscoperoutes "candypro/api/internal/api/routes/authscope"
	publicroutes "candypro/api/internal/api/routes/public"
	systemroutes "candypro/api/internal/api/routes/system"
	userportalroutes "candypro/api/internal/api/routes/userportal"
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// RouterWithShutdown contains the router and rate limiter for proper shutdown
type RouterWithShutdown struct {
	Router  *gin.Engine
	Limiter *middleware.RateLimiter
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
		_ = err
	}

	// Custom recovery middleware
	router.Use(middleware.Recovery())

	// Request ID middleware for tracing
	router.Use(middleware.RequestID())

	// Security headers middleware
	router.Use(middleware.SecurityHeaders())

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

	// Rate limiting middleware
	rateLimitHandler, limiter := middleware.RateLimit(&cfg.Security)
	router.Use(rateLimitHandler)

	// Request logging
	router.Use(gin.Logger())

	// Static files for uploads
	router.Static("/uploads", cfg.Upload.UploadPath)

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (preferred explicit namespace)
		public := api.Group("/public")
		publicroutes.Register(public, h)

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
		systemroutes.Register(system, h, cfg)
	}

	// Swagger documentation (if enabled)
	if cfg.Security.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "candypro-api",
			"version": "1.0.0",
		})
	})

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "The requested resource was not found",
		})
	})

	return &RouterWithShutdown{
		Router:  router,
		Limiter: limiter,
	}
}
