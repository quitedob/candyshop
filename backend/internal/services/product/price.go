package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type priceRepository interface {
	FindAllPriceLists(ctx context.Context, page, limit int) ([]modelsProduct.PriceList, int64, error)
	FindPriceListByID(ctx context.Context, id string) (*modelsProduct.PriceList, error)
	CreatePriceList(ctx context.Context, list *modelsProduct.PriceList) error
	UpdatePriceList(ctx context.Context, list *modelsProduct.PriceList) error
	DeletePriceList(ctx context.Context, id string) error
	FindPriceRulesByProduct(ctx context.Context, productID string) ([]modelsProduct.PriceRule, error)
	FindPriceRulesByPriceListID(ctx context.Context, priceListID string) ([]modelsProduct.PriceRule, error)
	FindPriceRule(ctx context.Context, id string) (*modelsProduct.PriceRule, error)
	CreatePriceRule(ctx context.Context, rule *modelsProduct.PriceRule) error
	UpdatePriceRule(ctx context.Context, rule *modelsProduct.PriceRule) error
	DeletePriceRule(ctx context.Context, id string) error
	FindBestPrice(ctx context.Context, productID, priceListID string, quantity int) (*modelsProduct.PriceRule, error)
}

// PriceService handles pricing business logic.
type PriceService struct {
	repo priceRepository
}

// NewPriceService creates a new PriceService.
func NewPriceService(repo priceRepository) *PriceService {
	return &PriceService{repo: repo}
}

// GetPriceLists returns paginated price lists.
func (s *PriceService) GetPriceLists(ctx context.Context, page, limit int) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	lists, total, err := s.repo.FindAllPriceLists(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: lists,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetPriceList returns a price list by ID.
func (s *PriceService) GetPriceList(ctx context.Context, id string) (*modelsProduct.PriceList, error) {
	return s.repo.FindPriceListByID(ctx, id)
}

// CreatePriceList creates a new price list.
func (s *PriceService) CreatePriceList(ctx context.Context, list *modelsProduct.PriceList) error {
	if list.Status == "" {
		list.Status = "active"
	}
	return s.repo.CreatePriceList(ctx, list)
}

// UpdatePriceList updates a price list.
func (s *PriceService) UpdatePriceList(ctx context.Context, list *modelsProduct.PriceList) error {
	return s.repo.UpdatePriceList(ctx, list)
}

// DeletePriceList deletes a price list.
func (s *PriceService) DeletePriceList(ctx context.Context, id string) error {
	return s.repo.DeletePriceList(ctx, id)
}

// GetProductPrices returns all price rules for a product.
func (s *PriceService) GetProductPrices(ctx context.Context, productID string) ([]modelsProduct.PriceRule, error) {
	return s.repo.FindPriceRulesByProduct(ctx, productID)
}

// GetPriceListRules returns all price rules belonging to a price list.
func (s *PriceService) GetPriceListRules(ctx context.Context, priceListID string) ([]modelsProduct.PriceRule, error) {
	return s.repo.FindPriceRulesByPriceListID(ctx, priceListID)
}

// SetProductPrice creates or updates a price rule.
func (s *PriceService) SetProductPrice(ctx context.Context, rule *modelsProduct.PriceRule) error {
	// Check if a rule already exists for this product/priceList/minQuantity combo
	existing, err := s.repo.FindBestPrice(ctx, rule.ProductID, rule.PriceListID, rule.MinQuantity)
	if err == nil && existing != nil && existing.MinQuantity == rule.MinQuantity {
		existing.UnitPrice = rule.UnitPrice
		existing.Currency = rule.Currency
		return s.repo.UpdatePriceRule(ctx, existing)
	}
	return s.repo.CreatePriceRule(ctx, rule)
}

// GetPriceForProduct returns the applicable unit price for a product given a price list and quantity.
func (s *PriceService) GetPriceForProduct(ctx context.Context, productID, priceListID string, quantity int) (float64, error) {
	rule, err := s.repo.FindBestPrice(ctx, productID, priceListID, quantity)
	if err != nil {
		return 0, err
	}
	return rule.UnitPrice, nil
}

// MinQuantityForPriceList returns the lowest tier minimum quantity for a product on a price list.
func (s *PriceService) MinQuantityForPriceList(ctx context.Context, productID, priceListID string) (int, error) {
	rules, err := s.repo.FindPriceRulesByProduct(ctx, productID)
	if err != nil {
		return 1, err
	}
	minQty := 0
	for _, r := range rules {
		if r.PriceListID != priceListID {
			continue
		}
		if minQty == 0 || r.MinQuantity < minQty {
			minQty = r.MinQuantity
		}
	}
	if minQty <= 0 {
		return 1, nil
	}
	return minQty, nil
}

// DeleteProductPrice deletes a specific price rule by ID.
func (s *PriceService) DeleteProductPrice(ctx context.Context, id string) error {
	return s.repo.DeletePriceRule(ctx, id)
}
