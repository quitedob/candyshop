package order

import (
	"context"
	"strings"
	"testing"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupPaymentTestDB creates an in-memory SQLite with minimal orders + payments
// tables for the balance-check lock tests.
func setupPaymentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE orders (
		id TEXT PRIMARY KEY, total_amount REAL,
		deleted_at DATETIME, created_at DATETIME, updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create orders: %v", err)
	}
	if err := db.Exec(`CREATE TABLE payments (
		id TEXT PRIMARY KEY, order_id TEXT, amount REAL, currency TEXT, method TEXT, status TEXT,
		reference TEXT, proof_url TEXT, notes TEXT,
		confirmed_by TEXT, confirmed_at DATETIME,
		gateway_transaction_id TEXT, authorized_amount REAL, captured_amount REAL, refunded_amount REAL,
		gateway_metadata TEXT, created_at DATETIME, updated_at DATETIME, version INTEGER, deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create payments: %v", err)
	}
	if err := db.Exec(`INSERT INTO orders (id, total_amount) VALUES ('ord-1', 100)`).Error; err != nil {
		t.Fatalf("seed order: %v", err)
	}
	if err := db.Exec(`INSERT INTO payments (id, order_id, amount, status) VALUES ('pay-1', 'ord-1', 60, 'confirmed')`).Error; err != nil {
		t.Fatalf("seed confirmed payment: %v", err)
	}
	return db
}

func TestCreateWithBalanceCheck_AllowsWithinRemaining(t *testing.T) {
	db := setupPaymentTestDB(t)
	r := NewPaymentRepository(db)

	pay := &modelsOrder.Payment{ID: "pay-2", OrderID: "ord-1", Amount: 40, Status: "pending"}
	if err := r.CreateWithBalanceCheck(context.Background(), 100, pay); err != nil {
		t.Fatalf("expected payment within remaining balance to succeed, got %v", err)
	}
}

func TestCreateWithBalanceCheck_RejectsExceedingRemaining(t *testing.T) {
	db := setupPaymentTestDB(t)
	r := NewPaymentRepository(db)

	pay := &modelsOrder.Payment{ID: "pay-2", OrderID: "ord-1", Amount: 50, Status: "pending"}
	err := r.CreateWithBalanceCheck(context.Background(), 100, pay)
	if err == nil {
		t.Fatal("expected error when payment exceeds remaining balance")
	}
	if !strings.Contains(err.Error(), "exceeds remaining balance") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateWithBalanceCheck_RejectsOverpaymentWhenFullyPaid(t *testing.T) {
	db := setupPaymentTestDB(t)
	r := NewPaymentRepository(db)

	// Bring the order to fully paid: confirmed 60 + 40 = order total 100,
	// leaving remaining == 0 — exactly the double-submit / overpayment case.
	if err := db.Exec(`INSERT INTO payments (id, order_id, amount, status) VALUES ('pay-1b', 'ord-1', 40, 'confirmed')`).Error; err != nil {
		t.Fatalf("seed second confirmed payment: %v", err)
	}

	pay := &modelsOrder.Payment{ID: "pay-2", OrderID: "ord-1", Amount: 1, Status: "pending"}
	err := r.CreateWithBalanceCheck(context.Background(), 100, pay)
	if err == nil {
		t.Fatal("expected error when payment exceeds remaining balance of 0")
	}
	if !strings.Contains(err.Error(), "exceeds remaining balance") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreateWithBalanceCheck_AllowsZeroPaymentWhenFullyPaid(t *testing.T) {
	db := setupPaymentTestDB(t)
	r := NewPaymentRepository(db)

	// Fully paid order (remaining == 0): a zero-amount payment is not an
	// overpayment and must still be allowed.
	if err := db.Exec(`INSERT INTO payments (id, order_id, amount, status) VALUES ('pay-1b', 'ord-1', 40, 'confirmed')`).Error; err != nil {
		t.Fatalf("seed second confirmed payment: %v", err)
	}

	pay := &modelsOrder.Payment{ID: "pay-2", OrderID: "ord-1", Amount: 0, Status: "pending"}
	if err := r.CreateWithBalanceCheck(context.Background(), 100, pay); err != nil {
		t.Fatalf("expected zero payment on fully-paid order to succeed, got %v", err)
	}
}

func TestCreateWithBalanceCheck_MissingOrderReturnsNotFound(t *testing.T) {
	db := setupPaymentTestDB(t)
	r := NewPaymentRepository(db)

	pay := &modelsOrder.Payment{ID: "pay-2", OrderID: "ord-missing", Amount: 1, Status: "pending"}
	if err := r.CreateWithBalanceCheck(context.Background(), 100, pay); err != gorm.ErrRecordNotFound {
		t.Fatalf("expected ErrRecordNotFound for missing order, got %v", err)
	}
}
