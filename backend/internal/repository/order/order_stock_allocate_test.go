package order

import (
	"errors"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupStockTestDB 创建内存 SQLite 用于库存单元测试
func setupStockTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE products (id TEXT PRIMARY KEY, stock_quantity INTEGER NOT NULL DEFAULT 0, deleted_at DATETIME, created_at DATETIME, updated_at DATETIME)`).Error; err != nil {
		t.Fatalf("create products: %v", err)
	}
	if err := db.Exec(`CREATE TABLE warehouse_stocks (id TEXT PRIMARY KEY, warehouse_id TEXT, product_id TEXT, quantity INTEGER, reserved INTEGER)`).Error; err != nil {
		t.Fatalf("create warehouse_stocks: %v", err)
	}
	if err := db.Exec(`CREATE TABLE oem_project_inventory_holds (
		id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, status TEXT
	)`).Error; err != nil {
		t.Fatalf("create holds: %v", err)
	}
	if err := db.Exec(`CREATE TABLE product_batches (
		id TEXT PRIMARY KEY, product_id TEXT, batch_number TEXT, quantity INTEGER,
		expiry_date DATETIME, is_expired INTEGER DEFAULT 0, warehouse_id TEXT,
		created_at DATETIME, updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create batches: %v", err)
	}
	return db
}

// TestDeductLegacyProductStock_ErrInsufficientStock OEM 预留后库存不足应返回 ErrInsufficientStock
func TestDeductLegacyProductStock_ErrInsufficientStock(t *testing.T) {
	db := setupStockTestDB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO oem_project_inventory_holds (id, product_id, quantity, status) VALUES ('h1', 'p1', 8, 'active')`).Error; err != nil {
		t.Fatal(err)
	}

	_, err := deductLegacyProductStock(db, "p1", 5, modelsOrder.StockReasonStockReserved, "ord-1", "user-1", time.Now())
	if !errors.Is(err, modelsOrder.ErrInsufficientStock) {
		t.Fatalf("expected ErrInsufficientStock, got %v", err)
	}
}

// TestReserveStockForOrderLine_FEFORoute 无仓记录但有批次时应走 FEFO 路径
func TestReserveStockForOrderLine_FEFORoute(t *testing.T) {
	db := setupStockTestDB(t)
	now := time.Now().UTC().Add(24 * time.Hour)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p2', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO product_batches (id, product_id, batch_number, quantity, expiry_date, is_expired) VALUES ('b1', 'p2', 'LOT-1', 20, ?, 0)`, now).Error; err != nil {
		t.Fatal(err)
	}

	recs, err := reserveStockForOrderLine(db, "", "p2", 5, modelsOrder.StockReasonStockReserved, "ord-2", "user-1", time.Now())
	if err != nil {
		t.Fatalf("reserveStockForOrderLine: %v", err)
	}
	if len(recs) == 0 {
		t.Fatal("expected stock transaction records from FEFO path")
	}
	var stockQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id = 'p2'`).Scan(&stockQty).Error; err != nil {
		t.Fatal(err)
	}
	if stockQty != 15 {
		t.Fatalf("expected stock 15 after reserve, got %d", stockQty)
	}
}
