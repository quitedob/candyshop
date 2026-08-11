package order

import (
	"testing"

	modelsOrder "candypro/api/internal/models/order"
)

func TestOrderFinancialLineageHash_NilOrder(t *testing.T) {
	if got := OrderFinancialLineageHash(nil); got != "" {
		t.Fatalf("OrderFinancialLineageHash(nil) = %q, want empty", got)
	}
}

// TestOrderFinancialLineageHash_DistinctNearbyValues is the G26 contract:
// orders whose cent-level financials differ by a penny must never produce the
// same lineage hash, otherwise invoice deduplication would treat them as equal.
func TestOrderFinancialLineageHash_DistinctNearbyValues(t *testing.T) {
	a := &modelsOrder.Order{
		ID: "order-g26a", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{},
	}
	b := &modelsOrder.Order{
		ID: "order-g26b", Currency: "USD",
		Subtotal: 20.00, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{},
	}
	if OrderFinancialLineageHash(a) == OrderFinancialLineageHash(b) {
		t.Fatal("lineage hashes collided for 19.99 vs 20.00")
	}
}

// TestOrderFinancialLineageHash_CentStableAgainstFloatNoise is the G26 fix:
// the same nominal cent value recomputed with sub-cent float noise must hash
// identically. The previous fmt.Sprintf("%f", ...) kept 6 decimals of float
// noise, so 19.99 and 19.9900006 produced different hashes for the same
// cent value, falsely flagging derived invoices as stale (wrong dedup).
func TestOrderFinancialLineageHash_CentStableAgainstFloatNoise(t *testing.T) {
	a := &modelsOrder.Order{
		ID: "order-g26c", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{},
	}
	b := &modelsOrder.Order{
		ID: "order-g26c", Currency: "USD",
		Subtotal: 19.9900006, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{},
	}
	if OrderFinancialLineageHash(a) != OrderFinancialLineageHash(b) {
		t.Fatal("lineage hash changed for the same cent value under sub-cent float noise")
	}
}

// TestOrderFinancialLineageHash_DistinctLineItems ensures line items still
// participate in the hash (empty vs non-empty items must differ).
func TestOrderFinancialLineageHash_DistinctLineItems(t *testing.T) {
	a := &modelsOrder.Order{
		ID: "order-g26d", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{},
	}
	b := &modelsOrder.Order{
		ID: "order-g26d", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{{ProductID: "sku-x", Quantity: 1, UnitPrice: 19.99}},
	}
	if OrderFinancialLineageHash(a) == OrderFinancialLineageHash(b) {
		t.Fatal("lineage hash ignored line items")
	}
}

// TestOrderFinancialLineageHash_CentStableAgainstFloatNoise_WithLineItems is
// the arguer-refutation regression: sub-cent float noise in a line item's
// UnitPrice must NOT change the hash. The prior fix normalized only the
// aggregates and still fed json.Marshal(o.Items) — where OrderItem.UnitPrice is
// float64 — into the hash, so 19.99 vs 19.9900006 in a line item produced
// different hashes for the same nominal value. This test fails on that path.
func TestOrderFinancialLineageHash_CentStableAgainstFloatNoise_WithLineItems(t *testing.T) {
	a := &modelsOrder.Order{
		ID: "order-g26e", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{
			{ProductID: "sku-x", Quantity: 1, UnitPrice: 19.99},
			{ProductID: "sku-y", Quantity: 2, UnitPrice: 1.25, FulfilledQuantity: 1, Specifications: "red"},
		},
	}
	b := &modelsOrder.Order{
		ID: "order-g26e", Currency: "USD",
		Subtotal: 19.9900006, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{
			{ProductID: "sku-x", Quantity: 1, UnitPrice: 19.9900006},
			{ProductID: "sku-y", Quantity: 2, UnitPrice: 1.2500001, FulfilledQuantity: 1, Specifications: "red"},
		},
	}
	if OrderFinancialLineageHash(a) != OrderFinancialLineageHash(b) {
		t.Fatal("lineage hash changed for the same cent value under sub-cent float noise in line items")
	}
}

// TestOrderFinancialLineageHash_DistinctUnitPriceStillDistinct guards against
// the canonical cents normalization collapsing genuinely different line-item
// prices (wrong dedup would treat a 19.99 line and a 20.00 line as equal).
func TestOrderFinancialLineageHash_DistinctUnitPriceStillDistinct(t *testing.T) {
	a := &modelsOrder.Order{
		ID: "order-g26g", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{{ProductID: "sku-x", Quantity: 1, UnitPrice: 19.99}},
	}
	b := &modelsOrder.Order{
		ID: "order-g26g", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{{ProductID: "sku-x", Quantity: 1, UnitPrice: 20.00}},
	}
	if OrderFinancialLineageHash(a) == OrderFinancialLineageHash(b) {
		t.Fatal("distinct line-item unit prices collided")
	}
}

// TestOrderFinancialLineageHash_NilVsEmptyItemsStable guards the preload-shape
// instability: json.Marshal of a nil Items slice yields "null" while an empty
// non-nil slice yields "[]", so the same order hashed differently depending on
// whether line items were preloaded. The canonical projection renders both as
// "[]". This test fails on the plain json.Marshal(o.Items) path.
func TestOrderFinancialLineageHash_NilVsEmptyItemsStable(t *testing.T) {
	a := &modelsOrder.Order{
		ID: "order-g26f", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
	}
	b := &modelsOrder.Order{
		ID: "order-g26f", Currency: "USD",
		Subtotal: 19.99, TaxAmount: 0, ShippingAmount: 0,
		Items: modelsOrder.OrderItemArray{},
	}
	if OrderFinancialLineageHash(a) != OrderFinancialLineageHash(b) {
		t.Fatal("nil Items vs empty Items produced different lineage hashes")
	}
}
