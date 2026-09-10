package trade

import (
	"context"
	"errors"
	"testing"

	modelsTrade "candypro/api/internal/models/trade"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTradeStateGuardTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&modelsTrade.TradeTransaction{},
		&modelsTrade.TradeDocument{},
		&modelsTrade.ShipmentTracking{},
		&modelsTrade.SettlementRecord{},
	); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

// TestTradeUpdateTransactionStatus_GuardsStaleRead proves the guarded conditional
// UPDATE blocks a second transition keyed on an outdated status.
func TestTradeUpdateTransactionStatus_GuardsStaleRead(t *testing.T) {
	db := setupTradeStateGuardTestDB(t)
	r := NewTradeRepository(db)

	trans := &modelsTrade.TradeTransaction{
		UserID:    "u1",
		Reference: "TRD-1",
		Status:    modelsTrade.TradeStatusDraft,
	}
	if err := r.CreateTransaction(context.Background(), trans); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := r.UpdateTransactionStatus(context.Background(), trans.ID, modelsTrade.TradeStatusDraft, modelsTrade.TradeStatusPending); err != nil {
		t.Fatalf("first transition should succeed, got %v", err)
	}
	err := r.UpdateTransactionStatus(context.Background(), trans.ID, modelsTrade.TradeStatusDraft, modelsTrade.TradeStatusPending)
	if !errors.Is(err, ErrTradeStateMismatch) {
		t.Fatalf("expected ErrTradeStateMismatch, got %v", err)
	}
}

// TestTradeUpdateDocumentStatus_GuardsStaleRead proves the guarded document-status
// write rejects a transition keyed on an outdated status.
func TestTradeUpdateDocumentStatus_GuardsStaleRead(t *testing.T) {
	db := setupTradeStateGuardTestDB(t)
	r := NewTradeRepository(db)

	doc := &modelsTrade.TradeDocument{
		TransactionID: 1,
		Type:          modelsTrade.DocTypeProformaInvoice,
		DocNumber:     "PI-1",
		Status:        modelsTrade.TradeDocumentStatusDraft,
	}
	if err := r.CreateDocument(context.Background(), doc); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := r.UpdateDocumentStatus(context.Background(), doc.ID, modelsTrade.TradeDocumentStatusDraft, modelsTrade.TradeDocumentStatusSent); err != nil {
		t.Fatalf("first transition should succeed, got %v", err)
	}
	err := r.UpdateDocumentStatus(context.Background(), doc.ID, modelsTrade.TradeDocumentStatusDraft, modelsTrade.TradeDocumentStatusSent)
	if !errors.Is(err, ErrTradeDocumentStateMismatch) {
		t.Fatalf("expected ErrTradeDocumentStateMismatch, got %v", err)
	}
}

// TestShipmentUpdateStatus_GuardsStaleRead proves the guarded shipment-status write
// rejects a transition keyed on an outdated status.
func TestShipmentUpdateStatus_GuardsStaleRead(t *testing.T) {
	db := setupTradeStateGuardTestDB(t)
	r := NewShipmentRepository(db)

	shipment := &modelsTrade.ShipmentTracking{TransactionID: 1, Status: "PENDING"}
	if err := r.Create(context.Background(), shipment); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := r.UpdateStatus(context.Background(), shipment.ID, "PENDING", "IN_TRANSIT", nil); err != nil {
		t.Fatalf("first transition should succeed, got %v", err)
	}
	err := r.UpdateStatus(context.Background(), shipment.ID, "PENDING", "IN_TRANSIT", nil)
	if !errors.Is(err, ErrShipmentStateMismatch) {
		t.Fatalf("expected ErrShipmentStateMismatch, got %v", err)
	}
}

// TestSettlementUpdateStatus_GuardsStaleRead proves the guarded settlement-status
// write rejects a transition keyed on an outdated status.
func TestSettlementUpdateStatus_GuardsStaleRead(t *testing.T) {
	db := setupTradeStateGuardTestDB(t)
	r := NewTradeDocumentDetailRepository(db)

	rec := &modelsTrade.SettlementRecord{TransactionID: 1, Status: modelsTrade.SettlementStatusUnpaid}
	if err := r.CreateSettlement(context.Background(), rec); err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := r.UpdateSettlementStatus(context.Background(), rec.ID, modelsTrade.SettlementStatusUnpaid, modelsTrade.SettlementStatusPaid, nil); err != nil {
		t.Fatalf("first transition should succeed, got %v", err)
	}
	err := r.UpdateSettlementStatus(context.Background(), rec.ID, modelsTrade.SettlementStatusUnpaid, modelsTrade.SettlementStatusPaid, nil)
	if !errors.Is(err, ErrSettlementStateMismatch) {
		t.Fatalf("expected ErrSettlementStateMismatch, got %v", err)
	}
}
