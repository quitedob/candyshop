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
	"candypro/api/internal/pkg/applog"
	"candypro/api/internal/pkg/eino"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/safego"
	"candypro/api/internal/pkg/workerlock"

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
		log.Fatalf("配置加载失败: %v", err)
	}
	applog.Init(cfg.Server.Environment)
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
		if err := database.NormalizeLegacyLeadTimeCopy(db); err != nil {
			log.Printf("Warning: legacy lead-time copy normalization failed: %v", err)
		}
		if err := database.EnsureDefaultWarehouseStock(db); err != nil {
			log.Printf("Warning: default warehouse stock ensure failed: %v", err)
		}
		if err := database.EnsureStartupData(db); err != nil {
			log.Printf("Warning: startup data ensure failed: %v", err)
		}
		if err := database.SeedCategoryTranslations(db); err != nil {
			log.Printf("Warning: category translation seed failed: %v", err)
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
			if err := database.SeedProductTranslations(db); err != nil {
				log.Printf("Warning: Failed to seed product translations: %v", err)
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
		log.Println("Trade AI agent initialized successfully")
		log.Printf("Trade AI agent tool count: %d", eino.LastTradeAgentToolCount)
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
	// Wire PostgreSQL checkpoint store for agent persistence across restarts
	if db != nil {
		h.System.InitCheckPointStore(db)
		if migrateErr := eino.AutoMigrateErr(); migrateErr != nil {
			log.Printf("Warning: checkpoint table migration failed: %v", migrateErr)
		} else if h.System.CheckPointStore() != nil {
			log.Println("Agent checkpoint store initialized (PostgreSQL)")
			if h.AdminPortal != nil {
				h.AdminPortal.AttachCheckPointStore(h.System.CheckPointStore())
			}
		}
	}

	backgroundCtx, stopBackground := context.WithCancel(context.Background())
	defer stopBackground()
	// L-6: distributed worker mutex. Multi-instance deployments must avoid
	// running the same scheduled job in parallel; jobs wrap their tick body in
	// workerlock.WithLock(...) so only one instance executes per interval.
	workerLocker := workerlock.NewFromEnv(cfg.Security.RedisURL)
	stopOrderDraftCleanup := startOrderDraftCleanup(backgroundCtx, svcs, workerLocker)
	stopOutboxRelay := startEventOutboxTradeRelay(backgroundCtx, svcs, workerLocker)
	stopCheckpointCleanup := startCheckpointCleanup(backgroundCtx, h.System.CheckPointStore(), workerLocker)
	stopShipmentTrackingSync := startShipmentTrackingSync(backgroundCtx, svcs, workerLocker)
	stopNotificationRelay := startNotificationRelay(backgroundCtx, svcs)

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
	stopCheckpointCleanup()
	stopShipmentTrackingSync()
	stopNotificationRelay()
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
func startEventOutboxTradeRelay(ctx context.Context, svcs *servicesCommon.Services, locker workerlock.Locker) func() {
	if svcs == nil || svcs.AdminPortal == nil || svcs.AdminPortal.Order == nil || svcs.AdminPortal.Trade == nil {
		return func() {}
	}

	intervalSec := getEnvPositiveInt("EVENT_OUTBOX_RELAY_INTERVAL_SECONDS", defaultEventOutboxRelayIntervalSeconds)
	workerCtx, cancel := context.WithCancel(ctx)
	interval := time.Duration(intervalSec) * time.Second

	runRelay := func() {
		// L-6: only one instance runs the relay per tick.
		_, _ = workerlock.WithLock(workerCtx, locker, "trade-outbox-relay", interval, func(c context.Context) error {
			n, err := svcs.AdminPortal.Order.ProcessPendingTradeOutbox(c, svcs.AdminPortal.Trade, 50)
			if err != nil {
				log.Printf("Warning: trade outbox relay: %v", err)
				return err
			}
			if n > 0 {
				log.Printf("Trade outbox relay processed %d event(s)", n)
			}
			return nil
		})
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		safego.Run("trade-outbox-relay.initial", runRelay)
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				safego.Run("trade-outbox-relay.tick", runRelay)
			}
		}
	}()

	log.Printf("Trade outbox relay started (interval=%ds)", intervalSec)
	return cancel
}

