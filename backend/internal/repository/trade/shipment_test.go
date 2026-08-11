package trade

import (
	"context"
	"testing"
	"time"

	modelsTrade "candypro/api/internal/models/trade"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupShipmentTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsTrade.ShipmentTracking{}, &modelsTrade.ShipmentEvent{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// TestShipmentRepository_DispatchOnlyWinsOnce verifies the optimistic-lock guard
// on the PENDING -> DISPATCHED transition: a second, concurrent dispatch of the
// same shipment must lose (RowsAffected == 0) instead of unconditionally
// overwriting the row (H4/MEDIUM-4).
func TestShipmentRepository_DispatchOnlyWinsOnce(t *testing.T) {
	db := setupShipmentTestDB(t)
	r := NewShipmentRepository(db)

	shipment := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := r.Create(context.Background(), shipment); err != nil {
		t.Fatalf("create: %v", err)
	}

	// First dispatch wins the transition.
	won, err := r.Dispatch(context.Background(), db, shipment.ID, time.Now())
	if err != nil {
		t.Fatalf("first dispatch: %v", err)
	}
	if !won {
		t.Fatal("expected first dispatch to win the PENDING -> DISPATCHED transition")
	}

	// A second, concurrent dispatch of the same shipment must lose.
	won, err = r.Dispatch(context.Background(), db, shipment.ID, time.Now())
	if err != nil {
		t.Fatalf("second dispatch: %v", err)
	}
	if won {
		t.Fatal("expected second dispatch to lose the transition (optimistic lock)")
	}

	got, err := r.FindByID(context.Background(), shipment.ID)
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	if got.Status != "DISPATCHED" {
		t.Fatalf("expected status DISPATCHED, got %q", got.Status)
	}
}

// TestShipmentRepository_DispatchUnknownID verifies that dispatching a missing
// shipment reports the transition as lost rather than creating phantom rows.
func TestShipmentRepository_DispatchUnknownID(t *testing.T) {
	db := setupShipmentTestDB(t)
	r := NewShipmentRepository(db)

	won, err := r.Dispatch(context.Background(), db, 999, time.Now())
	if err != nil {
		t.Fatalf("dispatch unknown: %v", err)
	}
	if won {
		t.Fatal("expected dispatch of a missing shipment to lose the transition")
	}
}
