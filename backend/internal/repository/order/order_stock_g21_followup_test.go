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

// setupG21FollowupTestDB creates an in-memory SQLite DB with the tables needed
// to exercise the G21 follow-up fixes: the draft-cleanup worker's audit-row
// atomicity. orders carries created_at because
// ReleaseExpiredPendingConfirmationOrders filters on it.
func setupG21FollowupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE products (id TEXT PRIMARY KEY, stock_quantity INTEGER NOT NULL DEFAULT 0, deleted_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE warehouses (id TEXT PRIMARY KEY, code TEXT, is_active INTEGER DEFAULT 1, is_default INTEGER DEFAULT 0)`,
		`CREATE TABLE warehouse_stocks (id INTEGER PRIMARY KEY AUTOINCREMENT, warehouse_id TEXT, product_id TEXT, quantity INTEGER, reserved INTEGER, updated_at DATETIME)`,
		`CREATE TABLE oem_project_inventory_holds (id TEXT PRIMARY KEY, product_id TEXT, quantity INTEGER, status TEXT)`,
		`CREATE TABLE product_batches (id TEXT PRIMARY KEY, product_id TEXT, batch_number TEXT, quantity INTEGER, expiry_date DATETIME, is_expired INTEGER DEFAULT 0, warehouse_id TEXT, created_at DATETIME, updated_at DATETIME)`,
		`CREATE TABLE orders (id TEXT PRIMARY KEY, user_id TEXT, status TEXT, stock_reserved INTEGER DEFAULT 0, items BLOB, created_at DATETIME, updated_at DATETIME, version INTEGER DEFAULT 0, deleted_at DATETIME)`,
		`CREATE TABLE stock_transactions (id INTEGER PRIMARY KEY AUTOINCREMENT, product_id TEXT, change INTEGER, stock_before INTEGER, stock_after INTEGER, reason TEXT, reference_id TEXT, operator_id TEXT, batch_id TEXT, lot_number TEXT, warehouse_id TEXT, created_at DATETIME)`,
	}
	for _, s := range stmts {
		if err := db.Exec(s).Error; err != nil {
			t.Fatalf("schema: %v", err)
		}
	}
	return db
}

// insertExpiredDraft inserts a pending_confirmation order (with reserved stock)
// created olderThanCutoff so the draft-cleanup worker is eligible to expire it.
func insertExpiredDraft(t *testing.T, db *gorm.DB, id, productID string, qty int, created time.Time) {
	t.Helper()
	itemsJSON, _ := json.Marshal(modelsOrder.OrderItemArray{{ProductID: productID, Quantity: qty}})
	if err := db.Exec(`INSERT INTO orders (id, user_id, status, stock_reserved, items, created_at, updated_at) VALUES (?, 'user-1', 'pending_confirmation', 1, ?, ?, ?)`,
		id, itemsJSON, created, time.Now()).Error; err != nil {
		t.Fatalf("insert order %s: %v", id, err)
	}
}

// TestReleaseExpiredOrders_WritesAuditRows documents the happy path of
// ReleaseExpiredPendingConfirmationOrders: the worker releases the reserved
// stock, marks the draft expired, AND persists the draft_expired audit rows in
// the same transaction. The audit trail is the only record of why the stock
// moved, so losing it would make the stock change unaccountable.
func TestReleaseExpiredOrders_WritesAuditRows(t *testing.T) {
	db := setupG21FollowupTestDB(t)
	now := time.Now()
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	insertExpiredDraft(t, db, "ord-ok", "p1", 5, now.Add(-2*time.Hour))

	repo := NewOrderRepository(db)
	released, err := repo.ReleaseExpiredPendingConfirmationOrders(context.Background(), now.Add(-time.Hour), 100)
	if err != nil {
		t.Fatalf("ReleaseExpiredPendingConfirmationOrders: %v", err)
	}
	if released != 1 {
		t.Fatalf("released = %d, want 1", released)
	}

	var order modelsOrder.Order
	if err := db.First(&order, "id = ?", "ord-ok").Error; err != nil {
		t.Fatal(err)
	}
	if order.Status != modelsOrder.OrderStatusExpired {
		t.Fatalf("order status = %q, want %q", order.Status, modelsOrder.OrderStatusExpired)
	}
	// The release restores the previously reserved 5 units to products.stock_quantity.
	var prodQty int
	if err := db.Raw(`SELECT stock_quantity FROM products WHERE id='p1'`).Scan(&prodQty).Error; err != nil {
		t.Fatal(err)
	}
	if prodQty != 15 {
		t.Fatalf("product stock = %d, want 15 (5 reserved units restored)", prodQty)
	}
	// The draft_expired audit row must be persisted with the release.
	var auditChange int
	if err := db.Raw(`SELECT COALESCE(SUM(change),0) FROM stock_transactions WHERE reference_id='ord-ok' AND reason='draft_expired'`).Scan(&auditChange).Error; err != nil {
		t.Fatal(err)
	}
	if auditChange != 5 {
		t.Fatalf("draft_expired audit change = %d, want 5 (release recorded in audit trail)", auditChange)
	}
}

// TestReleaseExpiredOrders_AuditFailureRollsBackRelease (G21-f) injects a hard
// failure into the stock_transactions insert for one order and verifies the
// release is atomic with its audit rows: the per-order transaction rolls back,
// leaving the draft pending_confirmation with its stock intact instead of
// committing the release + expired transition while the audit rows are silently
// lost. (The batch worker itself logs-and-continues per order by design, so the
// discriminator is that this order was NOT released.) Pre-fix the
// writeStockAuditEntries error was swallowed with log.Printf, so the order still
// flipped to expired and was counted as released=1; post-fix the error
// propagates, the transaction rolls back, and released stays 0.
func TestReleaseExpiredOrders_AuditFailureRollsBackRelease(t *testing.T) {
	db := setupG21FollowupTestDB(t)
	now := time.Now()
	if err := db.Exec(`INSERT INTO products (id, stock_quantity) VALUES ('p1', 10)`).Error; err != nil {
		t.Fatal(err)
	}
	// Abort only the audit insert for this order; the order status UPDATE would
	// otherwise still succeed, reproducing the pre-fix "silent audit loss" window.
	if err := db.Exec(`CREATE TRIGGER trg_fail_audit BEFORE INSERT ON stock_transactions
		WHEN NEW.reference_id = 'ord-fail'
		BEGIN
			SELECT RAISE(ABORT, 'injected audit failure');
		END`).Error; err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	insertExpiredDraft(t, db, "ord-fail", "p1", 5, now.Add(-2*time.Hour))

	repo := NewOrderRepository(db)
	released, err := repo.ReleaseExpiredPendingConfirmationOrders(context.Background(), now.Add(-time.Hour), 100)
	if err != nil {
		t.Fatalf("batch worker must not fail the whole run on one order: %v", err)
	}
	if released != 0 {
		t.Fatalf("released = %d, want 0 (the failing order must not be counted as released)", released)
	}

	// The per-order transaction rolled back: the order is still pending_confirmation
	// with its reserved stock intact, and no draft_expired audit rows were committed.
	var order modelsOrder.Order
	if err := db.First(&order, "id = ?", "ord-fail").Error; err != nil {
		t.Fatal(err)
	}
	if order.Status != modelsOrder.OrderStatusPendingConfirm {
		t.Fatalf("order status = %q, want %q (release must roll back when audit rows cannot be written)", order.Status, modelsOrder.OrderStatusPendingConfirm)
	}
	if !order.StockReserved {
		t.Fatalf("order stock_reserved = false, want true (reserved stock must survive the rollback)")
	}
	var auditCount int64
	if err := db.Model(&modelsOrder.StockTransaction{}).
		Where("reference_id = ? AND reason = ?", "ord-fail", modelsOrder.StockReasonDraftExpired).
		Count(&auditCount).Error; err != nil {
		t.Fatal(err)
	}
	if auditCount != 0 {
		t.Fatalf("draft_expired audit rows = %d, want 0 (rolled back)", auditCount)
	}
}
