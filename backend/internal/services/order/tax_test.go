package order

import (
	"context"
	"errors"
	"testing"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

// stubTaxRepo lets CalculateTax* exercise the service layer without a real DB.
type stubTaxRepo struct {
	bestRate func(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error)
}

func (s *stubTaxRepo) FindAll(ctx context.Context) ([]modelsOrder.TaxRate, error) { return nil, nil }
func (s *stubTaxRepo) FindByID(ctx context.Context, id string) (*modelsOrder.TaxRate, error) {
	return nil, nil
}
func (s *stubTaxRepo) Create(ctx context.Context, rate *modelsOrder.TaxRate) error { return nil }
func (s *stubTaxRepo) Update(ctx context.Context, rate *modelsOrder.TaxRate) error { return nil }
func (s *stubTaxRepo) Delete(ctx context.Context, id string) error                 { return nil }
func (s *stubTaxRepo) FindBestRate(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error) {
	return s.bestRate(ctx, country, region)
}

// TestCalculateTax_DBErrorFailsClosed pins the G24c fix: a real rate-lookup DB
// error must be surfaced, not collapsed into "no rate configured" (which
// previously produced 0% tax and under-collection).
func TestCalculateTax_DBErrorFailsClosed(t *testing.T) {
	svc := NewTaxService(&stubTaxRepo{bestRate: func(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error) {
		return nil, errors.New("connection reset")
	}})

	amount, name, rate, configured, err := svc.CalculateTaxWithFlag(context.Background(), 100, "US", "NY")
	if err == nil {
		t.Fatal("expected the DB error to be surfaced, got nil")
	}
	if configured {
		t.Fatal("expected configured=false on DB error")
	}
	if amount != 0 || name != "" || rate != 0 {
		t.Fatalf("expected zeroed result on DB error, got amount=%v name=%q rate=%v", amount, name, rate)
	}
}

// TestCalculateTax_NoRateIsNotAnError pins that an uncovered destination remains
// the "not configured" flag with a nil error (M-10) — the fail-closed change
// must not regress the ordinary no-rate path.
func TestCalculateTax_NoRateIsNotAnError(t *testing.T) {
	svc := NewTaxService(&stubTaxRepo{bestRate: func(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error) {
		return nil, gorm.ErrRecordNotFound
	}})

	amount, name, rate, configured, err := svc.CalculateTaxWithFlag(context.Background(), 100, "XX", "")
	if err != nil {
		t.Fatalf("no-rate must not be an error, got %v", err)
	}
	if configured {
		t.Fatal("expected configured=false when no rate matches")
	}
	if amount != 0 || name != "" || rate != 0 {
		t.Fatalf("expected zeroed result, got amount=%v name=%q rate=%v", amount, name, rate)
	}
}

// TestCalculateTax_WrapperPropagatesDBError pins that the 4-arg convenience
// wrapper (the API the checkout confirm handler actually calls) surfaces a real
// rate-lookup DB error instead of collapsing it to a nil-error zero. This is the
// contract the handler-side fail-closed fix relies on: the wrapper must never
// re-introduce the old fail-open behavior on the calling path (G24c).
func TestCalculateTax_WrapperPropagatesDBError(t *testing.T) {
	svc := NewTaxService(&stubTaxRepo{bestRate: func(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error) {
		return nil, errors.New("connection reset")
	}})

	amount, name, rate, err := svc.CalculateTax(context.Background(), 100, "US", "NY")
	if err == nil {
		t.Fatal("expected the 4-arg wrapper to surface the DB error, got nil")
	}
	if amount != 0 || name != "" || rate != 0 {
		t.Fatalf("expected zeroed result on DB error, got amount=%v name=%q rate=%v", amount, name, rate)
	}
}

// TestCalculateTax_RateFound pins the happy path still applies the matched rate.
func TestCalculateTax_RateFound(t *testing.T) {
	svc := NewTaxService(&stubTaxRepo{bestRate: func(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error) {
		return &modelsOrder.TaxRate{Name: "NY Sales Tax", Rate: 0.08875}, nil
	}})

	amount, name, rate, configured, err := svc.CalculateTaxWithFlag(context.Background(), 100, "US", "NY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !configured {
		t.Fatal("expected configured=true")
	}
	if rate != 0.08875 || name != "NY Sales Tax" {
		t.Fatalf("rate/name mismatch: rate=%v name=%q", rate, name)
	}
	if amount < 8.874 || amount > 8.876 {
		t.Fatalf("amount = %v, want ~8.875 (100 * 0.08875)", amount)
	}
}
