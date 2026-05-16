package main

import (
	_ "candypro/api/docs"
	repositoryCommon "candypro/api/internal/repository/common"
	servicesCommon "candypro/api/internal/services/common"
	"context"
	"log"
	stdhttp "net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"candypro/api/internal/api"
	"candypro/api/internal/config"
	"candypro/api/internal/database"
	"candypro/api/internal/handlers"
	"candypro/api/internal/pkg/i18n"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

const (
	defaultOrderDraftCleanupIntervalMinutes = 5
	defaultOrderDraftExpireMinutes          = 60
	defaultOrderDraftCleanupBatchSize       = 100
)

// @title CandyPro OEM API
// @version 1.0
// @description API for CandyPro OEM candy manufacturing website
// @contact.name API Support
// @contact.email info@candypro.com
// @host localhost:8080
// @BasePath /api/v1
// @schemes http https
func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {

		log.Fatalf("Failed to load configuration: %v", err)
	}
	// Connect to database
	var db *gorm.DB
	db, err = database.Connect(&cfg.Database)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
		log.Println("Running without database - some features may not work")
	} else {
		// Run migrations
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		// Always seed essential system data (roles, superadmin)
		if err := database.SeedEssential(db); err != nil {
			log.Printf("Warning: Failed to seed essential data: %v", err)
		}

		// Seed i18n translations into DB and initialize the translation engine
		if err := database.SeedTranslations(db, "internal/database/seed_locales"); err != nil {
			log.Printf("Warning: Failed to seed translations: %v", err)
		}
		if err := i18n.Init(db); err != nil {
			log.Printf("Warning: Failed to initialize i18n engine: %v", err)
		}

		// Optional: seed sample business data (products, categories, blog posts, etc.)
		// Controlled by AUTO_SEED_DATA env var (default: false)
		if getEnvBool("AUTO_SEED_DATA", false) {
			if err := database.SeedDatabase(db); err != nil {
				log.Printf("Warning: Failed to seed database: %v", err)
			}
			database.SeedCountryPaymentPolicies(db)
			if err := database.SeedDemoWorkspace(db); err != nil {
				log.Printf("Warning: demo workspace seed: %v", err)
			}
		}
	}

	// Initialize repositories and services
	var svcs *servicesCommon.Services
	if db != nil {
		repos := repositoryCommon.NewRepositories(db)
		svcs = servicesCommon.NewServices(repos, cfg, db)
	} else {
		log.Println("Warning: Running in degraded mode - database unavailable")
	}

	// Create handlers
	h := handlers.New(cfg, svcs)
	if initErr := h.System.InitAgent(); initErr != nil {
		log.Printf("Warning: trade AI agent initialization failed: %v", initErr)
	} else if h.System.IsAgentReady() {
		// Propagate the full TradeAgent to admin and system AIService
		if h.AdminPortal != nil {
			h.AdminPortal.AttachTradeAgent(h.System.TradeAgent())
		}
		if h.System != nil {
			h.System.AttachAgentToAIService(h.System.TradeAgent())
		}
		log.Println("Trade AI agent initialized successfully (13 tools)")
	}
	// Initialize B2B DeepAgent coordinator (ProductExpert, PricingExpert, LogisticsExpert)
	if initErr := h.System.InitDeepAgent(); initErr != nil {
		log.Printf("Warning: B2B coordinator DeepAgent initialization failed: %v", initErr)
	} else if h.System.IsB2BCoordinatorReady() {
		log.Println("B2B Coordinator DeepAgent initialized successfully")
	}

	// Initialize Plan-Execute-Replan agent for structured order processing
	if initErr := h.System.InitOrderProcessingAgent(); initErr != nil {
		log.Printf("Warning: Order Processing P-E-R agent initialization failed: %v", initErr)
	} else if h.System.IsOrderProcessingReady() {
		log.Println("Order Processing P-E-R agent initialized successfully")
	}
	backgroundCtx, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()
	stopOrderDraftCleanup := startOrderDraftCleanup(backgroundCtx, svcs)
	stopOutboxRelay := startEventOutboxTradeRelay(backgroundCtx, svcs)

	// Setup router
	routerWithShutdown := api.SetupRouter(h, cfg, db)

	// Create server
	srv := &stdhttp.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      routerWithShutdown.Router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", srv.Addr)
		log.Printf("Environment: %s", cfg.Server.Environment)
		if cfg.Security.EnableSwagger {
			log.Printf("Swagger UI: http://localhost:%s/swagger/index.html", cfg.Server.Port)
		}
		if err := srv.ListenAndServe(); err != nil && err != stdhttp.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Stop rate limiter cleanup goroutine
	if routerWithShutdown.Limiter != nil {
		routerWithShutdown.Limiter.Stop()
	}
	if routerWithShutdown.AIPublicLimiter != nil {
		routerWithShutdown.AIPublicLimiter.Stop()
	}
	if routerWithShutdown.InquiryPublicLimiter != nil {
		routerWithShutdown.InquiryPublicLimiter.Stop()
	}
	if h.AuthScope != nil {
		h.AuthScope.StopLoginTracker()
	}
	stopOrderDraftCleanup()
	stopOutboxRelay()
	stopBackground()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

