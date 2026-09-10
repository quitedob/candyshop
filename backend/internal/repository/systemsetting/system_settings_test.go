package systemsetting

import (
	"context"
	"testing"

	modelsCommon "candypro/api/internal/models/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestFindByCategoryGeneral(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&modelsCommon.SystemSetting{}); err != nil {
		t.Fatal(err)
	}
	fixtures := []modelsCommon.SystemSetting{
		{Key: "site_name", Value: "Test storefront", Category: "general"},
		{Key: "smtp_host", Value: "localhost", Category: "email"},
	}
	if err := db.Create(&fixtures).Error; err != nil {
		t.Fatal(err)
	}
	r := NewSystemSettingRepository(db)
	settings, err := r.FindByCategory(context.Background(), "general")
	if err != nil {
		t.Fatalf("FindByCategory: %v", err)
	}
	if len(settings) != 1 || settings[0].Key != "site_name" {
		t.Fatalf("expected only general settings, got %+v", settings)
	}
}
