package order

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupG21StockTestDB creates an in-memory SQLite DB with the tables needed to
// exercise the G21 stock-allocation fixes (warehouse transfer, FEFO sync,
// warehouse deduction, fulfillment create/ship).
func setupG21StockTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE products (id TEXT PRIMARY KEY, stock_quantity INTEGER NOT NULL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE warehouses (id TEXT PRIMARY KEY, code TEXT, is_active INTEGER DEFAULT 1, is_default INTEGER DEFAULT 0)`,
		// Unique (warehouse_id, product_id) mirrors the production idx_wh_prod index,
		// required for the additive ON CONFLICT upsert in addStockToWarehouseUpsert.
		`CREATE TABLE warehouse_stocks (id INTEGER PRIMARY KEY AUTOINCREMENT, warehouse_id TEXT, product_id TEXT, variant_id TEXT, quantity INTEGER, reserved INTEGER, updated_at DATETIME, UNIQUE(warehouse_id, product_id))`,
		`CREATE TABLE oem_project_inventory_holds (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, status TEXT)`,
		`CREATE TABLE product_batches (id TEXT PRIMARY KEY, product_id TEXT, batch_number TEXT, quantity INTEGER, expiry_date DATETIME, is_expired INTEGER DEFAULT 0, warehouse_id TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE orders (id TEXT PRIMARY KEY, user_id TEXT, status TEXT, stock_reserved INTEGER DEFAULT 0, items BLOB, updated_at DATETIME, version INTEGER DEFAULT 0, deleted_at DATETIME)`,
		`CREATE TABLE stock_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT, change INTEGER, stock_before INTEGER, stock_after INTEGER, reason TEXT, reference_id TEXT, operator_id TEXT, batch_id TEXT, lot_number TEXT, warehouse_id TEXT, created_at DATETIME)`,
		`CREATE TABLE fulfillments (id TEXT PRIMARY KEY, order_id TEXT, warehouse_id TEXT, tracking_number TEXT, carrier TEXT, status TEXT, notes TEXT, shipped_at DATETIME, delivered_at DATETIME, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE fulfillment_items (id INTEGER PRIMARY KEY AUTOINCREMENT, fulfillment_id TEXT, order_item_idx INTEGER, product_id TEXT, quantity INTEGER)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// TestTransferStockBetweenWarehouses_CreatesTargetRow (G21-a) verifies that a
// transfer into a warehouse that has no warehouse_stock row yet creates the row
// instead of silently dropping the transferred stock. Pre-fix the plain UPDATE
// affected zero rows and the 4 units vanished.
func TestTransferStockBetweenWarehouses_CreatesTargetRow(t *testing.T) {
	db := setupG21StockTestDB(t)
	now := time.Now()
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'A', 1, 1), ('WH-B', 'B', 1, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 0)`).Error; err != nil {
		t.Fatal(err)
	}

	recs, err := TransferStockBetweenWarehouses(db, "WH-A", "WH-B", "p1", 4, modelsOrder.StockReasonStockTransfer, "t-1", "op", now)
	if err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(recs))
	}
	var srcQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&srcQty).Error; err != nil {
		t.Fatal(err)
	}
	if srcQty != 6 {
		t.Fatalf("source qty = %d, want 6", srcQty)
	}
	var dstCount int64
	if err := db.Model(&modelsProduct.WarehouseStock{}).Where("warehouse_id = ? AND product_id = ?", "WH-B", "p1").Count(&dstCount).Error; err != nil {
		t.Fatal(err)
	}
	if dstCount != 1 {
		t.Fatalf("target warehouse_stock row was not created (count=%d): transferred stock dropped", dstCount)
	}
	var dstQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-B' AND product_id='p1'`).Scan(&dstQty).Error; err != nil {
		t.Fatal(err)
	}
	if dstQty != 4 {
		t.Fatalf("target qty = %d, want 4", dstQty)
	}
}

