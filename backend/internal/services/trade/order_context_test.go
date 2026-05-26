package trade

import (
	"encoding/json"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
)

func TestBuildOrderContextJSON_includesLineItems(t *testing.T) {
	order := &modelsOrder.Order{
		ID:          "ord-1",
		OrderNumber: "ORD-001",
		Status:      "confirmed",
		Currency:    "USD",
		TotalAmount: 1200,
		Items: modelsOrder.OrderItemArray{
			{ProductID: "p1", Quantity: 100, UnitPrice: 12, Specifications: "FOB"},
		},
	}
	raw := BuildOrderContextJSON(order)
	if raw == "" {
		t.Fatal("expected non-empty order context")
	}
	var parsed map[string]any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	items, ok := parsed["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected 1 line item, got %#v", parsed["items"])
	}
}

func TestBatchTranslateTexts_empty(t *testing.T) {
	s := &AIService{Client: nil}
	out, warnings, err := s.BatchTranslateTexts(t.Context(), map[string]string{}, "en", "zh")
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 || len(warnings) != 0 {
		t.Fatalf("expected empty result, got out=%v warnings=%v", out, warnings)
	}
}
