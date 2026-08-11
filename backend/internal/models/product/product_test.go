package product

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupProductModelTestDB migrates the products table in-memory for the model's
// index contract tests.
func setupProductModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Product{}); err != nil {
		t.Fatalf("migrate products: %v", err)
	}
	return db
}

// TestSoftDeletedSlugCanBeReused pins the G24a fix: the product slug unique
// index must not span soft-deleted rows, otherwise a deleted product's slug can
// never be reused by a new catalog entry (the full-table unique index makes the
// second INSERT fail with a constraint error).
func TestSoftDeletedSlugCanBeReused(t *testing.T) {
	db := setupProductModelTestDB(t)

	first := Product{ID: "p1", Slug: "gummy-bears", Name: "Gummy Bears"}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("create first product: %v", err)
	}
	if err := db.Delete(&first).Error; err != nil {
		t.Fatalf("soft delete first product: %v", err)
	}

	second := Product{ID: "p2", Slug: "gummy-bears", Name: "Gummy Bears 2"}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("reusing a soft-deleted slug must succeed, got: %v", err)
	}

	var live Product
	if err := db.Where("slug = ?", "gummy-bears").First(&live).Error; err != nil {
		t.Fatalf("load live product: %v", err)
	}
	if live.ID != "p2" {
		t.Fatalf("live product id = %q, want p2 (soft-deleted row must stay out of default scope)", live.ID)
	}
}