// TestTransferStockBetweenWarehouses_MovesReserved (G21-a edge) verifies that the
// customer reservations sitting on the moved goods travel with them: reserved is
// decremented on the source and carried onto the target row, so the sum of
// reserved across warehouses is preserved and the target sellable is not inflated.
// It also exercises the additive ON CONFLICT path by transferring into a target
// that already has a row (the previous First-then-Create helper dropped reserved to
// 0 on create and never updated it on the target).
func TestTransferStockBetweenWarehouses_MovesReserved(t *testing.T) {
	db := setupG21StockTestDB(t)
	now := time.Now()
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'A', 1, 1), ('WH-B', 'B', 1, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 5)`).Error; err != nil {
		t.Fatal(err)
	}

	// First transfer: 4 units (4 of them reserved) into an empty target WH-B.
	if _, err := TransferStockBetweenWarehouses(db, "WH-A", "WH-B", "p1", 4, modelsOrder.StockReasonStockTransfer, "t-1", "op", now); err != nil {
		t.Fatalf("transfer 1: %v", err)
	}
	var srcQty, srcRes int
	if err := db.Raw(`SELECT quantity, reserved FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Row().Scan(&srcQty, &srcRes); err != nil {
		t.Fatal(err)
	}
	if srcQty != 6 || srcRes != 1 {
		t.Fatalf("source after transfer 1: qty=%d reserved=%d, want 6/1", srcQty, srcRes)
	}
	var dstQty, dstRes int
	if err := db.Raw(`SELECT quantity, reserved FROM warehouse_stocks WHERE warehouse_id='WH-B' AND product_id='p1'`).Row().Scan(&dstQty, &dstRes); err != nil {
		t.Fatalf("target row was not created: %v", err)
	}
	if dstQty != 4 || dstRes != 4 {
		t.Fatalf("target after transfer 1: qty=%d reserved=%d, want 4/4 (reserved must move with the goods)", dstQty, dstRes)
	}

	// Second transfer: 3 more units. Source has 1 reserved left, so 1 reserved
	// moves and 2 unreserved. The target already has a row -> ON CONFLICT update.
	if _, err := TransferStockBetweenWarehouses(db, "WH-A", "WH-B", "p1", 3, modelsOrder.StockReasonStockTransfer, "t-2", "op", now); err != nil {
		t.Fatalf("transfer 2: %v", err)
	}
	if err := db.Raw(`SELECT quantity, reserved FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Row().Scan(&srcQty, &srcRes); err != nil {
		t.Fatal(err)
	}
	if srcQty != 3 || srcRes != 0 {
		t.Fatalf("source after transfer 2: qty=%d reserved=%d, want 3/0", srcQty, srcRes)
	}
	if err := db.Raw(`SELECT quantity, reserved FROM warehouse_stocks WHERE warehouse_id='WH-B' AND product_id='p1'`).Row().Scan(&dstQty, &dstRes); err != nil {
		t.Fatal(err)
	}
	if dstQty != 7 || dstRes != 5 {
		t.Fatalf("target after transfer 2: qty=%d reserved=%d, want 7/5", dstQty, dstRes)
	}
	// Invariant: total reserved across warehouses is preserved (5 originally).
	if dstRes+srcRes != 5 {
		t.Fatalf("total reserved = %d, want 5 preserved across transfer", dstRes+srcRes)
	}
}

// TestDeductFEFOWithWarehouse_SyncErrorPropagated (G21-b) verifies that a FEFO
// batch deduction whose warehouse_stock mirror cannot be updated returns an
// error (so the caller rolls back) instead of log-only success. Pre-fix the sync
// failure was swallowed and the deduction reported success.
func TestDeductFEFOWithWarehouse_SyncErrorPropagated(t *testing.T) {
	db := setupG21StockTestDB(t)
	now := time.Now().UTC().Add(24 * time.Hour)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'A', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	// Batch exists in WH-A, but there is no warehouse_stock row to mirror onto.
	if err := db.Exec(`INSERT INTO product_batches (id, product_id, batch_number, quantity, expiry_date, is_expired, warehouse_id, created_at, updated_at) VALUES ('b1', 'p1', 'LOT-1', 20, ?, 0, 'WH-A', ?, ?)`, now, now, now).Error; err != nil {
		t.Fatal(err)
	}

	_, err := deductFEFOFromBatchesWithWarehouse(db, "WH-A", "p1", 5, modelsOrder.StockReasonStockDeducted, "ord-1", "op", time.Now())
	if err == nil {
		t.Fatalf("expected FEFO warehouse sync error when warehouse_stock row is missing, got nil")
	}
	if !strings.Contains(err.Error(), "warehouse_stock sync failed") {
		t.Fatalf("expected warehouse_stock sync failure, got: %v", err)
	}
}

// TestDeductWarehouseStock_DecrementsProductStock (G21-c) verifies a direct
// (unreserved) warehouse deduction also decrements Product.StockQuantity so the
// product-level aggregate and warehouse-level stock stay in sync. Pre-fix the
// product row was read but never written.
func TestDeductWarehouseStock_DecrementsProductStock(t *testing.T) {
	db := setupG21StockTestDB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'A', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 0)`).Error; err != nil {
		t.Fatal(err)
	}

	recs, err := deductWarehouseStock(db, "WH-A", "p1", 3, false, modelsOrder.StockReasonGoodsIssued, "ord-1", "op", time.Now())
	if err != nil {
		t.Fatalf("deduct: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(recs))
	}
	if recs[0].StockAfter != 7 {
		t.Fatalf("audit StockAfter = %d, want 7", recs[0].StockAfter)
	}
	var productQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id='p1'`).Scan(&productQty).Error; err != nil {
		t.Fatal(err)
	}
	if productQty != 7 {
		t.Fatalf("product stock = %d, want 7", productQty)
	}
	var whQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 7 {
		t.Fatalf("warehouse qty = %d, want 7", whQty)
	}
}

// TestDeductWarehouseStock_ReservedKeepsProductStock (G21-c) verifies that when
// the goods were previously reserved (Product.StockQuantity already decremented
// at reservation time) a reserved deduction does NOT decrement it again.
func TestDeductWarehouseStock_ReservedKeepsProductStock(t *testing.T) {
	db := setupG21StockTestDB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'A', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 3)`).Error; err != nil {
		t.Fatal(err)
	}

	if _, err := deductWarehouseStock(db, "WH-A", "p1", 3, true, modelsOrder.StockReasonGoodsIssued, "ord-1", "op", time.Now()); err != nil {
		t.Fatalf("deduct: %v", err)
	}
	var productQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id='p1'`).Scan(&productQty).Error; err != nil {
		t.Fatal(err)
	}
	if productQty != 10 {
		t.Fatalf("product stock = %d, want 10 (must not be decremented when reserved)", productQty)
	}
	var whQty, whReserved int
	if err := db.Raw(`SELECT quantity, reserved FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Row().Scan(&whQty, &whReserved); err != nil {
		t.Fatal(err)
	}
	if whQty != 7 || whReserved != 0 {
		t.Fatalf("warehouse qty=%d reserved=%d, want 7/0", whQty, whReserved)
	}
}

