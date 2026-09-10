package trade

import (
	"context"
	"testing"

	modelsTrade "candypro/api/internal/models/trade"
	tradeRepo "candypro/api/internal/repository/trade"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestUpdateTransactionStatus_LegacyStatusCase(t *testing.T) {
	for _, storedStatus := range []string{"draft", " draft ", modelsTrade.TradeStatusDraft} {
		t.Run(storedStatus, func(t *testing.T) {
			database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
			if err != nil {
				t.Fatal(err)
			}
			if err := database.AutoMigrate(&modelsTrade.TradeTransaction{}, &modelsTrade.TradeDocument{}, &modelsTrade.ComplianceRequirement{}); err != nil {
				t.Fatal(err)
			}
			repository := tradeRepo.NewTradeRepository(database)
			transaction := modelsTrade.TradeTransaction{UserID: "trade-status-owner", Reference: "trade-status-reference", Status: storedStatus}
			if err := repository.CreateTransaction(context.Background(), &transaction); err != nil {
				t.Fatal(err)
			}
			service := NewTradeService(repository)
			if err := service.UpdateTransactionStatus(context.Background(), transaction.ID, modelsTrade.TradeStatusPending); err != nil {
				t.Fatalf("legacy transition: %v", err)
			}
			var persisted modelsTrade.TradeTransaction
			if err := database.First(&persisted, transaction.ID).Error; err != nil {
				t.Fatal(err)
			}
			if persisted.Status != modelsTrade.TradeStatusPending {
				t.Fatalf("status=%q", persisted.Status)
			}
			if err := repository.UpdateTransactionStatus(context.Background(), transaction.ID, storedStatus, modelsTrade.TradeStatusPending); err == nil {
				t.Fatal("stale legacy status must still lose the compare-and-set")
			}
		})
	}
}
