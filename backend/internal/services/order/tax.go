package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	countrypkg "candypro/api/internal/pkg/country"
	"candypro/api/internal/pkg/crypto"

	"gorm.io/gorm"
)

type taxRepository interface {
	FindAll(ctx context.Context) ([]modelsOrder.TaxRate, error)
	FindByID(ctx context.Context, id string) (*modelsOrder.TaxRate, error)
	Create(ctx context.Context, rate *modelsOrder.TaxRate) error
	Update(ctx context.Context, rate *modelsOrder.TaxRate) error
	Delete(ctx context.Context, id string) error
	FindBestRate(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error)
}

type TaxService struct {
	repo taxRepository
}

func NewTaxService(repo taxRepository) *TaxService {
	return &TaxService{repo: repo}
}

func (s *TaxService) GetAllRates(ctx context.Context) ([]modelsOrder.TaxRate, error) {
	return s.repo.FindAll(ctx)
}

func (s *TaxService) GetRate(ctx context.Context, id string) (*modelsOrder.TaxRate, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TaxService) CreateRate(ctx context.Context, rate *modelsOrder.TaxRate) error {
	rate.ID = crypto.GenerateID()
	rate.CreatedAt = time.Now()
	rate.UpdatedAt = time.Now()
	return s.repo.Create(ctx, rate)
}

func (s *TaxService) UpdateRate(ctx context.Context, rate *modelsOrder.TaxRate) error {
	rate.UpdatedAt = time.Now()
	return s.repo.Update(ctx, rate)
}

func (s *TaxService) DeleteRate(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CalculateTax calculates tax amount for a given subtotal, country, and optional region.
// Returns the tax amount, the applied rate's name and rate value, and a configured flag
// that is true when a tax rate was found. When no rate matches the function returns
// (0, "", 0, nil, false) — callers can use the flag to distinguish "0% configured rate"
// from "tax not configured for destination" so checkout can surface a clearer warning
// instead of silently treating uncovered destinations as tax-exempt (M-10).
func (s *TaxService) CalculateTax(ctx context.Context, subtotal float64, country, region string) (float64, string, float64, error) {
	amount, name, rate, _, err := s.CalculateTaxWithFlag(ctx, subtotal, country, region)
	return amount, name, rate, err
}

// CalculateTaxWithFlag is the configured-aware variant of CalculateTax (M-10).
func (s *TaxService) CalculateTaxWithFlag(ctx context.Context, subtotal float64, country, region string) (float64, string, float64, bool, error) {
	country = countrypkg.NormalizeCountryCode(country)
	rate, err := s.repo.FindBestRate(ctx, country, region)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No rate configured for the destination — legitimately uncovered,
			// not an error; callers surface the "not configured" flag (M-10).
			return 0, "", 0, false, nil
		}
		// A real DB failure must not be misread as "0% tax for this
		// destination": fail closed so checkout surfaces the lookup failure
		// instead of silently under-charging (G24c).
		return 0, "", 0, false, fmt.Errorf("tax: rate lookup failed for %s: %w", country, err)
	}
	if rate == nil {
		return 0, "", 0, false, nil
	}
	return subtotal * rate.Rate, rate.Name, rate.Rate, true, nil
}
