package order

import (
	"context"
	"errors"
	"testing"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

// stubShippingRepo lets CalculateShippingCost* exercise the service layer
// without a real DB.
type stubShippingRepo struct {
	bestRate func(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error)
}

func (s *stubShippingRepo) FindAll(ctx context.Context) ([]modelsOrder.ShippingRate, error) {
	return nil, nil
}
func (s *stubShippingRepo) FindByID(ctx context.Context, id string) (*modelsOrder.ShippingRate, error) {
	return nil, nil
}
func (s *stubShippingRepo) Create(ctx context.Context, rate *modelsOrder.ShippingRate) error {
	return nil
}
func (s *stubShippingRepo) Update(ctx context.Context, rate *modelsOrder.ShippingRate) error {
	return nil
}
func (s *stubShippingRepo) Delete(ctx context.Context, id string) error {
	return nil
}
func (s *stubShippingRepo) FindByDestination(ctx context.Context, destination string) ([]modelsOrder.ShippingRate, error) {
	return nil, nil
}
func (s *stubShippingRepo) FindBestRate(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
	return s.bestRate(ctx, destination, weightKg)
}

// TestCalculateShipping_DBErrorFailsClosed pins the G24c fix: a real
// rate-lookup DB error must be surfaced, not collapsed into "no shipping
// configured" (which previously produced 0 shipping cost and under-collection).
func TestCalculateShipping_DBErrorFailsClosed(t *testing.T) {
	svc := NewShippingService(&stubShippingRepo{bestRate: func(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
		return nil, errors.New("connection reset")
	}})

	cost, currency, configured, err := svc.CalculateShippingCostWithFlag(context.Background(), "US", 10)
	if err == nil {
		t.Fatal("expected the DB error to be surfaced, got nil")
	}
	if configured {
		t.Fatal("expected configured=false on DB error")
	}
	if cost != 0 || currency != "" {
		t.Fatalf("expected zeroed result on DB error, got cost=%v currency=%q", cost, currency)
	}
}

// TestCalculateShipping_NoRateIsNotAnError pins that an uncovered destination
// remains the "not configured" flag with a nil error (M-11).
func TestCalculateShipping_NoRateIsNotAnError(t *testing.T) {
	svc := NewShippingService(&stubShippingRepo{bestRate: func(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
		return nil, gorm.ErrRecordNotFound
	}})

	cost, currency, configured, err := svc.CalculateShippingCostWithFlag(context.Background(), "XX", 10)
	if err != nil {
		t.Fatalf("no-rate must not be an error, got %v", err)
	}
	if configured {
		t.Fatal("expected configured=false when no rate matches")
	}
	if cost != 0 || currency != "" {
		t.Fatalf("expected zeroed result, got cost=%v currency=%q", cost, currency)
	}
}

// TestCalculateShippingCost_WrapperPropagatesDBError pins that the 4-arg
// convenience wrapper (the API the checkout confirm handler actually calls)
// surfaces a real rate-lookup DB error instead of collapsing it to a nil-error
// zero (G24c).
func TestCalculateShippingCost_WrapperPropagatesDBError(t *testing.T) {
	svc := NewShippingService(&stubShippingRepo{bestRate: func(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
		return nil, errors.New("connection reset")
	}})

	cost, currency, err := svc.CalculateShippingCost(context.Background(), "US", 10)
	if err == nil {
		t.Fatal("expected the 4-arg wrapper to surface the DB error, got nil")
	}
	if cost != 0 || currency != "" {
		t.Fatalf("expected zeroed result on DB error, got cost=%v currency=%q", cost, currency)
	}
}

// TestCalculateShippingCostWithIncoterms_PropagatesDBError pins that the
// incoterms-adjusted path (used by the checkout estimate AND confirm callers)
// surfaces a real rate-lookup DB error and does not fall back to a nil-error
// zero for a non-EXW incoterm.
func TestCalculateShippingCostWithIncoterms_PropagatesDBError(t *testing.T) {
	svc := NewShippingService(&stubShippingRepo{bestRate: func(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
		return nil, errors.New("connection reset")
	}})

	cost, currency, err := svc.CalculateShippingCostWithIncoterms(context.Background(), "US", 10, "FOB")
	if err == nil {
		t.Fatal("expected the incoterms path to surface the DB error, got nil")
	}
	if cost != 0 || currency != "" {
		t.Fatalf("expected zeroed result on DB error, got cost=%v currency=%q", cost, currency)
	}
}

// TestCalculateShipping_RateFound pins the happy path still applies the matched
// rate (base cost + per-kg component).
func TestCalculateShipping_RateFound(t *testing.T) {
	svc := NewShippingService(&stubShippingRepo{bestRate: func(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
		return &modelsOrder.ShippingRate{BaseCost: 20, CostPerKg: 1.5, Currency: "USD"}, nil
	}})

	cost, currency, configured, err := svc.CalculateShippingCostWithFlag(context.Background(), "US", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !configured {
		t.Fatal("expected configured=true")
	}
	if currency != "USD" {
		t.Fatalf("currency = %q, want USD", currency)
	}
	if cost != 35 { // 20 + 1.5*10
		t.Fatalf("cost = %v, want 35", cost)
	}
}
