package systemsetting

import (
	"context"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestFindByCategoryGeneral(t *testing.T) {
	dsn := "host=localhost user=postgres password=1234 dbname=candypro port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skip("database unavailable:", err)
	}
	r := NewSystemSettingRepository(db)
	settings, err := r.FindByCategory(context.Background(), "general")
	if err != nil {
		t.Fatalf("FindByCategory: %v", err)
	}
	t.Logf("settings count: %d", len(settings))
}
