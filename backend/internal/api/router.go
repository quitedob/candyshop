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
	commonRepo "candypro/api/internal/repository/common"
	productRepo "candypro/api/internal/repository/product"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Cookie CSRF：变更请求校验 Origin（Bearer 客户端跳过）
	router.Use(middleware.CookieCSRFGuard(cfg))

	// Locale middleware — sets "locale" in gin.Context for i18n
	router.Use(middleware.LocaleMiddleware())

	// Rate limiting middleware
	rateLimitHandler, limiter := middleware.RateLimit(&cfg.Security)
	router.Use(rateLimitHandler)

	// Prometheus metrics middleware
	router.Use(middleware.PrometheusMetrics())

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

	// Idempotency middleware: when a request carries an Idempotency-Key header
	// we cache (status, body) per (user, method, path, key). Replays with the
	// same body return the cached response; conflicting bodies get 409 (M-17).
	var idempotency gin.HandlerFunc = func(c *gin.Context) { c.Next() }
	if db != nil {
		idempotency = middleware.Idempotency(commonRepo.NewIdempotencyKeyRepository(db), nil)
	}

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (preferred explicit namespace)
		public := api.Group("/public")
		publicroutes.Register(public, h, cfg, publicInquiryMiddlewares...)

		// Auth routes
		auth := api.Group("/auth")
		authscoperoutes.Register(auth, h, cfg)

		// User routes (Protected)
		user := api.Group("/user")
		user.Use(idempotency)
		userportalroutes.Register(user, h, cfg, db)

		// Admin routes (Protected)
		admin := api.Group("/admin")
		admin.Use(idempotency)
		adminportalroutes.Register(admin, h, cfg)

		// System routes (preferred explicit namespace)
		system := api.Group("/system")
		registerUploadScanHook(system, h.Storage, cfg)
		systemroutes.Register(system, h, cfg, publicAIMiddlewares...)

		// Supplier portal routes（默认关闭，设 ENABLE_SUPPLIER_PORTAL=true 启用）
		if cfg.Security.EnableSupplierPortal && db != nil {
			supplierRepo := productRepo.NewSupplierRepository(db)
			supplier := api.Group("/supplier")
			supplier.Use(middleware.SupplierAuthMiddleware(cfg, supplierRepo))
			{
				supplier.GET("/profile", h.AdminPortal.SupplierGetProfile)
				supplier.PUT("/profile", h.AdminPortal.SupplierUpdateProfile)
				supplier.GET("/purchase-orders", h.AdminPortal.SupplierGetMyPOs)
				supplier.PUT("/purchase-orders/:id/status", h.AdminPortal.SupplierUpdatePOStatus)
			}
		}
	}

	// Swagger documentation (if enabled)
	if cfg.Security.EnableSwagger {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Health check — liveness with optional DB ping.
	router.GET("/health", func(c *gin.Context) {
		status := gin.H{
			"status":  "ok",
			"service": "candypro-api",
			"version": "1.0.0",
		}
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				status["db"] = "down"
				c.JSON(http.StatusServiceUnavailable, status)
				return
			}
			status["db"] = "up"
		}
		c.JSON(http.StatusOK, status)
	})

	// Readiness check — verifies critical dependencies before serving traffic.
	// The DB is a hard dependency: /ready must report 503 whenever the DB is
	// unreachable, even when the optional AI agent hasn't initialized. Only
	// after the DB passes do we return the degraded-200 with the agent warning
	// (G27-a).
	router.GET("/ready", func(c *gin.Context) {
		failures := make([]string, 0)

		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				failures = append(failures, "db_connection_pool")
			} else if err := sqlDB.Ping(); err != nil {
				failures = append(failures, "db_ping")
			}
		} else {
			failures = append(failures, "db_not_initialized")
		}

		if len(failures) > 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "not_ready",
				"failures": failures,
			})
			return
		}

		if h.System != nil && !h.System.IsAgentReady() {
			// AI 为可选依赖，不影响核心 B2B 流量就绪
			status := gin.H{
				"status":   "ready",
				"service":  "candypro-api",
				"version":  "1.0.0",
				"warnings": []string{"ai_agent_not_ready"},
			}
			c.JSON(http.StatusOK, status)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ready",
			"service": "candypro-api",
			"version": "1.0.0",
		})
	})

	// Prometheus metrics endpoint
	router.GET("/metrics", middleware.MetricsAuth(), gin.WrapH(promhttp.Handler()))

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
