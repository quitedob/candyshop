package trade

import (
	"context"
	"testing"

	modelsTrade "candypro/api/internal/models/trade"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupQuotationReviewTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsTrade.QuotationReview{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestQuotationReviewRepository_CreateListGet(t *testing.T) {
	db := setupQuotationReviewTestDB(t)
	r := NewQuotationReviewRepository(db)

	if err := r.Create(context.Background(), &modelsTrade.QuotationReview{CustomerRef: "INQ-1", Currency: "USD", TotalAmount: 100}); err != nil {
		t.Fatalf("create 1: %v", err)
	}
	if err := r.Create(context.Background(), &modelsTrade.QuotationReview{CustomerRef: "INQ-2", Currency: "EUR", TotalAmount: 200}); err != nil {
		t.Fatalf("create 2: %v", err)
	}

	rows, total, err := r.List(context.Background(), "", 1, 10)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 || len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d (total %d)", len(rows), total)
	}

	pending, _, err := r.List(context.Background(), modelsTrade.QuotationReviewStatusPending, 1, 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("expected 2 pending, got %d", len(pending))
	}

	got, err := r.GetByID(context.Background(), rows[0].ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if got.CustomerRef == "" {
		t.Fatal("expected customer ref on fetched review")
	}
}

func TestQuotationReviewRepository_DecideTransition(t *testing.T) {
	db := setupQuotationReviewTestDB(t)
	r := NewQuotationReviewRepository(db)

	qr := &modelsTrade.QuotationReview{CustomerRef: "INQ-1"}
	if err := r.Create(context.Background(), qr); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := r.UpdateStatus(context.Background(), qr.ID, modelsTrade.QuotationReviewStatusApproved, "ok"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	got, err := r.GetByID(context.Background(), qr.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Status != modelsTrade.QuotationReviewStatusApproved {
		t.Fatalf("expected approved, got %q", got.Status)
	}

	// Re-deciding an already-decided row must fail (pending-only transition guard).
	if err := r.UpdateStatus(context.Background(), qr.ID, modelsTrade.QuotationReviewStatusRejected, ""); err == nil {
		t.Fatal("expected error re-deciding a non-pending review")
	}
}

func TestQuotationReviewRepository_UpdateStatusSetsUpdatedAt(t *testing.T) {
	db := setupQuotationReviewTestDB(t)
	r := NewQuotationReviewRepository(db)

	qr := &modelsTrade.QuotationReview{CustomerRef: "INQ-TS"}
	if err := r.Create(context.Background(), qr); err != nil {
		t.Fatalf("create: %v", err)
	}
	if qr.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at set on create")
	}
	before := qr.UpdatedAt

	if err := r.UpdateStatus(context.Background(), qr.ID, modelsTrade.QuotationReviewStatusApproved, "ok"); err != nil {
		t.Fatalf("approve: %v", err)
	}
	got, err := r.GetByID(context.Background(), qr.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.UpdatedAt.IsZero() {
		t.Fatal("expected updated_at set after decision")
	}
	if got.UpdatedAt.Before(before) {
		t.Fatalf("expected updated_at to advance, before=%v after=%v", before, got.UpdatedAt)
	}
}
