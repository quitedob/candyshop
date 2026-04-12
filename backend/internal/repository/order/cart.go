package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"

	"gorm.io/gorm"
)

// CartRepository handles cart item persistence.
type CartRepository struct {
	db *gorm.DB
}

// NewCartRepository creates a new CartRepository.
func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

// FindByUserID returns all cart items for a user.
func (r *CartRepository) FindByUserID(ctx context.Context, userID string) ([]modelsOrder.CartItem, error) {
	var items []modelsOrder.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at asc").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

// FindByID returns a single cart item by ID.
func (r *CartRepository) FindByID(ctx context.Context, id uint) (*modelsOrder.CartItem, error) {
	var item modelsOrder.CartItem
	if err := r.db.WithContext(ctx).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// FindByUserIDAndProductID returns an existing cart item for a user+product combo.
func (r *CartRepository) FindByUserIDAndProductID(ctx context.Context, userID, productID string) (*modelsOrder.CartItem, error) {
	var item modelsOrder.CartItem
	if err := r.db.WithContext(ctx).Where("user_id = ? AND product_id = ?", userID, productID).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// Create adds a cart item.
func (r *CartRepository) Create(ctx context.Context, item *modelsOrder.CartItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// Update saves changes to a cart item.
func (r *CartRepository) Update(ctx context.Context, item *modelsOrder.CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete removes a single cart item.
func (r *CartRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsOrder.CartItem{}, id).Error
}

// UpsertItem atomically increments quantity if item exists, or creates it.
// Uses FirstOrCreate + atomic UPDATE to prevent TOCTOU race conditions.
func (r *CartRepository) UpsertItem(ctx context.Context, item *modelsOrder.CartItem) (*modelsOrder.CartItem, error) {
	existing := &modelsOrder.CartItem{}
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND product_id = ?", item.UserID, item.ProductID).
		FirstOrCreate(existing, modelsOrder.CartItem{
			UserID:         item.UserID,
			ProductID:      item.ProductID,
			ProductName:    item.ProductName,
			Quantity:       item.Quantity,
			UnitPrice:      item.UnitPrice,
			Currency:       item.Currency,
			Specifications: item.Specifications,
		})
	if result.Error != nil {
		return nil, result.Error
	}
	// Record already existed — increment quantity atomically
	if result.RowsAffected == 0 {
		updates := map[string]interface{}{
			"quantity": gorm.Expr("quantity + ?", item.Quantity),
		}
		if item.UnitPrice > 0 {
			updates["unit_price"] = item.UnitPrice
		}
		if err := r.db.WithContext(ctx).Model(existing).Updates(updates).Error; err != nil {
			return nil, err
		}
		// Reload to get updated values
		if err := r.db.WithContext(ctx).First(existing, existing.ID).Error; err != nil {
			return nil, err
		}
	}
	return existing, nil
}

// ClearByUserID removes all cart items for a user.
func (r *CartRepository) ClearByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&modelsOrder.CartItem{}).Error
}

// CountByUserID returns the number of items in a user's cart.
func (r *CartRepository) CountByUserID(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&modelsOrder.CartItem{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
