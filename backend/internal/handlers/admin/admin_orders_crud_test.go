package admin

import (
	"testing"

	modelsOrder "candypro/api/internal/models/order"
)

func TestHasNonZeroStockDelta(t *testing.T) {
	if hasNonZeroStockDelta(map[string]int{}) {
		t.Fatal("empty map should not have non-zero delta")
	}
	if hasNonZeroStockDelta(map[string]int{"p1": 0, "p2": 0}) {
		t.Fatal("all-zero map should not have non-zero delta")
	}
	if !hasNonZeroStockDelta(map[string]int{"p1": 5}) {
		t.Fatal("expected non-zero delta")
	}
	if !hasNonZeroStockDelta(map[string]int{"p1": -3}) {
		t.Fatal("expected negative delta to count")
	}
}

func TestCalculateStockAdjustment_UnchangedItems(t *testing.T) {
	items := []modelsOrder.OrderItem{
		{ProductID: "p1", Quantity: 8000},
		{ProductID: "p2", Quantity: 7000},
	}
	adj := calculateStockAdjustment(items, items)
	if hasNonZeroStockDelta(adj) {
		t.Fatalf("unchanged items should produce zero deltas, got %v", adj)
	}
}

func TestCalculateStockAdjustment_QuantityChange(t *testing.T) {
	oldItems := []modelsOrder.OrderItem{{ProductID: "p1", Quantity: 5}}
	newItems := []modelsOrder.OrderItem{{ProductID: "p1", Quantity: 8}}
	adj := calculateStockAdjustment(oldItems, newItems)
	if adj["p1"] != 3 {
		t.Fatalf("expected delta +3, got %d", adj["p1"])
	}
}
