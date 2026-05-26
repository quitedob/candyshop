package order

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupOrderReserveTestDB 创建订单库存预留集成测试库
func setupOrderReserveTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE products (id TEXT PRIMARY KEY, stock_quantity INTEGER NOT NULL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE warehouses (id TEXT PRIMARY KEY, code TEXT, is_active INTEGER DEFAULT 1, is_default INTEGER DEFAULT 0)`,
		`CREATE TABLE warehouse_stocks (id INTEGER PRIMARY KEY AUTOINCREMENT, warehouse_id TEXT, product_id TEXT, variant_id TEXT, quantity INTEGER, reserved INTEGER, updated_at DATETIME)`,
		`CREATE TABLE oem_project_inventory_holds (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, status TEXT)`,
		`CREATE TABLE product_batches (id TEXT PRIMARY KEY, product_id TEXT, batch_number TEXT, quantity INTEGER, expiry_date DATETIME, is_expired INTEGER DEFAULT 0, warehouse_id TEXT)`,
		`CREATE TABLE orders (id TEXT PRIMARY KEY, user_id TEXT, status TEXT, stock_reserved INTEGER DEFAULT 0, items BLOB, updated_at DATETIME, version INTEGER DEFAULT 0, deleted_at DATETIME)`,
		`CREATE TABLE stock_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT, change INTEGER, stock_before INTEGER, stock_after INTEGER, reason TEXT, reference_id TEXT, operator_id TEXT, batch_id TEXT, lot_number TEXT, warehouse_id TEXT, created_at DATETIME)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// TestReserveWarehouseStock_AutoProvision 仓级行缺失时应从产品汇总库存自动创建
func TestReserveWarehouseStock_AutoProvision(t *testing.T) {
	db := setupOrderReserveTestDB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 100)`).Error; err != nil {
		t.Fatal(err)
	}
	_, err := reserveWarehouseStock(db, "wh-main", "p1", 10, modelsOrder.StockReasonStockReserved, "ord-1", "user-1", time.Now())
	if err != nil {
		t.Fatalf("reserveWarehouseStock: %v", err)
	}
	var qty, reserved int
	if err := db.Raw(`SELECT quantity, reserved FROM warehouse_stocks WHERE warehouse_id = 'wh-main' AND product_id = 'p1'`).Row().Scan(&qty, &reserved); err != nil {
		t.Fatal(err)
	}
	if qty != 100 || reserved != 10 {
		t.Fatalf("warehouse stock qty=%d reserved=%d, want 100/10", qty, reserved)
	}
}

// TestReserveStockForOrder_LegacyNoWarehouse 无仓记录时 legacy 路径应成功预留
func TestReserveStockForOrder_LegacyNoWarehouse(t *testing.T) {
	db := setupOrderReserveTestDB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 50)`).Error; err != nil {
		t.Fatal(err)
	}
	itemsJSON, _ := json.Marshal([]modelsOrder.OrderItem{{ProductID: "p1", Quantity: 5}})
	if err := db.Exec(`INSERT INTO orders (id, user_id, status, stock_reserved, items, updated_at) VALUES ('ord-1', 'user-1', 'pending', 0, ?, datetime('now'))`, itemsJSON).Error; err != nil {
		t.Fatal(err)
	}
	order := &modelsOrder.Order{ID: "ord-1", UserID: "user-1", Items: modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}}}
	repo := NewOrderRepository(db)
	if err := repo.ReserveStockForOrder(context.Background(), order, map[string]int{"p1": 5}); err != nil {
		t.Fatalf("ReserveStockForOrder legacy: %v", err)
	}
	var stockQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id = 'p1'`).Scan(&stockQty).Error; err != nil {
		t.Fatal(err)
	}
	if stockQty != 45 {
		t.Fatalf("expected stock 45, got %d", stockQty)
	}
	var reservedFlag int
	if err := db.Raw(`SELECT stock_reserved FROM orders WHERE id = 'ord-1'`).Scan(&reservedFlag).Error; err != nil {
		t.Fatal(err)
	}
	if reservedFlag != 1 {
		t.Fatalf("expected stock_reserved=1, got %d", reservedFlag)
	}
}

// TestReserveStockForOrder_WithMainWarehouse 有 MAIN 仓及仓级库存时应走仓级预留
func TestReserveStockForOrder_WithMainWarehouse(t *testing.T) {
	db := setupOrderReserveTestDB(t)
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('wh-main', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 50)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('wh-main', 'p1', 50, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	itemsJSON, _ := json.Marshal([]modelsOrder.OrderItem{{ProductID: "p1", Quantity: 5}})
	if err := db.Exec(`INSERT INTO orders (id, user_id, status, stock_reserved, items, updated_at) VALUES ('ord-2', 'user-1', 'pending', 0, ?, datetime('now'))`, itemsJSON).Error; err != nil {
		t.Fatal(err)
	}
	order := &modelsOrder.Order{ID: "ord-2", UserID: "user-1", Items: modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}}}
	repo := NewOrderRepository(db)
	if err := repo.ReserveStockForOrder(context.Background(), order, map[string]int{"p1": 5}); err != nil {
		t.Fatalf("ReserveStockForOrder warehouse: %v", err)
	}
	var whReserved int
	if err := db.Raw(`SELECT reserved FROM warehouse_stocks WHERE warehouse_id = 'wh-main' AND product_id = 'p1'`).Scan(&whReserved).Error; err != nil {
		t.Fatal(err)
	}
	if whReserved != 5 {
		t.Fatalf("expected warehouse reserved=5, got %d", whReserved)
	}
}

// TestReserveStockForOrder_PreservesInMemoryStatus BUG-02: locking must not overwrite handler mutations.
func TestReserveStockForOrder_PreservesInMemoryStatus(t *testing.T) {
	db := setupOrderReserveTestDB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 50)`).Error; err != nil {
		t.Fatal(err)
	}
	itemsJSON, _ := json.Marshal([]modelsOrder.OrderItem{{ProductID: "p1", Quantity: 5}})
	if err := db.Exec(`INSERT INTO orders (id, user_id, status, stock_reserved, items, updated_at) VALUES ('ord-3', 'user-1', 'pending', 0, ?, datetime('now'))`, itemsJSON).Error; err != nil {
		t.Fatal(err)
	}

	order := &modelsOrder.Order{
		ID:            "ord-3",
		UserID:        "user-1",
		Status:        "confirmed",
		PaymentStatus: "paid",
		Items:         modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}},
	}
	repo := NewOrderRepository(db)
	if err := repo.ReserveStockForOrder(context.Background(), order, map[string]int{"p1": 5}); err != nil {
		t.Fatalf("ReserveStockForOrder: %v", err)
	}
	if order.Status != "confirmed" {
		t.Fatalf("expected in-memory status=confirmed, got %q", order.Status)
	}
	if order.PaymentStatus != "paid" {
		t.Fatalf("expected in-memory payment_status=paid, got %q", order.PaymentStatus)
	}
	if !order.StockReserved {
		t.Fatal("expected StockReserved=true in memory")
	}
}