func startOrderDraftCleanup(ctx context.Context, svcs *servicesCommon.Services, locker workerlock.Locker) func() {
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
		_, _ = workerlock.WithLock(workerCtx, locker, "order-draft-cleanup", interval, func(c context.Context) error {
			cutoff := time.Now().Add(-expire)
			released, err := svcs.System.Order.ReleaseExpiredPendingConfirmationOrders(c, cutoff, batchSize)
			if err != nil {
				log.Printf("Warning: order draft cleanup failed: %v", err)
				return err
			}
			if released > 0 {
				log.Printf(
					"Order draft cleanup released %d expired draft(s) older than %d minutes",
					released,
					expireMinutes,
				)
			}
			return nil
		})
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		safego.Run("order-draft-cleanup.initial", runCleanup)
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				safego.Run("order-draft-cleanup.tick", runCleanup)
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

const (
	defaultCheckpointCleanupIntervalMinutes = 60
	defaultCheckpointTTLHours               = 24
)

// startCheckpointCleanup periodically deletes stale agent checkpoints.
func startCheckpointCleanup(ctx context.Context, store *eino.PostgresCheckPointStore, locker workerlock.Locker) func() {
	if store == nil {
		return func() {}
	}

	intervalMin := getEnvPositiveInt("CHECKPOINT_CLEANUP_INTERVAL_MINUTES", defaultCheckpointCleanupIntervalMinutes)
	ttlHours := getEnvPositiveInt("CHECKPOINT_TTL_HOURS", defaultCheckpointTTLHours)

	workerCtx, cancel := context.WithCancel(ctx)
	interval := time.Duration(intervalMin) * time.Minute
	ttl := time.Duration(ttlHours) * time.Hour

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				safego.Run("checkpoint-cleanup.tick", func() {
					_, _ = workerlock.WithLock(workerCtx, locker, "checkpoint-cleanup", interval, func(c context.Context) error {
						deleted, err := store.CleanupOlderThan(c, ttl)
						if err != nil {
							log.Printf("Warning: checkpoint cleanup failed: %v", err)
							return err
						}
						if deleted > 0 {
							log.Printf("Checkpoint cleanup deleted %d stale entry/entries (TTL=%dh)", deleted, ttlHours)
						}
						return nil
					})
				})
			}
		}
	}()

	log.Printf("Checkpoint cleanup worker started (interval=%dmin TTL=%dh)", intervalMin, ttlHours)
	return cancel
}

const (
	defaultShipmentTrackingSyncIntervalMinutes = 15
	defaultShipmentTrackingSyncBatchSize       = 50
)

// startShipmentTrackingSync 定期根据已有追踪事件同步发货状态。
func startShipmentTrackingSync(ctx context.Context, svcs *servicesCommon.Services, locker workerlock.Locker) func() {
	if svcs == nil || svcs.UserPortal == nil || svcs.UserPortal.Logistics == nil {
		return func() {}
	}
	intervalMin := getEnvPositiveInt("SHIPMENT_TRACKING_SYNC_INTERVAL_MINUTES", defaultShipmentTrackingSyncIntervalMinutes)
	batchSize := getEnvPositiveInt("SHIPMENT_TRACKING_SYNC_BATCH_SIZE", defaultShipmentTrackingSyncBatchSize)
	workerCtx, cancel := context.WithCancel(ctx)

	go func() {
		ticker := time.NewTicker(time.Duration(intervalMin) * time.Minute)
		defer ticker.Stop()
		runSync := func() {
			_, _ = workerlock.WithLock(workerCtx, locker, "shipment-tracking-sync", time.Duration(intervalMin)*time.Minute, func(c context.Context) error {
				n, err := svcs.UserPortal.Logistics.SyncActiveShipmentTracking(c, batchSize)
				if err != nil {
					log.Printf("Shipment tracking sync error: %v", err)
					return err
				}
				if n > 0 {
					log.Printf("Shipment tracking sync updated %d shipment(s)", n)
				}
				return nil
			})
		}
		safego.Run("shipment-tracking-sync.initial", runSync)
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				safego.Run("shipment-tracking-sync.tick", runSync)
			}
		}
	}()

	log.Printf("Shipment tracking sync worker started (interval=%dmin batch=%d)", intervalMin, batchSize)
	return cancel
}


const (
	defaultNotificationRelayIntervalSeconds = 15
	defaultNotificationRelayBatchSize       = 50
)

// startNotificationRelay drains the notification_outbox table on a tick (C-8).
// Producers insert rows in the same DB tx as their business mutation; this
// worker performs the network call asynchronously with retry/backoff.
func startNotificationRelay(ctx context.Context, svcs *servicesCommon.Services) func() {
	if svcs == nil || svcs.NotificationRelay == nil {
		return func() {}
	}
	intervalSec := getEnvPositiveInt("NOTIFICATION_RELAY_INTERVAL_SECONDS", defaultNotificationRelayIntervalSeconds)
	batchSize := getEnvPositiveInt("NOTIFICATION_RELAY_BATCH_SIZE", defaultNotificationRelayBatchSize)
	workerCtx, cancel := context.WithCancel(ctx)

	run := func() {
		processed, err := svcs.NotificationRelay.ProcessBatch(workerCtx, batchSize)
		if err != nil {
			log.Printf("notification relay: %v", err)
			return
		}
		if processed > 0 {
			log.Printf("Notification relay processed %d notification(s)", processed)
		}
	}

	go func() {
		ticker := time.NewTicker(time.Duration(intervalSec) * time.Second)
		defer ticker.Stop()
		safego.Run("notification-relay.initial", run)
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				safego.Run("notification-relay.tick", run)
			}
		}
	}()
	log.Printf("Notification relay worker started (interval=%ds batch=%d)", intervalSec, batchSize)
	return cancel
}
