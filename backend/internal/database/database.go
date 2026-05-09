package database

import (
	modelsAuth "candypro/api/internal/models/auth"
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsTrade "candypro/api/internal/models/trade"
	modelsUser "candypro/api/internal/models/user"
)

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"candypro/api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// db is the global database connection (unexported)
var db *gorm.DB

var pgvectorAvailable bool

var safeDBNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Connect initializes the database connection
func Connect(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	// First, connect to postgres default database to create target database if needed
	if err := ensureDatabase(cfg); err != nil {
		log.Printf("Warning: could not ensure database exists: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port, cfg.SSLMode,
	)

	// Configure GORM logger based on environment
	gormLogger := logger.Default.LogMode(logger.Warn)
	if cfg.Environment == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := dbConn.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool (values from config, defaulting to sensible defaults)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	available, err := ensurePgvector(dbConn)
	if err != nil {
		log.Printf("Note: pgvector capability check failed (semantic search may be keyword-only): %v", err)
	}
	pgvectorAvailable = available
	if pgvectorAvailable {
		log.Println("pgvector extension available; semantic search enabled")
	} else {
		log.Println("pgvector extension unavailable; semantic search will fall back to keyword-only mode")
	}

	log.Println("Database connection established")
	db = dbConn
	return dbConn, nil
}

func ensurePgvector(db *gorm.DB) (bool, error) {
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		log.Printf("Note: CREATE EXTENSION vector failed: %v", err)
	}

	var installed bool
	if err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'vector')").Scan(&installed).Error; err != nil {
		return false, err
	}
	return installed, nil
}

func PgvectorAvailable() bool {
	return pgvectorAvailable
}

// ensureDatabase connects to the postgres default database and creates the target
// database if it does not already exist.
func ensureDatabase(cfg *config.DatabaseConfig) error {
	// Connect to postgres database (default admin database)
	adminDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		cfg.Host, cfg.User, cfg.Password, cfg.Port, cfg.SSLMode,
	)
	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres admin database: %w", err)
	}

	// Check if target database exists
	var count int64
	adminDB.Raw("SELECT 1 FROM pg_database WHERE datname = ?", cfg.Database).Count(&count)
	if count > 0 {
		log.Printf("Database %s already exists", cfg.Database)
		return nil
	}

	// Validate database name against safe pattern before DDL
	if !safeDBNamePattern.MatchString(cfg.Database) {
		return fmt.Errorf("database name %q contains unsafe characters", cfg.Database)
	}

	log.Printf("Creating database %s...", cfg.Database)
	if err := adminDB.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, cfg.Database)).Error; err != nil {
		// If CREATE fails because another process just created it, that's okay
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create database %s: %w", cfg.Database, err)
		}
	}
	log.Printf("Database %s created successfully", cfg.Database)
	return nil
}

