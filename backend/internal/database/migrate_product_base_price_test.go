package database

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupBasePriceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// Minimal products table; deleted_at is required because the Product model
	// declares gorm.DeletedAt (GORM appends `deleted_at IS NULL` to updates), and
	// updated_at because GORM stamps it on every Update.
	if err := db.Exec(`CREATE TABLE products (
		id TEXT PRIMARY KEY,
		slug TEXT,
		base_price REAL DEFAULT 0,
		updated_at DATETIME,
		deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("schema: %v", err)
	}
	return db
}

// TestEnsureProductBasePrices_KeepsExistingPrices proves admin-edited base_price
// values are not silently reverted to the catalog default on startup.
func TestEnsureProductBasePrices_KeepsExistingPrices(t *testing.T) {
	db := setupBasePriceTestDB(t)

	if err := db.Exec(`INSERT INTO products (id, slug, base_price) VALUES
		('p-edited', '4d-fruit-gummy', 9.99),   -- catalog default 8.50, admin set to 9.99
		('p-unset',  'crystal-hard-candy', 0),  -- catalog default 5.10, never set
		('p-custom', 'custom-candy', 0)`).Error; err != nil { // not a catalog slug
		t.Fatal(err)
	}

	if err := EnsureProductBasePrices(db); err != nil {
		t.Fatalf("EnsureProductBasePrices: %v", err)
	}

	var editedPrice float64
	if err := db.Raw(`SELECT base_price FROM products WHERE id = 'p-edited'`).Scan(&editedPrice).Error; err != nil {
		t.Fatal(err)
	}
	if editedPrice != 9.99 {
		t.Fatalf("admin-edited base_price overwritten: got %v, want 9.99 (catalog default 8.50 must not revert)", editedPrice)
	}

	var unsetPrice float64
	if err := db.Raw(`SELECT base_price FROM products WHERE id = 'p-unset'`).Scan(&unsetPrice).Error; err != nil {
		t.Fatal(err)
	}
	if unsetPrice != 5.10 {
		t.Fatalf("unset base_price not backfilled: got %v, want 5.10", unsetPrice)
	}

	var customPrice float64
	if err := db.Raw(`SELECT base_price FROM products WHERE id = 'p-custom'`).Scan(&customPrice).Error; err != nil {
		t.Fatal(err)
	}
	if customPrice != 0 {
		t.Fatalf("non-catalog slug unexpectedly priced: got %v, want 0", customPrice)
	}
}
