package order

import (
	"context"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	countrypkg "candypro/api/internal/pkg/country"
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
	cost, currency, _, err := s.CalculateShippingCostWithFlag(ctx, destinationCountry, estimatedWeightKg)
	return cost, currency, err
}

// CalculateShippingCostWithFlag returns the configured flag alongside cost so callers
// can distinguish "no rate configured for destination" from "rate is 0" (M-11). The
// previous CalculateShippingCost API silently collapsed both cases to 0.
func (s *ShippingService) CalculateShippingCostWithFlag(ctx context.Context, destinationCountry string, estimatedWeightKg float64) (float64, string, bool, error) {
	destinationCountry = countrypkg.NormalizeCountryCode(destinationCountry)
	rate, err := s.repo.FindBestRate(ctx, destinationCountry, estimatedWeightKg)
	if err != nil || rate == nil {
		return 0, "", false, nil
	}
	return rate.BaseCost + (rate.CostPerKg * estimatedWeightKg), rate.Currency, true, nil
}

// GetDestinationRates returns all active rates for a destination.
func (s *ShippingService) GetDestinationRates(ctx context.Context, destination string) ([]modelsOrder.ShippingRate, error) {
	return s.repo.FindByDestination(ctx, destination)
}

// CalculateShippingCostWithIncoterms applies incoterms-based freight adjustment (EXW=0, others use rate table).
func (s *ShippingService) CalculateShippingCostWithIncoterms(ctx context.Context, destinationCountry string, estimatedWeightKg float64, incoterms string) (float64, string, error) {
	inc := strings.ToUpper(strings.TrimSpace(incoterms))
	if inc == "EXW" {
		return 0, "USD", nil
	}
	cost, currency, err := s.CalculateShippingCost(ctx, destinationCountry, estimatedWeightKg)
	if err != nil {
		return 0, "", err
	}
	switch inc {
	case "FOB", "FCA":
		return cost * 0.85, currency, nil
	case "CIF", "CFR":
		return cost * 1.15, currency, nil
	case "DDP":
		return cost * 1.25, currency, nil
	default:
		return cost, currency, nil
	}
}