// AutoMigrate runs auto migration for all models in dependency order
func AutoMigrate(db *gorm.DB) error {
	// Phase 1: Migrate tables without foreign key dependencies first
	phase1 := []any{
		&modelsAuth.Role{},
		&modelsUser.Company{},
		&modelsAuth.RefreshToken{},
		&modelsProduct.Category{},
		&modelsProduct.OEMFlow{},
		&modelsProduct.OEMSolution{},
		&modelsProduct.FactoryInfo{},
		&modelsProduct.Certification{},
		&modelsProduct.PriceList{},
		&modelsProduct.PriceRule{},
		&modelsProduct.ProcessControl{},
		&modelsCommon.ActivityLog{},
		&modelsCommon.SystemSetting{},
		&modelsTrade.ComplianceRequirement{},
		&modelsOrder.StockTransaction{},
		&modelsOrder.EventOutbox{},
		&modelsOrder.CountryPaymentPolicy{},
		&modelsOrder.DocumentAdjustment{},
		&modelsTrade.ShipmentEvent{},
	}
	for _, model := range phase1 {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to auto migrate (phase 1): %w", err)
		}
	}

	// Phase 2: Migrate tables with foreign key dependencies
	phase2 := []any{
		&modelsUser.User{},
		&modelsProduct.Product{},
		&modelsProduct.ProductVariant{},
		&modelsProduct.ProductMarketProfile{},
		&modelsProduct.ProductMarketCostStack{},
		&modelsProduct.Warehouse{},
		&modelsProduct.WarehouseStock{},
		&modelsProduct.ProductBatch{},
		&modelsProduct.Inquiry{},
		&modelsProduct.OEMProject{},
		&modelsProduct.OEMProjectInventoryHold{},
		&modelsProduct.ChannelInventory{},
		&modelsProduct.BlogPost{},
		&modelsProduct.CaseStudy{},
		&modelsOrder.Order{},
		&modelsOrder.Payment{},
		&modelsOrder.CartItem{},
		&modelsOrder.Invoice{},
		&modelsCommon.Notification{},
		&modelsTrade.TradeTransaction{},
		&modelsTrade.TradeDocument{},
		&modelsTrade.SalesContract{},
		&modelsTrade.PackingList{},
		&modelsTrade.CertificateOfOrigin{},
		&modelsTrade.HealthCertificate{},
		&modelsTrade.ShipmentTracking{},
		&modelsTrade.SettlementRecord{},
		&modelsTrade.ProformaInvoice{},
		&modelsTrade.CommercialInvoice{},
		&modelsTrade.BillOfLading{},
	}
	for _, model := range phase2 {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to auto migrate (phase 2): %w", err)
		}
	}

	if pgvectorAvailable {
		if err := db.AutoMigrate(&modelsProduct.ProductEmbedding{}); err != nil {
			return fmt.Errorf("failed to auto migrate product embeddings: %w", err)
		}
	}

	log.Println("Database migration completed")
	return nil
}

// GetDB returns the global database connection
func GetDB() *gorm.DB {
	return db
}

// SeedDatabase seeds the database with initial data
func SeedDatabase(db *gorm.DB) error {
	// Check if we already have data
	var count int64
	db.Model(&modelsProduct.Category{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	log.Println("Seeding database with initial data...")

	// Run seed functions
	if err := seedSuperadmin(db); err != nil {
		return err
	}
	if err := seedCategories(db); err != nil {
		return err
	}
	if err := seedProducts(db); err != nil {
		return err
	}
	if err := seedOEMFlows(db); err != nil {
		return err
	}
	if err := seedOEMSolutions(db); err != nil {
		return err
	}
	if err := seedCertifications(db); err != nil {
		return err
	}
	if err := seedFactoryInfo(db); err != nil {
		return err
	}
	if err := seedBlogPosts(db); err != nil {
		return err
	}
	if err := seedCaseStudies(db); err != nil {
		return err
	}

	log.Println("Database seeding completed")
	return nil
}

// SeedCountryPaymentPolicies ensures default country payment policies exist.
// Called separately from the main seed to support existing databases.
func SeedCountryPaymentPolicies(db *gorm.DB) {
	var count int64
	db.Model(&modelsOrder.CountryPaymentPolicy{}).Count(&count)
	if count > 0 {
		return
	}
	policies := []modelsOrder.CountryPaymentPolicy{
		{
			Country:                "india",
			RequiresFullPrepayment: true,
			AllowedTerms:           "100% T/T before production",
			Note:                   "India orders follow stricter risk-control policy: full prepayment is required before production scheduling.",
		},
		{
			Country:                "pakistan",
			RequiresFullPrepayment: true,
			AllowedTerms:           "100% T/T before production",
			Note:                   "Pakistan orders follow stricter risk-control policy: full prepayment is required before production scheduling.",
		},
	}
	for _, p := range policies {
		if err := db.Where("country = ?", p.Country).FirstOrCreate(&p).Error; err != nil {
			log.Printf("Warning: failed to seed payment policy for %s: %v", p.Country, err)
		}
	}
}
