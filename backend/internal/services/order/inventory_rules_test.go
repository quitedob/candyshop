package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"strings"
	"testing"
)

func TestValidateInventory_InsufficientStock(t *testing.T) {
	items := []modelsOrder.OrderItem{
		{ProductID: "p-1", Quantity: 120, UnitPrice: 1.0},
	}
	products := map[string]modelsProduct.Product{
		"p-1": {
			ID:            "p-1",
			MOQ:           10,
			StockQuantity: 50,
			Status:        "active",
		},
	}

	result := validateInventoryWithSellable(items, products, nil)
	if len(result.Violations) == 0 {
		t.Fatalf("expected stock violation")
	}
	joined := strings.ToLower(strings.Join(result.Violations, " | "))
	if !strings.Contains(joined, "exceeds stock") {
		t.Fatalf("expected stock overflow message, got: %v", result.Violations)
	}
}

func TestValidateInventory_LowStockWarning(t *testing.T) {
	items := []modelsOrder.OrderItem{
		{ProductID: "p-2", Quantity: 100, UnitPrice: 1.0},
	}
	products := map[string]modelsProduct.Product{
		"p-2": {
			ID:            "p-2",
			MOQ:           80,
			StockQuantity: 150,
			Status:        "active",
		},
	}

	result := validateInventoryWithSellable(items, products, nil)
	if len(result.Violations) != 0 {
		t.Fatalf("expected no violation, got: %v", result.Violations)
	}
	if len(result.Warnings) == 0 {
		t.Fatalf("expected low stock warning")
	}
}

func TestValidateInventory_BelowMOQViolation(t *testing.T) {
	items := []modelsOrder.OrderItem{
		{ProductID: "p-3", Quantity: 100, UnitPrice: 1.0},
	}
	products := map[string]modelsProduct.Product{
		"p-3": {
			ID:            "p-3",
			MOQ:           500,
			StockQuantity: 1000,
			Status:        "active",
		},
	}

	result := validateInventoryWithSellable(items, products, nil)
	if len(result.Violations) == 0 {
		t.Fatalf("expected MOQ violation")
	}
	joined := strings.ToLower(strings.Join(result.Violations, " | "))
	if !strings.Contains(joined, "below moq") {
		t.Fatalf("expected MOQ message, got: %v", result.Violations)
	}
}

func TestValidateInventoryWithSellable_ChannelCap(t *testing.T) {
	items := []modelsOrder.OrderItem{
		{ProductID: "p-4", Quantity: 50, UnitPrice: 1.0},
	}
	products := map[string]modelsProduct.Product{
		"p-4": {
			ID:            "p-4",
			MOQ:           1,
			StockQuantity: 200,
			Status:        "active",
		},
	}
	sellable := map[string]int{"p-4": 30}
	result := validateInventoryWithSellable(items, products, sellable)
	if len(result.Violations) == 0 {
		t.Fatalf("expected violation when OMS cap below order qty")
	}
}