// TestFulfillmentCreate_SkipsDeductionWhenOrderDispatched (G21-d) verifies that
// a fulfillment on an order whose stock was already issued by the trade dispatch
// path does not deduct the same stock a second time. Pre-fix the warehouse
// deduction ran unconditionally and the order was double-deducted.
func TestFulfillmentCreate_SkipsDeductionWhenOrderDispatched(t *testing.T) {
	db := setupG21StockTestDB(t)
	now := time.Now()
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'A', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	itemsJSON, _ := json.Marshal(modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}})
	if err := db.Exec(`INSERT INTO orders (id, user_id, status, stock_reserved, items, updated_at) VALUES ('ord-1', 'user-1', 'production', 1, ?, ?)`, itemsJSON, now).Error; err != nil {
		t.Fatal(err)
	}
	// Simulate the trade dispatch path having already issued this order's stock.
	if err := db.Exec(`INSERT INTO stock_transactions (product_id, change, stock_before, stock_after, reason, reference_id, operator_id, created_at) VALUES ('p1', -5, 10, 5, 'dispatched', 'ord-1', 'op', ?)`, now).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewFulfillmentRepository(db)
	fulfillment := &modelsOrder.Fulfillment{ID: "FUL1", OrderID: "ord-1", WarehouseID: "WH-A", Status: modelsOrder.FulfillmentStatusPending}
	items := []modelsOrder.FulfillmentItem{{OrderItemIdx: 0, ProductID: "p1", Quantity: 5}}
	if err := repo.Create(context.Background(), fulfillment, items, "WH-A"); err != nil {
		t.Fatalf("fulfillment create: %v", err)
	}

	var whQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 10 {
		t.Fatalf("warehouse qty = %d, want 10 (deduction must be skipped after dispatch)", whQty)
	}
	var order modelsOrder.Order
	if err := db.First(&order, "id = ?", "ord-1").Error; err != nil {
		t.Fatal(err)
	}
	if len(order.Items) == 0 || order.Items[0].FulfilledQuantity != 5 {
		t.Fatalf("FulfilledQuantity not booked, got items=%+v", order.Items)
	}
}

