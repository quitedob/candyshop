package order

import (
	"context"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupNegotiationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsOrder.NegotiationOffer{}); err != nil {
		t.Fatalf("migrate negotiation_offers: %v", err)
	}
	return db
}

func seedNegotiationOffers(t *testing.T, db *gorm.DB) {
	t.Helper()
	now := time.Now()
	offers := []modelsOrder.NegotiationOffer{
		{ID: "offer-a", InquiryID: "inq-1", UserID: "admin-1", SenderType: "admin", Status: "pending", TotalAmount: 100, Currency: "USD", CreatedAt: now, UpdatedAt: now},
		{ID: "offer-b", InquiryID: "inq-2", UserID: "admin-1", SenderType: "admin", Status: "pending", TotalAmount: 200, Currency: "USD", CreatedAt: now, UpdatedAt: now},
	}
	for _, o := range offers {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("seed offer: %v", err)
		}
	}
}

// TestTransitionStatus_ScopedByInquiry H-3: TransitionStatus 必须同时按 id 和
// inquiry_id 限定，调用方在询盘 X 上即使知道询盘 Y 的 offer ID 也无法改它的状态。
func TestTransitionStatus_ScopedByInquiry(t *testing.T) {
	db := setupNegotiationTestDB(t)
	seedNegotiationOffers(t, db)
	repo := NewNegotiationRepository(db)
	now := time.Now()

	// Acting on inq-2 must not touch offer-a (which belongs to inq-1).
	rows, err := repo.TransitionStatus(context.Background(), "offer-a", "inq-2", "pending", "accepted", now)
	if err != nil {
		t.Fatalf("TransitionStatus: %v", err)
	}
	if rows != 0 {
		t.Fatalf("expected 0 rows for cross-inquiry transition, got %d", rows)
	}

	// Acting on the correct inquiry succeeds.
	rows, err = repo.TransitionStatus(context.Background(), "offer-a", "inq-1", "pending", "accepted", now)
	if err != nil {
		t.Fatalf("TransitionStatus: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected 1 row, got %d", rows)
	}

	var status string
	if err := db.Raw(`SELECT status FROM negotiation_offers WHERE id = 'offer-a'`).Scan(&status).Error; err != nil {
		t.Fatal(err)
	}
	if status != "accepted" {
		t.Fatalf("expected offer-a accepted, got %q", status)
	}

	// offer-b on the other inquiry is untouched.
	var statusB string
	if err := db.Raw(`SELECT status FROM negotiation_offers WHERE id = 'offer-b'`).Scan(&statusB).Error; err != nil {
		t.Fatal(err)
	}
	if statusB != "pending" {
		t.Fatalf("expected offer-b still pending, got %q", statusB)
	}
}

// TestTransitionStatus_ScopedByStatus A-4: 状态条件（FROM 状态不匹配）时不得写入。
func TestTransitionStatus_ScopedByStatus(t *testing.T) {
	db := setupNegotiationTestDB(t)
	seedNegotiationOffers(t, db)
	repo := NewNegotiationRepository(db)
	now := time.Now()

	rows, err := repo.TransitionStatus(context.Background(), "offer-b", "inq-2", "accepted", "rejected", now)
	if err != nil {
		t.Fatalf("TransitionStatus: %v", err)
	}
	if rows != 0 {
		t.Fatalf("expected 0 rows for status mismatch, got %d", rows)
	}

	var status string
	if err := db.Raw(`SELECT status FROM negotiation_offers WHERE id = 'offer-b'`).Scan(&status).Error; err != nil {
		t.Fatal(err)
	}
	if status != "pending" {
		t.Fatalf("expected offer-b still pending, got %q", status)
	}
}
