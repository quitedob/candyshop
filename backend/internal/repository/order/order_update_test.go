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

func setupOrderUpdateTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE users (id TEXT PRIMARY KEY, first_name TEXT, last_name TEXT, email TEXT)`,
		`CREATE TABLE products (id TEXT PRIMARY KEY, stock_quantity INTEGER NOT NULL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE warehouses (id TEXT PRIMARY KEY, code TEXT, is_active INTEGER DEFAULT 1, is_default INTEGER DEFAULT 0)`,
		`CREATE TABLE warehouse_stocks (id INTEGER PRIMARY KEY AUTOINCREMENT, warehouse_id TEXT, product_id TEXT, variant_id TEXT, quantity INTEGER, reserved INTEGER, updated_at DATETIME)`,
		`CREATE TABLE oem_project_inventory_holds (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, status TEXT)`,
		`CREATE TABLE product_batches (id TEXT PRIMARY KEY, product_id TEXT, batch_number TEXT, quantity INTEGER, expiry_date DATETIME, is_expired INTEGER DEFAULT 0, warehouse_id TEXT)`,
		`CREATE TABLE orders (
			id TEXT PRIMARY KEY,
			order_number TEXT UNIQUE NOT NULL,
			user_id TEXT NOT NULL,
			inquiry_id TEXT,
			source TEXT DEFAULT '',
			status TEXT DEFAULT 'pending',
			payment_status TEXT DEFAULT 'unpaid',
			stock_reserved INTEGER DEFAULT 0,
			items BLOB NOT NULL,
			warehouse_id TEXT,
			compliance_official_evidence INTEGER DEFAULT 0,
			subtotal REAL,
			tax_amount REAL DEFAULT 0,
			shipping_amount REAL DEFAULT 0,
			cogs REAL DEFAULT 0,
			total_amount REAL,
			currency TEXT DEFAULT 'USD',
			production_start_date DATETIME,
			estimated_completion DATETIME,
			actual_completion DATETIME,
			shipping_address BLOB,
			tracking_number TEXT,
			confirmed_at DATETIME,
			shipped_at DATETIME,
			delivered_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME,
			version INTEGER DEFAULT 0,
			deleted_at DATETIME
		)`,
		`CREATE TABLE stock_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT, change INTEGER, stock_before INTEGER, stock_after INTEGER, reason TEXT, reference_id TEXT, operator_id TEXT, batch_id TEXT, lot_number TEXT, warehouse_id TEXT, created_at DATETIME)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// TestUpdateWithStockAdjustment_PreservesPreloadedAssociations BUG-01: Save must not touch User association.
func TestUpdateWithStockAdjustment_PreservesPreloadedAssociations(t *testing.T) {
	db := setupOrderUpdateTestDB(t)
	if err := db.Exec(`INSERT INTO users (id, first_name, last_name, email) VALUES ('user-1', 'Test', 'Buyer', 'buyer@test.com')`).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 100)`).Error; err != nil {
		t.Fatal(err)
	}
	itemsJSON, _ := json.Marshal([]modelsOrder.OrderItem{{ProductID: "p1", Quantity: 5, UnitPrice: 1.0}})
	now := time.Now()
	if err := db.Exec(`INSERT INTO orders (id, order_number, user_id, status, payment_status, stock_reserved, items, subtotal, total_amount, currency, created_at, updated_at)
		VALUES ('ord-1', 'ORD-001', 'user-1', 'confirmed', 'paid', 1, ?, 5, 5, 'USD', ?, ?)`, itemsJSON, now, now).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewOrderRepository(db)
	order, err := repo.FindByID(context.Background(), "ord-1")
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if order.User == nil {
		t.Fatal("expected preloaded user")
	}

	order.PaymentStatus = "partial"
	order.UpdatedAt = time.Now()

	if err := repo.UpdateWithStockAdjustment(context.Background(), order, map[string]int{}); err != nil {
		t.Fatalf("UpdateWithStockAdjustment zero delta: %v", err)
	}

	var paymentStatus string
	if err := db.Raw(`SELECT payment_status FROM orders WHERE id = 'ord-1'`).Scan(&paymentStatus).Error; err != nil {
		t.Fatal(err)
	}
	if paymentStatus != "partial" {
		t.Fatalf("expected payment_status=partial, got %q", paymentStatus)
	}
}

// TestUpdateWithOptionalStockReservationAndOutbox_AtomicConfirm BUG-02: reserve + status update in one transaction.
func TestUpdateWithOptionalStockReservationAndOutbox_AtomicConfirm(t *testing.T) {
	db := setupOrderUpdateTestDB(t)
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 50)`).Error; err != nil {
		t.Fatal(err)
	}
	itemsJSON, _ := json.Marshal([]modelsOrder.OrderItem{{ProductID: "p1", Quantity: 5, UnitPrice: 1.0}})
	now := time.Now()
	if err := db.Exec(`INSERT INTO orders (id, order_number, user_id, status, payment_status, stock_reserved, items, subtotal, total_amount, currency, created_at, updated_at)
		VALUES ('ord-2', 'ORD-002', 'user-1', 'pending', 'paid', 0, ?, 5, 5, 'USD', ?, ?)`, itemsJSON, now, now).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewOrderRepository(db)
	order := &modelsOrder.Order{
		ID:            "ord-2",
		OrderNumber:   "ORD-002",
		UserID:        "user-1",
		Status:        "confirmed",
		PaymentStatus: "paid",
		Items:         modelsOrder.OrderItemArray{{ProductID: "p1", Quantity: 5, UnitPrice: 1.0}},
		Subtotal:      5,
		TotalAmount:   5,
		Currency:      "USD",
		UpdatedAt:     time.Now(),
	}

	if err := repo.UpdateWithOptionalStockReservationAndOutbox(context.Background(), order, map[string]int{"p1": 5}, true, nil); err != nil {
		t.Fatalf("atomic confirm: %v", err)
	}

	var status string
	var reserved int
	if err := db.Raw(`SELECT status, stock_reserved FROM orders WHERE id = 'ord-2'`).Row().Scan(&status, &reserved); err != nil {
		t.Fatal(err)
	}
	if status != "confirmed" {
		t.Fatalf("expected status=confirmed, got %q", status)
	}
	if reserved != 1 {
		t.Fatalf("expected stock_reserved=1, got %d", reserved)
	}
}
