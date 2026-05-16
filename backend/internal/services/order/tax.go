package order

import (
	"context"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
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
// Returns the tax amount and the applied rate name, or 0 if no rate matches.
func (s *TaxService) CalculateTax(ctx context.Context, subtotal float64, country, region string) (float64, string, float64, error) {
	rate, err := s.repo.FindBestRate(ctx, country, region)
	if err != nil {
		return 0, "", 0, nil // no rate found, return 0 silently
	}
	taxAmount := subtotal * rate.Rate
	return taxAmount, rate.Name, rate.Rate, nil
}
