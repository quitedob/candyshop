package product

import (
	"context"
	"strings"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupSupplierTestDB creates an in-memory SQLite with the minimal tables used
// by ReceivePO (purchase_orders, purchase_order_items, products,
// warehouse_stocks, product_batches, stock_transactions).
func setupSupplierTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE products (
			id TEXT PRIMARY KEY, stock_quantity INTEGER DEFAULT 0, weighted_avg_cost REAL DEFAULT 0,
			created_at DATETIME, updated_at DATETIME, deleted_at DATETIME
		)`,
		`CREATE TABLE purchase_orders (
			id TEXT PRIMARY KEY, po_number TEXT NOT NULL, supplier_id TEXT, warehouse_id TEXT,
			status TEXT DEFAULT 'draft', expected_date DATETIME, received_date DATETIME, notes TEXT,
			created_by TEXT, created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE purchase_order_items (
			id INTEGER PRIMARY KEY AUTOINCREMENT, po_id TEXT NOT NULL, product_id TEXT NOT NULL,
			quantity INTEGER NOT NULL, unit_cost REAL DEFAULT 0, received_qty INTEGER DEFAULT 0
		)`,
		`CREATE TABLE warehouse_stocks (
			id INTEGER PRIMARY KEY AUTOINCREMENT, warehouse_id TEXT NOT NULL, product_id TEXT NOT NULL,
			variant_id TEXT, quantity INTEGER DEFAULT 0, reserved INTEGER DEFAULT 0, updated_at DATETIME
		)`,
		`CREATE TABLE product_batches (
			id TEXT PRIMARY KEY, product_id TEXT NOT NULL, variant_id TEXT, warehouse_id TEXT NOT NULL,
			batch_number TEXT NOT NULL, quantity INTEGER DEFAULT 0, unit_cost REAL DEFAULT 0,
			production_date DATETIME, expiry_date DATETIME, is_expired BOOLEAN DEFAULT 0, notes TEXT,
			created_at DATETIME, updated_at DATETIME
		)`,
		`CREATE TABLE stock_transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT NOT NULL, change INTEGER NOT NULL,
			stock_before INTEGER NOT NULL, stock_after INTEGER NOT NULL, reason TEXT NOT NULL,
			reference_id TEXT, operator_id TEXT, batch_id TEXT, lot_number TEXT, warehouse_id TEXT,
			created_at DATETIME
		)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// seedReceivePO seeds a PO for p1 (quantity qty, unit cost cost) plus the
// matching product and warehouse stock rows so a valid receive can apply.
func seedReceivePO(t *testing.T, db *gorm.DB, poID string, qty int, unitCost float64, status string) {
	t.Helper()
	if err := db.Exec(`INSERT INTO purchase_orders (id, po_number, supplier_id, warehouse_id, status, created_at, updated_at)
		VALUES (?, ?, 'sup-1', 'wh-1', ?, datetime('now'), datetime('now'))`,
		poID, "PO-"+poID, status).Error; err != nil {
		t.Fatalf("seed po: %v", err)
	}
	if err := db.Exec(`INSERT INTO purchase_order_items (po_id, product_id, quantity, unit_cost, received_qty)
		VALUES (?, 'p1', ?, ?, 0)`, poID, qty, unitCost).Error; err != nil {
		t.Fatalf("seed po item: %v", err)
	}
	if err := db.Exec(`INSERT INTO products (id, stock_quantity, weighted_avg_cost) VALUES ('p1', 0, 0)`).Error; err != nil {
		t.Fatalf("seed product: %v", err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity) VALUES ('wh-1', 'p1', 0)`).Error; err != nil {
		t.Fatalf("seed warehouse stock: %v", err)
	}
}

// TestReceivePO_RejectsUnknownProduct H12: a typo'd productID must not inflate
// an unrelated SKU or credit stock — the whole receive is rejected.
func TestReceivePO_RejectsUnknownProduct(t *testing.T) {
	db := setupSupplierTestDB(t)
	seedReceivePO(t, db, "po-1", 10, 5.0, modelsProduct.POStatusSent)
	r := NewSupplierRepository(db)

	err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 5, "typo-id": 1})
	if err == nil {
		t.Fatal("expected error for unknown productID")
	}
	if !strings.Contains(err.Error(), "not on PO") {
		t.Fatalf("unexpected error: %v", err)
	}
	// Nothing may have been applied: status stays sent, stock untouched.
	var po modelsProduct.PurchaseOrder
	if err := db.Where("id = 'po-1'").First(&po).Error; err != nil {
		t.Fatal(err)
	}
	if po.Status != modelsProduct.POStatusSent {
		t.Fatalf("PO status changed on rejected receive: %s", po.Status)
	}
	var ws modelsProduct.WarehouseStock
	if err := db.Where("warehouse_id = 'wh-1' AND product_id = 'p1'").First(&ws).Error; err != nil {
		t.Fatal(err)
	}
	if ws.Quantity != 0 {
		t.Fatalf("warehouse stock changed on rejected receive: %d", ws.Quantity)
	}
}

// TestReceivePO_RejectsZeroQty H12: qty <= 0 must be rejected.
func TestReceivePO_RejectsZeroQty(t *testing.T) {
	db := setupSupplierTestDB(t)
	seedReceivePO(t, db, "po-1", 10, 5.0, modelsProduct.POStatusSent)
	r := NewSupplierRepository(db)

	if err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 0}); err == nil {
		t.Fatal("expected error for zero qty")
	}
	if err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": -3}); err == nil {
		t.Fatal("expected error for negative qty")
	}
}

// TestReceivePO_RejectsOverOrderedQty H12: receiving past ordered - received
// must be rejected instead of silently creating a surplus batch.
func TestReceivePO_RejectsOverOrderedQty(t *testing.T) {
	db := setupSupplierTestDB(t)
	seedReceivePO(t, db, "po-1", 10, 5.0, modelsProduct.POStatusSent)
	r := NewSupplierRepository(db)

	err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 11})
	if err == nil {
		t.Fatal("expected error for over-order receive")
	}
	if !strings.Contains(err.Error(), "exceeds remaining ordered qty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestReceivePO_PartialThenCompleteReceive H12: valid partial receives still
// work; remaining ordered qty is enforced across calls and the PO flips to
// received once fully received.
func TestReceivePO_PartialThenCompleteReceive(t *testing.T) {
	db := setupSupplierTestDB(t)
	seedReceivePO(t, db, "po-1", 10, 5.0, modelsProduct.POStatusSent)
	r := NewSupplierRepository(db)

	// Partial receive of 6/10.
	if err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 6}); err != nil {
		t.Fatalf("partial receive: %v", err)
	}
	var po modelsProduct.PurchaseOrder
	if err := db.Where("id = 'po-1'").First(&po).Error; err != nil {
		t.Fatal(err)
	}
	if po.Status != modelsProduct.POStatusPartial {
		t.Fatalf("expected partially_received status, got %s", po.Status)
	}

	// Over the remaining 4 after the partial must fail.
	if err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 5}); err == nil {
		t.Fatal("expected error for over-receipt after partial")
	}

	// Completing the remaining 4/10 flips the PO to received.
	if err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 4}); err != nil {
		t.Fatalf("completing receive: %v", err)
	}
	if err := db.Where("id = 'po-1'").First(&po).Error; err != nil {
		t.Fatal(err)
	}
	if po.Status != modelsProduct.POStatusReceived {
		t.Fatalf("expected received status, got %s", po.Status)
	}

	var item modelsProduct.PurchaseOrderItem
	if err := db.Where("po_id = 'po-1' AND product_id = 'p1'").First(&item).Error; err != nil {
		t.Fatal(err)
	}
	if item.ReceivedQty != 10 {
		t.Fatalf("expected received_qty 10, got %d", item.ReceivedQty)
	}
}

// TestReceivePO_DoubleReceiveRejected H12: a second receive after the PO is
// fully received must be rejected (row-lock + status guard path).
func TestReceivePO_DoubleReceiveRejected(t *testing.T) {
	db := setupSupplierTestDB(t)
	seedReceivePO(t, db, "po-1", 10, 5.0, modelsProduct.POStatusSent)
	r := NewSupplierRepository(db)

	if err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 10}); err != nil {
		t.Fatalf("first receive: %v", err)
	}
	err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 5})
	if err == nil {
		t.Fatal("expected double-receive to be rejected")
	}
	if !strings.Contains(err.Error(), "already") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestReceivePO_ValidReceiveCreditsStock H12: a valid receive still credits
// warehouse + product stock, creates a batch at the PO's unit cost (so COGS is
// not dragged to zero), and recomputes the weighted average cost.
func TestReceivePO_ValidReceiveCreditsStock(t *testing.T) {
	db := setupSupplierTestDB(t)
	seedReceivePO(t, db, "po-1", 10, 5.0, modelsProduct.POStatusSent)
	r := NewSupplierRepository(db)

	if err := r.ReceivePO(context.Background(), "po-1", "wh-1", map[string]int{"p1": 4}); err != nil {
		t.Fatalf("receive: %v", err)
	}

	var ws modelsProduct.WarehouseStock
	if err := db.Where("warehouse_id = 'wh-1' AND product_id = 'p1'").First(&ws).Error; err != nil {
		t.Fatal(err)
	}
	if ws.Quantity != 4 {
		t.Fatalf("expected warehouse stock 4, got %d", ws.Quantity)
	}

	var p modelsProduct.Product
	if err := db.Where("id = 'p1'").First(&p).Error; err != nil {
		t.Fatal(err)
	}
	if p.StockQuantity != 4 {
		t.Fatalf("expected product stock 4, got %d", p.StockQuantity)
	}
	if p.WeightedAvgCost != 5.0 {
		t.Fatalf("expected weighted avg cost 5.0, got %v", p.WeightedAvgCost)
	}

	// A batch must exist carrying the PO's unit cost, not 0.
	var batch modelsProduct.ProductBatch
	if err := db.Where("product_id = 'p1' AND warehouse_id = 'wh-1'").First(&batch).Error; err != nil {
		t.Fatalf("expected a ProductBatch to be created: %v", err)
	}
	if batch.UnitCost != 5.0 {
		t.Fatalf("expected batch unit cost 5.0, got %v", batch.UnitCost)
	}
	if batch.Quantity != 4 {
		t.Fatalf("expected batch quantity 4, got %d", batch.Quantity)
	}

	// A stock transaction must be recorded.
	var count int64
	if err := db.Model(&modelsOrder.StockTransaction{}).
		Where("product_id = 'p1' AND reason = ?", modelsOrder.StockReasonGoodsReceived).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 stock transaction, got %d", count)
	}
}
