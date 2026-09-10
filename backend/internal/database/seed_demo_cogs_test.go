package database

import (
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestUpsertDemoOrder_PopulatesConfirmedCostsOnFirstSeed(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(&modelsOrder.Order{}); err != nil {
		t.Fatal(err)
	}
	if err := database.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY, slug TEXT, base_price REAL, weighted_avg_cost REAL)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Exec(`INSERT INTO products VALUES ('demo-product', 'demo-product', 10, 4)`).Error; err != nil {
		t.Fatal(err)
	}
	confirmedAt := time.Now()
	order := modelsOrder.Order{
		ID: "demo-confirmed", OrderNumber: "DEMO-CONFIRMED", UserID: "demo-buyer",
		Status: modelsOrder.OrderStatusConfirmed, ConfirmedAt: &confirmedAt,
		Items:    modelsOrder.OrderItemArray{{ProductID: "demo-product", Quantity: 2, UnitPrice: 10}},
		Subtotal: 20, TotalAmount: 20,
	}
	if err := upsertDemoOrder(database, &order); err != nil {
		t.Fatal(err)
	}
	var persisted modelsOrder.Order
	if err := database.First(&persisted, "id = ?", order.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.COGS != 8 {
		t.Fatalf("first seed COGS = %v, want 2 × 4", persisted.COGS)
	}
	// Re-seeding can repair missing costs but must preserve a prior positive
	// historical cost even when the current product estimate has changed.
	if err := database.Model(&modelsOrder.Order{}).Where("id = ?", order.ID).Update("cogs", 6).Error; err != nil {
		t.Fatal(err)
	}
	if err := upsertDemoOrder(database, &order); err != nil {
		t.Fatal(err)
	}
	if err := database.First(&persisted, "id = ?", order.ID).Error; err != nil {
		t.Fatal(err)
	}
	if persisted.COGS != 6 || persisted.Version != 1 {
		t.Fatalf("reseed must preserve historical COGS and invalidate stale snapshots: %+v", persisted)
	}
}
