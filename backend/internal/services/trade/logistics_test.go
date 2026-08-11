package trade

import (
	"context"
	"errors"
	"testing"
	"time"

	modelsTrade "candypro/api/internal/models/trade"
	repoTrade "candypro/api/internal/repository/trade"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupLogisticsTestDB(t *testing.T) *gorm.DB {
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

// newLogisticsServiceForTest wires real repos over an in-memory sqlite DB. The
// order/trade repos are unused by the exercised methods, so nil is acceptable.
func newLogisticsServiceForTest(db *gorm.DB) *LogisticsService {
	return NewLogisticsService(
		repoTrade.NewShipmentRepository(db),
		repoTrade.NewShipmentEventRepository(db),
		nil,
		nil,
		db,
	)
}

// TestLogisticsService_GetTransactionShipmentTimeline_enforcesOwnership verifies
// the customer-portal IDOR guard: a shipment's timeline is only readable when the
// caller supplies the trade transaction the shipment belongs to (H4/MEDIUM-3).
func TestLogisticsService_GetTransactionShipmentTimeline_enforcesOwnership(t *testing.T) {
	db := setupLogisticsTestDB(t)
	s := newLogisticsServiceForTest(db)

	shipment := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := repoTrade.NewShipmentRepository(db).Create(context.Background(), shipment); err != nil {
		t.Fatalf("create shipment: %v", err)
	}
	now := time.Now()
	event := &modelsTrade.ShipmentEvent{
		ShipmentID:  shipment.ID,
		EventType:   modelsTrade.EventDispatched,
		Description: "dispatched",
		EventTime:   now,
		CreatedAt:   now,
	}
	if err := repoTrade.NewShipmentEventRepository(db).Create(context.Background(), event); err != nil {
		t.Fatalf("create event: %v", err)
	}

	// Owning transaction: events are returned.
	events, err := s.GetTransactionShipmentTimeline(context.Background(), 1, shipment.ID)
	if err != nil {
		t.Fatalf("timeline for owning transaction: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	// A different (foreign) transaction: access must be rejected.
	if _, err := s.GetTransactionShipmentTimeline(context.Background(), 2, shipment.ID); !errors.Is(err, ErrShipmentNotInTransaction) {
		t.Fatalf("expected ErrShipmentNotInTransaction, got %v", err)
	}
}