const defaultEventOutboxRelayIntervalSeconds = 30

// startEventOutboxTradeRelay 轮询发件箱，在订单确认为 confirmed 后异步创建 Trade（最终一致性）
func startEventOutboxTradeRelay(ctx context.Context, svcs *servicesCommon.Services) func() {
	if svcs == nil || svcs.AdminPortal == nil || svcs.AdminPortal.Order == nil || svcs.AdminPortal.Trade == nil {
		return func() {}
	}

	intervalSec := getEnvPositiveInt("EVENT_OUTBOX_RELAY_INTERVAL_SECONDS", defaultEventOutboxRelayIntervalSeconds)
	workerCtx, cancel := context.WithCancel(ctx)
	interval := time.Duration(intervalSec) * time.Second

	runRelay := func() {
		n, err := svcs.AdminPortal.Order.ProcessPendingTradeOutbox(workerCtx, svcs.AdminPortal.Trade, 50)
		if err != nil {
			log.Printf("Warning: trade outbox relay: %v", err)
			return
		}
		if n > 0 {
			log.Printf("Trade outbox relay processed %d event(s)", n)
		}
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		runRelay()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				runRelay()
			}
		}
	}()

	log.Printf("Trade outbox relay started (interval=%ds)", intervalSec)
	return cancel
}

func startOrderDraftCleanup(ctx context.Context, svcs *servicesCommon.Services) func() {
	if svcs == nil || svcs.System == nil || svcs.System.Order == nil {
		return func() {}
	}

	enabled := getEnvBool("ORDER_DRAFT_CLEANUP_ENABLED", true)
	if !enabled {
		log.Println("Order draft cleanup disabled by ORDER_DRAFT_CLEANUP_ENABLED=false")
		return func() {}
	}

	intervalMinutes := getEnvPositiveInt(
		"ORDER_DRAFT_CLEANUP_INTERVAL_MINUTES",
		defaultOrderDraftCleanupIntervalMinutes,
	)
	expireMinutes := getEnvPositiveInt(
		"ORDER_DRAFT_EXPIRE_MINUTES",
		defaultOrderDraftExpireMinutes,
	)
	batchSize := getEnvPositiveInt(
		"ORDER_DRAFT_CLEANUP_BATCH_SIZE",
		defaultOrderDraftCleanupBatchSize,
	)

	workerCtx, cancel := context.WithCancel(ctx)
	interval := time.Duration(intervalMinutes) * time.Minute
	expire := time.Duration(expireMinutes) * time.Minute

	runCleanup := func() {
		cutoff := time.Now().Add(-expire)
		released, err := svcs.System.Order.ReleaseExpiredPendingConfirmationOrders(workerCtx, cutoff, batchSize)
		if err != nil {
			log.Printf("Warning: order draft cleanup failed: %v", err)
			return
		}
		if released > 0 {
			log.Printf(
				"Order draft cleanup released %d expired draft(s) older than %d minutes",
				released,
				expireMinutes,
			)
		}
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		runCleanup()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				runCleanup()
			}
		}
	}()

	log.Printf(
		"Order draft cleanup worker started (interval=%dmin expire=%dmin batch=%d)",
		intervalMinutes,
		expireMinutes,
		batchSize,
	)
	return cancel
}

func getEnvBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Warning: invalid %s=%q, fallback to %t", key, value, defaultValue)
		return defaultValue
	}
	return parsed
}

func getEnvPositiveInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		log.Printf("Warning: invalid %s=%q, fallback to %d", key, value, defaultValue)
		return defaultValue
	}
	return parsed
}
