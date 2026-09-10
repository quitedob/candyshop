package order

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupStateGuardTestDB creates an in-memory SQLite and migrates the self-contained
// models exercised by the guarded state transitions below.
func setupStateGuardTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsOrder.ReturnRequest{}, &modelsOrder.StockTransfer{}, &modelsOrder.Fulfillment{}, &modelsOrder.Invoice{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

// TestReturnUpdateStatus_GuardsDoubleTransition proves the guarded conditional UPDATE
// blocks a second identical transition, so a double-click/retry cannot re-run the side
// effects (stock restore on refund) twice.
func TestReturnUpdateStatus_GuardsDoubleTransition(t *testing.T) {
	db := setupStateGuardTestDB(t)
	r := NewReturnRepository(db)

	if err := db.Create(&modelsOrder.ReturnRequest{
		ID:      "ret-1",
		OrderID: "ord-1",
		UserID:  "u1",
		Status:  modelsOrder.ReturnStatusPending,
		Reason:  modelsOrder.ReturnReasonDamaged,
	}).Error; err != nil {
		t.Fatalf("seed return: %v", err)
	}

	if err := r.UpdateStatus(context.Background(), "ret-1", modelsOrder.ReturnStatusApproved, "admin", ""); err != nil {
		t.Fatalf("first approve should succeed, got %v", err)
	}
	err := r.UpdateStatus(context.Background(), "ret-1", modelsOrder.ReturnStatusApproved, "admin", "")
	if err == nil {
		t.Fatal("second approve should fail with state mismatch")
	}
	if !errors.Is(err, ErrReturnStateMismatch) {
		t.Fatalf("expected ErrReturnStateMismatch, got %v", err)
	}
}

// TestStockTransferUpdateStatus_GuardsTerminalTransition proves a terminal transfer cannot
// be completed twice (the second caller loses the guarded UPDATE and never moves stock).
func TestStockTransferUpdateStatus_GuardsTerminalTransition(t *testing.T) {
	db := setupStateGuardTestDB(t)
	r := NewStockTransferRepository(db)

	if err := db.Create(&modelsOrder.StockTransfer{
		ID:              "tr-1",
		TransferNumber:  "TN-1",
		FromWarehouseID: "w1",
		ToWarehouseID:   "w2",
		Status:          modelsOrder.StockTransferStatusPending,
		CreatedBy:       "admin",
	}).Error; err != nil {
		t.Fatalf("seed transfer: %v", err)
	}

	if err := r.UpdateStatus(context.Background(), "tr-1", modelsOrder.StockTransferStatusCancelled, "admin"); err != nil {
		t.Fatalf("first cancel should succeed, got %v", err)
	}
	err := r.UpdateStatus(context.Background(), "tr-1", modelsOrder.StockTransferStatusCancelled, "admin")
	if err == nil {
		t.Fatal("second cancel should fail (already terminal)")
	}
	if !strings.Contains(err.Error(), "already in terminal status") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestFulfillmentCancel_GuardsTerminalStates proves Cancel is rejected for both terminal
// states — delivered and already-cancelled — so stock cannot be restored twice.
func TestFulfillmentCancel_GuardsTerminalStates(t *testing.T) {
	db := setupStateGuardTestDB(t)
	r := NewFulfillmentRepository(db)

	delivered := &modelsOrder.Fulfillment{
		ID:          "f-delivered",
		OrderID:     "ord-1",
		WarehouseID: "w1",
		Status:      modelsOrder.FulfillmentStatusDelivered,
	}
	cancelled := &modelsOrder.Fulfillment{
		ID:          "f-cancelled",
		OrderID:     "ord-1",
		WarehouseID: "w1",
		Status:      modelsOrder.FulfillmentStatusCancelled,
	}
	if err := db.Create(delivered).Error; err != nil {
		t.Fatalf("seed delivered: %v", err)
	}
	if err := db.Create(cancelled).Error; err != nil {
		t.Fatalf("seed cancelled: %v", err)
	}

	if err := r.Cancel(context.Background(), "f-delivered"); err == nil {
		t.Fatal("cancel of delivered fulfillment should fail")
	}
	if err := r.Cancel(context.Background(), "f-cancelled"); err == nil {
		t.Fatal("cancel of already-cancelled fulfillment should fail")
	}
}

// TestFulfillmentDeliver_GuardsDoubleDeliver proves the guarded conditional UPDATE
// blocks a second deliver, so a stale read cannot re-run the order-status advance.
func TestFulfillmentDeliver_GuardsDoubleDeliver(t *testing.T) {
	db := setupStateGuardTestDB(t)
	// Deliver advances the linked order's status, so a minimal orders table must
	// exist even though the order-status side effect is not under assertion here.
	if err := db.Exec(`CREATE TABLE orders (id TEXT PRIMARY KEY, status TEXT, delivered_at DATETIME, updated_at DATETIME, deleted_at DATETIME, version INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create orders table: %v", err)
	}
	if err := db.Exec(`INSERT INTO orders (id, status) VALUES ('ord-1', 'shipped')`).Error; err != nil {
		t.Fatalf("insert order: %v", err)
	}
	r := NewFulfillmentRepository(db)

	shipped := &modelsOrder.Fulfillment{
		ID:          "f-shipped",
		OrderID:     "ord-1",
		WarehouseID: "w1",
		Status:      modelsOrder.FulfillmentStatusShipped,
	}
	if err := db.Create(shipped).Error; err != nil {
		t.Fatalf("seed shipped: %v", err)
	}

	deliveredAt := time.Now()
	if err := r.Deliver(context.Background(), "f-shipped", deliveredAt); err != nil {
		t.Fatalf("first deliver should succeed, got %v", err)
	}
	err := r.Deliver(context.Background(), "f-shipped", deliveredAt)
	if err == nil {
		t.Fatal("second deliver should fail with state mismatch")
	}
	if !errors.Is(err, ErrFulfillmentStateMismatch) {
		t.Fatalf("expected ErrFulfillmentStateMismatch, got %v", err)
	}
}

// TestInvoiceMarkSent_GuardsStaleRead proves a stale read (still holding the
// pre-send status) loses the guarded conditional UPDATE.
func TestInvoiceMarkSent_GuardsStaleRead(t *testing.T) {
	db := setupStateGuardTestDB(t)
	r := NewInvoiceRepository(db)

	if err := r.Create(context.Background(), &modelsOrder.Invoice{ID: "inv-1", Status: modelsOrder.InvoiceStatusDraft}); err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if err := r.MarkSent(context.Background(), "inv-1", modelsOrder.InvoiceStatusDraft, time.Now()); err != nil {
		t.Fatalf("first send should succeed, got %v", err)
	}
	err := r.MarkSent(context.Background(), "inv-1", modelsOrder.InvoiceStatusDraft, time.Now())
	if !errors.Is(err, ErrInvoiceStateMismatch) {
		t.Fatalf("expected ErrInvoiceStateMismatch, got %v", err)
	}
}

// TestOrderUpdateStatusGuarded_StaleRead proves the guarded order-status write
// rejects a transition keyed on an outdated status.
func TestOrderUpdateStatusGuarded_StaleRead(t *testing.T) {
	db := setupStateGuardTestDB(t)
	if err := db.Exec(`CREATE TABLE orders (id TEXT PRIMARY KEY, status TEXT, updated_at DATETIME, deleted_at DATETIME, version INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create orders table: %v", err)
	}
	if err := db.Exec(`INSERT INTO orders (id, status) VALUES ('ord-1', 'production')`).Error; err != nil {
		t.Fatalf("insert order: %v", err)
	}
	r := NewOrderRepository(db)

	if err := r.UpdateStatusGuarded(context.Background(), "ord-1", "production", "shipped", nil); err != nil {
		t.Fatalf("first transition should succeed, got %v", err)
	}
	err := r.UpdateStatusGuarded(context.Background(), "ord-1", "production", "shipped", nil)
	if !errors.Is(err, ErrOrderStateMismatch) {
		t.Fatalf("expected ErrOrderStateMismatch, got %v", err)
	}
}
