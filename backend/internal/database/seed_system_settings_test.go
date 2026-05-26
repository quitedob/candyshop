package database

import (
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestEnsureDefaultSystemSettings(t *testing.T) {
	dsn := "host=localhost user=postgres password=1234 dbname=candypro port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("database unavailable:", err)
	}
	if err := EnsureDefaultSystemSettings(db); err != nil {
		t.Fatalf("EnsureDefaultSystemSettings: %v", err)
	}
	var count int64
	if err := db.Table("system_settings").Where("category = ?", "general").Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count == 0 {
		t.Fatalf("expected general settings to be seeded")
	}
}
