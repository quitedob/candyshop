package order

import (
	"context"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
)

type shippingRepository interface {
	FindAll(ctx context.Context) ([]modelsOrder.ShippingRate, error)
	FindByID(ctx context.Context, id string) (*modelsOrder.ShippingRate, error)
	Create(ctx context.Context, rate *modelsOrder.ShippingRate) error
	Update(ctx context.Context, rate *modelsOrder.ShippingRate) error
	Delete(ctx context.Context, id string) error
	FindByDestination(ctx context.Context, destination string) ([]modelsOrder.ShippingRate, error)
	FindBestRate(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error)
}

type ShippingService struct {
	repo shippingRepository
}

func NewShippingService(repo shippingRepository) *ShippingService {
	return &ShippingService{repo: repo}
}

func (s *ShippingService) GetAllRates(ctx context.Context) ([]modelsOrder.ShippingRate, error) {
	return s.repo.FindAll(ctx)
}

func (s *ShippingService) GetRate(ctx context.Context, id string) (*modelsOrder.ShippingRate, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ShippingService) CreateRate(ctx context.Context, rate *modelsOrder.ShippingRate) error {
	rate.ID = crypto.GenerateID()
	rate.CreatedAt = time.Now()
	rate.UpdatedAt = time.Now()
	return s.repo.Create(ctx, rate)
}

func (s *ShippingService) UpdateRate(ctx context.Context, rate *modelsOrder.ShippingRate) error {
	rate.UpdatedAt = time.Now()
	return s.repo.Update(ctx, rate)
}

func (s *ShippingService) DeleteRate(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CalculateShippingCost calculates shipping cost based on destination country and estimated weight.
// Returns the best matching rate cost or 0 if no rate matches.
func (s *ShippingService) CalculateShippingCost(ctx context.Context, destinationCountry string, estimatedWeightKg float64) (float64, string, error) {
	rate, err := s.repo.FindBestRate(ctx, destinationCountry, estimatedWeightKg)
	if err != nil {
		return 0, "", nil // no rate found, return 0 silently
	}
	totalCost := rate.BaseCost + (rate.CostPerKg * estimatedWeightKg)
	return totalCost, rate.Currency, nil
}

// GetDestinationRates returns all active rates for a destination.
func (s *ShippingService) GetDestinationRates(ctx context.Context, destination string) ([]modelsOrder.ShippingRate, error) {
	return s.repo.FindByDestination(ctx, destination)
}
