package order

import (
	"context"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTaxTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsOrder.TaxRate{}); err != nil {
		t.Fatalf("migrate tax_rates: %v", err)
	}
	return db
}

func TestFindBestRate_NYRegionalWinsOverUSExport(t *testing.T) {
	db := setupTaxTestDB(t)
	now := time.Now()
	seeds := []modelsOrder.TaxRate{
		{ID: "us-export", Country: "US", Region: "", Rate: 0, Name: "US Export (B2B)", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: "us-ny", Country: "US", Region: "NY", Rate: 0.08875, Name: "NY Sales Tax", IsActive: true, CreatedAt: now, UpdatedAt: now},
	}
	for _, tr := range seeds {
		if err := db.Create(&tr).Error; err != nil {
			t.Fatal(err)
		}
	}

	repo := NewTaxRepository(db)
	rate, err := repo.FindBestRate(context.Background(), "US", "NY")
	if err != nil {
		t.Fatalf("FindBestRate: %v", err)
	}
	if rate.Region != "NY" {
		t.Fatalf("region = %q, want NY", rate.Region)
	}
	if rate.Rate != 0.08875 {
		t.Fatalf("rate = %v, want 0.08875", rate.Rate)
	}
}

func TestFindBestRate_FallsBackToCountryWide(t *testing.T) {
	db := setupTaxTestDB(t)
	now := time.Now()
	tr := modelsOrder.TaxRate{
		ID: "us-export", Country: "US", Region: "", Rate: 0, Name: "US Export (B2B)",
		IsActive: true, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&tr).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewTaxRepository(db)
	rate, err := repo.FindBestRate(context.Background(), "US", "CA")
	if err != nil {
		t.Fatalf("FindBestRate: %v", err)
	}
	if rate.Region != "" {
		t.Fatalf("region = %q, want empty fallback", rate.Region)
	}
}
