package database

import (
	"testing"

	modelsCommon "candypro/api/internal/models/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestEnsureDefaultSystemSettings(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&modelsCommon.SystemSetting{}); err != nil {
		t.Fatal(err)
	}
	configuredSiteName := modelsCommon.SystemSetting{Key: "site_name", Value: "Configured storefront", Category: "general"}
	if err := db.Create(&configuredSiteName).Error; err != nil {
		t.Fatal(err)
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
	if err := EnsureDefaultSystemSettings(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&modelsCommon.SystemSetting{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != int64(len(defaultSystemSettings)) {
		t.Fatalf("default seed must remain idempotent: got %d settings, want %d", count, len(defaultSystemSettings))
	}
	var persisted modelsCommon.SystemSetting
	if err := db.First(&persisted, "key = ?", configuredSiteName.Key).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.Value != configuredSiteName.Value {
		t.Fatalf("seed overwrote existing configuration: %q", persisted.Value)
	}
}
