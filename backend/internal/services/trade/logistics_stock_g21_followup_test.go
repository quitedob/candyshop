package trade

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsTrade "candypro/api/internal/models/trade"
	repoTrade "candypro/api/internal/repository/trade"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupLogisticsStockG21DB creates an in-memory SQLite DB with the tables the
// G21 dispatch FEFO follow-up fixes touch (products, warehouses, warehouse
// stocks, batches, orders, audit trail) plus the shipment tables DispatchShipment
// needs. No UNIQUE constraint on warehouse_stocks, mirroring production's
// non-unique-at-insert usage by these helpers.
func setupLogisticsStockG21DB(t *testing.T) *gorm.DB {
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
		`CREATE TABLE warehouses (id TEXT PRIMARY KEY, code TEXT, is_active INTEGER DEFAULT 1, is_default INTEGER DEFAULT 0)`,
		`CREATE TABLE warehouse_stocks (id INTEGER PRIMARY KEY AUTOINCREMENT, warehouse_id TEXT, product_id TEXT, quantity INTEGER, reserved INTEGER, updated_at DATETIME)`,
		`CREATE TABLE oem_project_inventory_holds (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, status TEXT)`,
		`CREATE TABLE product_batches (id TEXT PRIMARY KEY, product_id TEXT, batch_number TEXT, quantity INTEGER, expiry_date DATETIME, is_expired INTEGER DEFAULT 0, warehouse_id TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE orders (id TEXT PRIMARY KEY, warehouse_id TEXT, status TEXT, stock_reserved INTEGER DEFAULT 0, items BLOB, updated_at DATETIME, deleted_at DATETIME)`,
		`CREATE TABLE stock_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT, change INTEGER, stock_before INTEGER, stock_after INTEGER, reason TEXT, reference_id TEXT, operator_id TEXT, batch_id TEXT, lot_number TEXT, warehouse_id TEXT, created_at DATETIME)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// TestLogisticsDeductFEFO_SyncsWarehouseStock (G21 dispatch FEFO) verifies that
// the dispatch FEFO path — which decrements product_batches.quantity and
// Product.StockQuantity — also mirrors the deduction onto warehouse_stock. Pre-fix
// the warehouse mirror never moved and the warehouse-level aggregate drifted from
// the product level by the exact shipped qty.
func TestLogisticsDeductFEFO_SyncsWarehouseStock(t *testing.T) {
	db := setupLogisticsStockG21DB(t)
	future := time.Now().UTC().Add(24 * time.Hour)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 20, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO product_batches (id, product_id, batch_number, quantity, expiry_date, is_expired, warehouse_id, created_at, updated_at) VALUES ('b1', 'p1', 'LOT-1', 20, ?, 0, 'WH-A', ?, ?)`, future, time.Now(), time.Now()).Error; err != nil {
		t.Fatal(err)
	}

	recs, err := logisticsDeductFEFOFromBatches(db, "p1", 5, modelsOrder.StockReasonDispatched, "ord-1", "op", time.Now())
	if err != nil {
		t.Fatalf("FEFO deduct: %v", err)
	}
	if len(recs) != 1 || recs[0].Change != -5 {
		t.Fatalf("expected 1 audit record with change -5, got %+v", recs)
	}
	var whQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 15 {
		t.Fatalf("warehouse qty = %d, want 15 (dispatch FEFO must mirror onto warehouse_stock)", whQty)
	}
	var prodQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id='p1'`).Scan(&prodQty).Error; err != nil {
		t.Fatal(err)
	}
	if prodQty != 15 {
		t.Fatalf("product stock = %d, want 15", prodQty)
	}
	var batchQty int
	if err := db.Raw(`SELECT quantity FROM product_batches WHERE id='b1'`).Scan(&batchQty).Error; err != nil {
		t.Fatal(err)
	}
	if batchQty != 15 {
		t.Fatalf("batch qty = %d, want 15", batchQty)
	}
}

// TestLogisticsDeductFEFO_WarehouseSyncErrorPropagated (G21 dispatch FEFO)
// verifies that when the warehouse_stock mirror cannot absorb the FEFO deduction
// (insufficient quantity on the row), the dispatch fails loudly instead of
// reporting success while the warehouse aggregate drifts. Pre-fix no sync was
// attempted, so the deduction returned nil error.
func TestLogisticsDeductFEFO_WarehouseSyncErrorPropagated(t *testing.T) {
	db := setupLogisticsStockG21DB(t)
	future := time.Now().UTC().Add(24 * time.Hour)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	// A warehouse_stock row exists for p1 but holds only 3 units — too few to
	// mirror the 5-unit FEFO deduction.
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 3, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO product_batches (id, product_id, batch_number, quantity, expiry_date, is_expired, warehouse_id, created_at, updated_at) VALUES ('b1', 'p1', 'LOT-1', 20, ?, 0, 'WH-A', ?, ?)`, future, time.Now(), time.Now()).Error; err != nil {
		t.Fatal(err)
	}

	_, err := logisticsDeductFEFOFromBatches(db, "p1", 5, modelsOrder.StockReasonDispatched, "ord-1", "op", time.Now())
	if err == nil {
		t.Fatalf("expected FEFO warehouse sync error when warehouse_stock cannot mirror the deduction, got nil")
	}
	if !strings.Contains(err.Error(), "warehouse_stock sync failed") {
		t.Fatalf("expected warehouse_stock sync failure, got: %v", err)
	}
}