// TestFulfillmentShip_AccumulatesAndRejectsDoubleShip (G21-e) exercises the
// Ship path: ShippedQuantity accumulates across separate fulfillments of the
// same order and the same fulfillment cannot be shipped twice. The conditional
// status transition plus the order row lock prevent lost ShippedQuantity updates
// under concurrent ships (true concurrency is not reproducible on in-memory
// sqlite, so the invariant is checked sequentially here).
func TestFulfillmentShip_AccumulatesAndRejectsDoubleShip(t *testing.T) {
	db := setupG21StockTestDB(t)
	now := time.Now()
	itemsJSON, _ := json.Marshal(modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 10}})
	if err := db.Exec(`INSERT INTO orders (id, user_id, status, stock_reserved, items, updated_at) VALUES ('ord-1', 'user-1', 'production', 0, ?, ?)`, itemsJSON, now).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO fulfillments (id, order_id, warehouse_id, status, created_at, updated_at) VALUES ('FUL1', 'ord-1', 'WH-A', 'pending', ?, ?)`, now, now).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO fulfillments (id, order_id, warehouse_id, status, created_at, updated_at) VALUES ('FUL2', 'ord-1', 'WH-A', 'pending', ?, ?)`, now, now).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO fulfillment_items (fulfillment_id, order_item_idx, product_id, quantity) VALUES ('FUL1', 0, 'p1', 3)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO fulfillment_items (fulfillment_id, order_item_idx, product_id, quantity) VALUES ('FUL2', 0, 'p1', 3)`).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewFulfillmentRepository(db)
	if err := repo.Ship(context.Background(), "FUL1", "TRK1", "DHL", now); err != nil {
		t.Fatalf("ship FUL1: %v", err)
	}
	if err := repo.Ship(context.Background(), "FUL2", "TRK2", "DHL", now); err != nil {
		t.Fatalf("ship FUL2: %v", err)
	}
	var order modelsOrder.Order
	if err := db.First(&order, "id = ?", "ord-1").Error; err != nil {
		t.Fatal(err)
	}
	if len(order.Items) == 0 || order.Items[0].ShippedQuantity != 6 {
		t.Fatalf("ShippedQuantity = %+v, want 6", order.Items)
	}
	if err := repo.Ship(context.Background(), "FUL1", "TRK1b", "DHL", now); err == nil {
		t.Fatalf("expected error when shipping an already-shipped fulfillment")
	}
}
