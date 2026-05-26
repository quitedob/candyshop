package order

import (
	"context"
	"math"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
)

type stubCostLookup struct {
	costs      map[string]float64
	basePrices map[string]float64
	slugRefs   map[string]float64
}

func (s stubCostLookup) ComputeWeightedAvgCost(_ context.Context, productID string) float64 {
	return s.costs[productID]
}

func (s stubCostLookup) GetCOGSReferencePrice(_ context.Context, productID string) float64 {
	if s.slugRefs != nil {
		if ref, ok := s.slugRefs[productID]; ok {
			return ref
		}
	}
	if s.basePrices == nil {
		return 0
	}
	return s.basePrices[productID]
}

func TestComputeOrderCOGS(t *testing.T) {
	items := []modelsOrder.OrderItem{
		{ProductID: "p1", Quantity: 2, UnitPrice: 10},
		{ProductID: "p2", Quantity: 3, UnitPrice: 5},
	}
	lookup := stubCostLookup{
		costs:      map[string]float64{"p1": 4.5, "p2": 2},
		basePrices: map[string]float64{"p1": 10, "p2": 5},
	}
	got := ComputeOrderCOGS(context.Background(), items, lookup)
	want := 2*4.5 + 3*2
	if got != want {
		t.Fatalf("ComputeOrderCOGS = %v, want %v", got, want)
	}
}

func TestComputeOrderCOGSContractPriceScaling(t *testing.T) {
	// 合同价 $0.99，Catalog 参考价 $8.50，加权成本 $3.40 → 有效成本 ≈ $0.396
	items := []modelsOrder.OrderItem{
		{ProductID: "gummy", Quantity: 6000, UnitPrice: 0.99},
	}
	lookup := stubCostLookup{
		costs:      map[string]float64{"gummy": 3.40},
		slugRefs:   map[string]float64{"gummy": 8.50},
	}
	got := ComputeOrderCOGS(context.Background(), items, lookup)
	want := 6000 * 0.99 * (3.40 / 8.50)
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("ComputeOrderCOGS = %v, want %v", got, want)
	}
	subtotal := 6000 * 0.99
	if got >= subtotal {
		t.Fatalf("COGS %v should be less than subtotal %v", got, subtotal)
	}
}

func TestComputeOrderCOGS_Simulation20260524ThreeProducts(t *testing.T) {
	// 2026-05-24 lifecycle simulation order: contract prices, catalog reference scaling.
	items := []modelsOrder.OrderItem{
		{ProductID: "4d-fruit-gummy", Quantity: 6000, UnitPrice: 0.99},
		{ProductID: "crystal-hard-candy", Quantity: 10000, UnitPrice: 0.99},
		{ProductID: "sour-belt", Quantity: 8000, UnitPrice: 0.12},
	}
	lookup := stubCostLookup{
		costs: map[string]float64{
			"4d-fruit-gummy":     3.40,
			"crystal-hard-candy": 2.04,
			"sour-belt":          1.68,
		},
		slugRefs: map[string]float64{
			"4d-fruit-gummy":     8.50,
			"crystal-hard-candy": 5.10,
			"sour-belt":          4.20,
		},
	}
	got := ComputeOrderCOGS(context.Background(), items, lookup)
	want := 6000*0.99*(3.40/8.50) + 10000*0.99*(2.04/5.10) + 8000*0.12*(1.68/4.20)
	if math.Abs(got-want) > 0.05 {
		t.Fatalf("ComputeOrderCOGS = %v, want %v", got, want)
	}
	subtotal := 6000*0.99 + 10000*0.99 + 8000*0.12
	if got >= subtotal {
		t.Fatalf("COGS %v should be less than subtotal %v", got, subtotal)
	}
	unscaled := 6000*3.40 + 10000*2.04 + 8000*1.68
	if math.Abs(got-unscaled) < 1 {
		t.Fatalf("COGS %v should not match unscaled cost %v", got, unscaled)
	}
}

func TestComputeOrderCOGS_ContractBasePricePollutionUsesCatalogRef(t *testing.T) {
	// DB base_price polluted to contract price; catalog ref still scales correctly.
	items := []modelsOrder.OrderItem{
		{ProductID: "gummy", Quantity: 1000, UnitPrice: 0.99},
	}
	lookup := stubCostLookup{
		costs:      map[string]float64{"gummy": 3.40},
		basePrices: map[string]float64{"gummy": 0.99},
		slugRefs:   map[string]float64{"gummy": 8.50},
	}
	got := ComputeOrderCOGS(context.Background(), items, lookup)
	want := 1000 * 0.99 * (3.40 / 8.50)
	if math.Abs(got-want) > 0.01 {
		t.Fatalf("ComputeOrderCOGS = %v, want %v", got, want)
	}
}

func TestEffectiveUnitCostFallback(t *testing.T) {
	if got := effectiveUnitCost(0, 3.4, 8.5); got != 0 {
		t.Fatalf("expected 0 when unitPrice missing, got %v", got)
	}
	if got := effectiveUnitCost(0.99, 3.4, 0); got != 0 {
		t.Fatalf("expected 0 when basePrice=0, got %v", got)
	}
	if got := effectiveUnitCost(0.99, 0, 8.5); got != 0 {
		t.Fatalf("expected 0 when weighted cost missing, got %v", got)
	}
}

func TestComputeOrderCOGSNilLookup(t *testing.T) {
	if got := ComputeOrderCOGS(context.Background(), []modelsOrder.OrderItem{{ProductID: "p1", Quantity: 1}}, nil); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
}