// TestLogisticsDeductReservedWarehouseStock_PortableClamp verifies the reserved
// drain in the reserved-warehouse deduction clamps at 0 (never negative) using
// SQL that runs on BOTH PostgreSQL and sqlite. Pre-fix the expression used the
// PostgreSQL-only GREATEST() and errored on sqlite with "no such function:
// GREATEST", so a dispatch over a reservation with reserved < qty could not be
// exercised by tests at all.
func TestLogisticsDeductReservedWarehouseStock_PortableClamp(t *testing.T) {
	db := setupLogisticsStockG21DB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 10, 3)`).Error; err != nil {
		t.Fatal(err)
	}

	// qty(5) > reserved(3): reserved must clamp to 0, not go negative.
	recs, err := logisticsDeductReservedWarehouseStock(db, "WH-A", "p1", 5, modelsOrder.StockReasonDispatched, "ord-1", "op", time.Now())
	if err != nil {
		t.Fatalf("reserved deduct: %v (deduction SQL must run on sqlite)", err)
	}
	if len(recs) != 1 {
		t.Fatalf("expected 1 audit record, got %d", len(recs))
	}
	var whQty, whRes int
	if err := db.Raw(`SELECT quantity, reserved FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Row().Scan(&whQty, &whRes); err != nil {
		t.Fatal(err)
	}
	if whQty != 5 || whRes != 0 {
		t.Fatalf("warehouse qty=%d reserved=%d, want 5/0", whQty, whRes)
	}
}

// dispatchFollowupOrderRepo returns a fixed order; Update is a no-op because the
// order-status advance after dispatch is not under assertion here.
type dispatchFollowupOrderRepo struct {
	order *modelsOrder.Order
}

func (f *dispatchFollowupOrderRepo) FindByID(ctx context.Context, id string) (*modelsOrder.Order, error) {
	return f.order, nil
}
func (f *dispatchFollowupOrderRepo) Update(ctx context.Context, order *modelsOrder.Order) error {
	return nil
}

type dispatchFollowupTradeRepo struct {
	trans *modelsTrade.TradeTransaction
}

func (f *dispatchFollowupTradeRepo) GetTransactionByID(ctx context.Context, id uint) (*modelsTrade.TradeTransaction, error) {
	return f.trans, nil
}

// TestDispatchShipment_FEFOSyncsWarehouseStock (G21 dispatch FEFO, end-to-end)
// reproduces the exact reachability the finding names: DispatchShipment on an
// order with StockReserved=false and a batched product routes through
// logisticsDeductFEFOFromBatches, which pre-fix left the warehouse_stock mirror
// untouched while product_batches and Product.StockQuantity moved. Post-fix the
// warehouse aggregate follows the FEFO deduction.
func TestDispatchShipment_FEFOSyncsWarehouseStock(t *testing.T) {
	db := setupLogisticsStockG21DB(t)
	future := time.Now().UTC().Add(24 * time.Hour)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 20)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouses (id, code, is_active, is_default) VALUES ('WH-A', 'MAIN', 1, 1)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO warehouse_stocks (warehouse_id, product_id, quantity, reserved) VALUES ('WH-A', 'p1', 20, 0)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO product_batches (id, product_id, batch_number, quantity, expiry_date, is_expired, warehouse_id, created_at, updated_at) VALUES ('b1', 'p1', 'LOT-1', 20, ?, 0, 'WH-A', ?, ?)`, future, time.Now(), time.Now()).Error; err != nil {
		t.Fatal(err)
	}

	shipRepo := repoTrade.NewShipmentRepository(db)
	shipment := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := shipRepo.Create(context.Background(), shipment); err != nil {
		t.Fatalf("create shipment: %v", err)
	}

	orderID := "ord-1"
	itemsJSON, _ := json.Marshal(modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}})
	if err := db.Exec(`INSERT INTO orders (id, warehouse_id, status, stock_reserved, items, updated_at) VALUES ('ord-1', 'WH-A', 'production', 0, ?, ?)`, itemsJSON, time.Now()).Error; err != nil {
		t.Fatalf("insert order: %v", err)
	}
	order := &modelsOrder.Order{ID: orderID, Status: "production", Items: modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5}}}
	trans := &modelsTrade.TradeTransaction{ID: 1, OrderID: &orderID}
	s := NewLogisticsService(
		shipRepo,
		repoTrade.NewShipmentEventRepository(db),
		&dispatchFollowupOrderRepo{order: order},
		&dispatchFollowupTradeRepo{trans: trans},
		db,
	)

	if err := s.DispatchShipment(context.Background(), shipment.ID, "op"); err != nil {
		t.Fatalf("DispatchShipment: %v", err)
	}

	var whQty int
	if err := db.Raw(`SELECT quantity FROM warehouse_stocks WHERE warehouse_id='WH-A' AND product_id='p1'`).Scan(&whQty).Error; err != nil {
		t.Fatal(err)
	}
	if whQty != 15 {
		t.Fatalf("warehouse qty = %d, want 15 (DispatchShipment FEFO must sync warehouse_stock when order.StockReserved=false)", whQty)
	}
	var prodQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id='p1'`).Scan(&prodQty).Error; err != nil {
		t.Fatal(err)
	}
	if prodQty != 15 {
		t.Fatalf("product stock = %d, want 15", prodQty)
	}
	var batchQty int
	if err := db.Raw(`SELECT quantity FROM product_batches WHERE id='b1'`).Scan(&batchQty).Error; err != nil {
		t.Fatal(err)
	}
	if batchQty != 15 {
		t.Fatalf("batch qty = %d, want 15", batchQty)
	}
}
