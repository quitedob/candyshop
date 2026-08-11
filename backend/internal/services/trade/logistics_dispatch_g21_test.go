package trade

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsTrade "candypro/api/internal/models/trade"
	repoTrade "candypro/api/internal/repository/trade"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// dispatchFakeOrderRepo returns a fixed order; its Update is a no-op because the
// order-status advance after dispatch is exercised by the service but not under
// assertion here.
type dispatchFakeOrderRepo struct {
	order *modelsOrder.Order
}

func (f *dispatchFakeOrderRepo) FindByID(ctx context.Context, id string) (*modelsOrder.Order, error) {
	return f.order, nil
}
func (f *dispatchFakeOrderRepo) Update(ctx context.Context, order *modelsOrder.Order) error {
	return nil
}

type dispatchFakeTradeRepo struct {
	trans *modelsTrade.TradeTransaction
}

func (f *dispatchFakeTradeRepo) GetTransactionByID(ctx context.Context, id uint) (*modelsTrade.TradeTransaction, error) {
	return f.trans, nil
}

func setupDispatchG21DB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsTrade.ShipmentTracking{}, &modelsTrade.ShipmentEvent{}); err != nil {
		t.Fatalf("migrate shipment: %v", err)
	}
	stmts := []string{
		`CREATE TABLE products (id TEXT PRIMARY KEY, stock_quantity INTEGER NOT NULL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE oem_project_inventory_holds (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, status TEXT)`,
		`CREATE TABLE warehouses (id TEXT PRIMARY KEY, code TEXT, is_active INTEGER DEFAULT 1, is_default INTEGER DEFAULT 0)`,
		`CREATE TABLE warehouse_stocks (id INTEGER PRIMARY KEY AUTOINCREMENT, warehouse_id TEXT, product_id TEXT, quantity INTEGER, reserved INTEGER, updated_at DATETIME)`,
		`CREATE TABLE product_batches (id TEXT PRIMARY KEY, product_id TEXT, batch_number TEXT, quantity INTEGER, expiry_date DATETIME, is_expired INTEGER DEFAULT 0, warehouse_id TEXT)`,
		`CREATE TABLE stock_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT, change INTEGER, stock_before INTEGER, stock_after INTEGER, reason TEXT, reference_id TEXT, operator_id TEXT, batch_id TEXT, lot_number TEXT, warehouse_id TEXT, created_at DATETIME)`,
		// orders is needed because DispatchShipment re-reads (and locks) the order
		// row inside the transaction to compute the unfulfilled remainder.
		`CREATE TABLE orders (id TEXT PRIMARY KEY, warehouse_id TEXT, status TEXT, stock_reserved INTEGER DEFAULT 0, items BLOB, updated_at DATETIME, deleted_at DATETIME)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// TestDispatchShipment_DeductsOnlyUnfulfilledRemainder (G21-d) verifies that the
// trade dispatch path deducts only the quantity not already issued by a
// fulfillment (item.Quantity - item.FulfilledQuantity) and writes its deduction
// to the stock audit trail, so an order flowing through both the fulfillment and
// dispatch paths is not double-deducted.
func TestDispatchShipment_DeductsOnlyUnfulfilledRemainder(t *testing.T) {
	db := setupDispatchG21DB(t)
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 0)`).Error; err != nil {
		t.Fatal(err)
	}

	shipRepo := repoTrade.NewShipmentRepository(db)
	shipment := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := shipRepo.Create(context.Background(), shipment); err != nil {
		t.Fatalf("create shipment: %v", err)
	}

	// 4 of the 10 units were already issued by a fulfillment. The order must live
	// in the DB because DispatchShipment re-reads (and locks) the order row inside
	// the transaction; the fake repo copy is now only an existence check.
	orderID := "ord-1"
	itemsJSON, _ := json.Marshal(modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 10, FulfilledQuantity: 4}})
	if err := db.Exec(`INSERT INTO orders (id, warehouse_id, status, stock_reserved, items, updated_at) VALUES ('ord-1', 'WH-A', 'production', 0, ?, ?)`, itemsJSON, time.Now()).Error; err != nil {
		t.Fatalf("insert order: %v", err)
	}
	order := &modelsOrder.Order{
		ID:     orderID,
		Status: "production",
		Items:  modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 10, FulfilledQuantity: 4}},
	}
	trans := &modelsTrade.TradeTransaction{ID: 1, OrderID: &orderID}
	s := NewLogisticsService(
		shipRepo,
		repoTrade.NewShipmentEventRepository(db),
		&dispatchFakeOrderRepo{order: order},
		&dispatchFakeTradeRepo{trans: trans},
		db,
	)

	if err := s.DispatchShipment(context.Background(), shipment.ID, "op"); err != nil {
		t.Fatalf("DispatchShipment: %v", err)
	}

	// Only the 6 unfulfilled units may be deducted from the warehouse.
	var whQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 4 {
		t.Fatalf("warehouse qty = %d, want 4 (only 6 of 10 deducted)", whQty)
	}
	var prodQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id='p1'`).Scan(&prodQty).Error; err != nil {
		t.Fatal(err)
	}
	if prodQty != 4 {
		t.Fatalf("product stock = %d, want 4", prodQty)
	}
	// The dispatch path writes its deduction to the stock audit trail so a later
	// fulfillment on the same order can detect it and skip its own deduction.
	var auditChange int
	if err := db.Raw(`SELECT COALESCE(SUM(change),0) FROM stock_transactions WHERE reference_id='ord-1' AND reason='dispatched'`).Scan(&auditChange).Error; err != nil {
		t.Fatal(err)
	}
	if auditChange != -6 {
		t.Fatalf("dispatched audit change = %d, want -6", auditChange)
	}
}

// TestDispatchShipment_SecondShipmentSkipsDeduction (G21-d, arguer point 2)
// verifies dispatch is idempotent per order: ShipmentTracking carries no
// quantity, so the FIRST dispatch on an order issues the entire unfulfilled
// remainder (StockReasonDispatched audit rows). A second shipment on the same
// order must NOT re-deduct that stock — otherwise two shipments silently
// double-deduct when warehouse stock >= 2x order qty (or fail loudly on the
// second when stock < 2x). Pre-fix the warehouse drops to 0 (audit -20);
// post-fix it stays 10 (audit -10).
func TestDispatchShipment_SecondShipmentSkipsDeduction(t *testing.T) {
	db := setupDispatchG21DB(t)
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 20, 0)`).Error; err != nil {
		t.Fatal(err)
	}

	shipRepo := repoTrade.NewShipmentRepository(db)
	shipA := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := shipRepo.Create(context.Background(), shipA); err != nil {
		t.Fatalf("create shipment A: %v", err)
	}
	shipB := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := shipRepo.Create(context.Background(), shipB); err != nil {
		t.Fatalf("create shipment B: %v", err)
	}

	orderID := "ord-1"
	itemsJSON, _ := json.Marshal(modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 10, FulfilledQuantity: 0}})
	if err := db.Exec(`INSERT INTO orders (id, warehouse_id, status, stock_reserved, items, updated_at) VALUES ('ord-1', 'WH-A', 'production', 0, ?, ?)`, itemsJSON, time.Now()).Error; err != nil {
		t.Fatalf("insert order: %v", err)
	}
	order := &modelsOrder.Order{
		ID:     orderID,
		Status: "production",
		Items:  modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 10, FulfilledQuantity: 0}},
	}
	trans := &modelsTrade.TradeTransaction{ID: 1, OrderID: &orderID}
	s := NewLogisticsService(
		shipRepo,
		repoTrade.NewShipmentEventRepository(db),
		&dispatchFakeOrderRepo{order: order},
		&dispatchFakeTradeRepo{trans: trans},
		db,
	)

	// First dispatch issues the whole 10.
	if err := s.DispatchShipment(context.Background(), shipA.ID, "op"); err != nil {
		t.Fatalf("DispatchShipment A: %v", err)
	}
	var whQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 10 {
		t.Fatalf("after first dispatch warehouse qty = %d, want 10", whQty)
	}

	// Second shipment on the same order must succeed but NOT re-deduct.
	if err := s.DispatchShipment(context.Background(), shipB.ID, "op"); err != nil {
		t.Fatalf("DispatchShipment B (legitimate second shipment) failed: %v", err)
	}
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 10 {
		t.Fatalf("after second dispatch warehouse qty = %d, want 10 (must not double-deduct)", whQty)
	}
	var prodQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id='p1'`).Scan(&prodQty).Error; err != nil {
		t.Fatal(err)
	}
	if prodQty != 10 {
		t.Fatalf("product stock = %d, want 10", prodQty)
	}
	var auditChange int
	if err := db.Raw(`SELECT COALESCE(SUM(change),0) FROM stock_transactions WHERE reference_id='ord-1' AND reason='dispatched'`).Scan(&auditChange).Error; err != nil {
		t.Fatal(err)
	}
	if auditChange != -10 {
		t.Fatalf("dispatched audit change = %d, want -10 (one goods-issue for the order)", auditChange)
	}
}

// TestDispatchShipment_ReReadsOrderInsideTx (G21-d, arguer point 2) simulates the
// concurrency window the arguer flagged: the outer FindByID snapshot is taken
// BEFORE a fulfillment commits (FulfilledQuantity=0 in the snapshot) while the DB
// order row already reflects the fulfillment (FulfilledQuantity=4). The dispatch
// must ignore the stale snapshot and compute the remaining deduction from a fresh,
// FOR UPDATE locked re-read inside the transaction; otherwise it would deduct the
// full 10 units a second time. Pre-fix (iterating the stale snapshot) the
// warehouse drops to 0; post-fix only 6 units are deducted and the warehouse holds 4.
func TestDispatchShipment_ReReadsOrderInsideTx(t *testing.T) {
	db := setupDispatchG21DB(t)
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 0)`).Error; err != nil {
		t.Fatal(err)
	}

	shipRepo := repoTrade.NewShipmentRepository(db)
	shipment := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := shipRepo.Create(context.Background(), shipment); err != nil {
		t.Fatalf("create shipment: %v", err)
	}

	orderID := "ord-1"
	// The DB order row reflects a fulfillment that has already issued 4 of 10 units.
	itemsJSON, _ := json.Marshal(modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 10, FulfilledQuantity: 4}})
	if err := db.Exec(`INSERT INTO orders (id, warehouse_id, status, stock_reserved, items, updated_at) VALUES ('ord-1', 'WH-A', 'production', 0, ?, ?)`, itemsJSON, time.Now()).Error; err != nil {
		t.Fatalf("insert order: %v", err)
	}
	// The outer FindByID returns a STALE snapshot as if it predated the fulfillment
	// commit. The dispatch must not use it for the deduction math.
	stale := &modelsOrder.Order{
		ID:     orderID,
		Status: "production",
		Items:  modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 10, FulfilledQuantity: 0}},
	}
	trans := &modelsTrade.TradeTransaction{ID: 1, OrderID: &orderID}
	s := NewLogisticsService(
		shipRepo,
		repoTrade.NewShipmentEventRepository(db),
		&dispatchFakeOrderRepo{order: stale},
		&dispatchFakeTradeRepo{trans: trans},
		db,
	)

	if err := s.DispatchShipment(context.Background(), shipment.ID, "op"); err != nil {
		t.Fatalf("DispatchShipment: %v", err)
	}

	var whQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 4 {
		t.Fatalf("warehouse qty = %d, want 4 (dispatch must use the fresh FulfilledQuantity=4 from the locked DB re-read, not the stale snapshot -> full 10 deducted)", whQty)
	}
	// Only 6 units may be recorded against the dispatch audit.
	var auditChange int
	if err := db.Raw(`SELECT COALESCE(SUM(change),0) FROM stock_transactions WHERE reference_id='ord-1' AND reason='dispatched'`).Scan(&auditChange).Error; err != nil {
		t.Fatal(err)
	}
	if auditChange != -6 {
		t.Fatalf("dispatched audit change = %d, want -6", auditChange)
	}
}
