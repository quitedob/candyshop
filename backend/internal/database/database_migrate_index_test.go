package database

import (
	"testing"

	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupMigrateIndexTestDB simulates an already-migrated deployment: users and
// products are migrated with the current partial unique indexes, and the
// pre-G24a full-table unique indexes (idx_users_email, idx_products_slug) are
// recreated manually because GORM AutoMigrate never drops an index removed from
// a model (verified against GORM v1.30.0 naming: idx_<table>_<db-column>).
func setupMigrateIndexTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsUser.User{}, &modelsProduct.Product{}); err != nil {
		t.Fatalf("AutoMigrate users+products: %v", err)
	}
	// Recreate the legacy full-table unique indexes exactly as GORM named them
	// before G24a. These are the indexes that survive on every already-migrated
	// deployment because AutoMigrate does not drop removed indexes.
	if err := db.Exec("CREATE UNIQUE INDEX idx_users_email ON users(email)").Error; err != nil {
		t.Fatalf("create legacy users.email index: %v", err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX idx_products_slug ON products(slug)").Error; err != nil {
		t.Fatalf("create legacy products.slug index: %v", err)
	}
	return db
}

// TestAutoMigrate_DropsLegacyUniqueIndexes_UnblocksReuse drives the production
// startup migration (database.AutoMigrate, the exact function cmd/api/main.go
// calls) against a simulated already-migrated database that still carries the
// pre-G24a full-table unique indexes. Before the G24a wiring, AutoMigrate never
// dropped them, so a soft-deleted email/slug could never be re-registered /
// re-created (the exact unique-constraint 5xx the finding describes). After the
// wiring, AutoMigrate drops the legacy indexes and reuse succeeds on the partial
// unique indexes (WHERE deleted_at IS NULL).
func TestAutoMigrate_DropsLegacyUniqueIndexes_UnblocksReuse(t *testing.T) {
	db := setupMigrateIndexTestDB(t)

	firstUser := modelsUser.User{ID: "u1", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&firstUser).Error; err != nil {
		t.Fatalf("create first user: %v", err)
	}
	if err := db.Delete(&firstUser).Error; err != nil {
		t.Fatalf("soft delete first user: %v", err)
	}
	secondUser := modelsUser.User{ID: "u2", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&secondUser).Error; err == nil {
		t.Fatal("expected legacy full-table users.email index to block email reuse before the migration drop")
	}

	firstProduct := modelsProduct.Product{ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears"}
	if err := db.Create(&firstProduct).Error; err != nil {
		t.Fatalf("create first product: %v", err)
	}
	if err := db.Delete(&firstProduct).Error; err != nil {
		t.Fatalf("soft delete first product: %v", err)
	}
	secondProduct := modelsProduct.Product{ID: "p2", Slug: "gummy-bears", Name: "Gummy Bears 2"}
	if err := db.Create(&secondProduct).Error; err == nil {
		t.Fatal("expected legacy full-table products.slug index to block slug reuse before the migration drop")
	}

	// Production startup path: AutoMigrate must drop the legacy indexes.
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	if db.Migrator().HasIndex(&modelsUser.User{}, modelsUser.LegacyUniqueIndexEmail) {
		t.Fatal("AutoMigrate did not drop legacy users.email unique index (G24a wiring missing)")
	}
	if db.Migrator().HasIndex(&modelsProduct.Product{}, modelsProduct.LegacyUniqueIndexSlug) {
		t.Fatal("AutoMigrate did not drop legacy products.slug unique index (G24a wiring missing)")
	}

	if err := db.Create(&secondUser).Error; err != nil {
		t.Fatalf("email reuse after AutoMigrate must succeed: %v", err)
	}
	if err := db.Create(&secondProduct).Error; err != nil {
		t.Fatalf("slug reuse after AutoMigrate must succeed: %v", err)
	}

	var liveUser modelsUser.User
	if err := db.Where("email = ?", "reuse@example.com").First(&liveUser).Error; err != nil {
		t.Fatalf("load live user: %v", err)
	}
	if liveUser.ID != "u2" {
		t.Fatalf("live user id = %q, want u2 (soft-deleted row must stay out of default scope)", liveUser.ID)
	}
	var liveProduct modelsProduct.Product
	if err := db.Where("slug = ?", "gummy-bears").First(&liveProduct).Error; err != nil {
		t.Fatalf("load live product: %v", err)
	}
	if liveProduct.ID != "p2" {
		t.Fatalf("live product id = %q, want p2 (soft-deleted row must stay out of default scope)", liveProduct.ID)
	}
}

// TestAutoMigrate_FreshDatabaseDropIsNoop_ReuseWorks guards the other side of
// the G24a wiring: on a fresh database there are no legacy full-table unique
// indexes, so the new dropLegacyUniqueIndexes step must be a no-op and must not
// break the startup migration. After the startup migration, a soft-deleted
// email/slug must be re-usable via the partial unique indexes alone.
func TestAutoMigrate_FreshDatabaseDropIsNoop_ReuseWorks(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate on fresh DB: %v", err)
	}

	firstUser := modelsUser.User{ID: "u1", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&firstUser).Error; err != nil {
		t.Fatalf("create first user: %v", err)
	}
	if err := db.Delete(&firstUser).Error; err != nil {
		t.Fatalf("soft delete first user: %v", err)
	}
	secondUser := modelsUser.User{ID: "u2", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&secondUser).Error; err != nil {
		t.Fatalf("email reuse on fresh DB must succeed: %v", err)
	}

	firstProduct := modelsProduct.Product{ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears"}
	if err := db.Create(&firstProduct).Error; err != nil {
		t.Fatalf("create first product: %v", err)
	}
	if err := db.Delete(&firstProduct).Error; err != nil {
		t.Fatalf("soft delete first product: %v", err)
	}
	secondProduct := modelsProduct.Product{ID: "p2", Slug: "gummy-bears", Name: "Gummy Bears 2"}
	if err := db.Create(&secondProduct).Error; err != nil {
		t.Fatalf("slug reuse on fresh DB must succeed: %v", err)
	}
}
