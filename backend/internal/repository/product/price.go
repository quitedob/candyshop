package product

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"

	"gorm.io/gorm"
)

// PriceRepository handles price list and price rule data operations.
type PriceRepository struct {
	db *gorm.DB
}

// NewPriceRepository creates a new PriceRepository.
func NewPriceRepository(db *gorm.DB) *PriceRepository {
	return &PriceRepository{db: db}
}

// --- PriceList ---

// FindAllPriceLists returns paginated price lists.
func (r *PriceRepository) FindAllPriceLists(ctx context.Context, page, limit int) ([]modelsProduct.PriceList, int64, error) {
	var lists []modelsProduct.PriceList
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsProduct.PriceList{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&lists).Error; err != nil {
		return nil, 0, err
	}

	return lists, total, nil
}

// FindPriceListByID returns a price list by ID.
func (r *PriceRepository) FindPriceListByID(ctx context.Context, id string) (*modelsProduct.PriceList, error) {
	var list modelsProduct.PriceList
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&list).Error; err != nil {
		return nil, err
	}
	return &list, nil
}

// CreatePriceList creates a new price list.
func (r *PriceRepository) CreatePriceList(ctx context.Context, list *modelsProduct.PriceList) error {
	return r.db.WithContext(ctx).Create(list).Error
}

// UpdatePriceList updates a price list.
func (r *PriceRepository) UpdatePriceList(ctx context.Context, list *modelsProduct.PriceList) error {
	return r.db.WithContext(ctx).Save(list).Error
}

// DeletePriceList deletes a price list by ID.
func (r *PriceRepository) DeletePriceList(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.PriceList{}, "id = ?", id).Error
}

// --- PriceRule ---

// FindPriceRulesByProduct returns all price rules for a product.
func (r *PriceRepository) FindPriceRulesByProduct(ctx context.Context, productID string) ([]modelsProduct.PriceRule, error) {
	var rules []modelsProduct.PriceRule
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).Order("min_quantity ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// FindPriceRulesByPriceListID returns all price rules for a price list.
func (r *PriceRepository) FindPriceRulesByPriceListID(ctx context.Context, priceListID string) ([]modelsProduct.PriceRule, error) {
	var rules []modelsProduct.PriceRule
	if err := r.db.WithContext(ctx).Where("price_list_id = ?", priceListID).Order("product_id ASC, min_quantity ASC").Find(&rules).Error; err != nil {
		return nil, err
	}
	return rules, nil
}

// FindPriceRule returns a specific price rule.
func (r *PriceRepository) FindPriceRule(ctx context.Context, id string) (*modelsProduct.PriceRule, error) {
	var rule modelsProduct.PriceRule
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&rule).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// CreatePriceRule creates a new price rule.
func (r *PriceRepository) CreatePriceRule(ctx context.Context, rule *modelsProduct.PriceRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

// UpdatePriceRule updates a price rule.
func (r *PriceRepository) UpdatePriceRule(ctx context.Context, rule *modelsProduct.PriceRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

// DeletePriceRule deletes a price rule by ID.
func (r *PriceRepository) DeletePriceRule(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.PriceRule{}, "id = ?", id).Error
}

// FindBestPrice returns the best matching price for a product/priceList/quantity combo.
func (r *PriceRepository) FindBestPrice(ctx context.Context, productID, priceListID string, quantity int) (*modelsProduct.PriceRule, error) {
	var rule modelsProduct.PriceRule
	err := r.db.WithContext(ctx).
		Where("product_id = ? AND price_list_id = ? AND min_quantity <= ?", productID, priceListID, quantity).
		Order("min_quantity DESC").
		First(&rule).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}
