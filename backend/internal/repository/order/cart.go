package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

// UpsertItem atomically inserts or updates a cart item.
// Uses INSERT ... ON CONFLICT DO UPDATE to prevent all TOCTOU race conditions,
// backed by a unique constraint on (user_id, product_id).
func (r *CartRepository) UpsertItem(ctx context.Context, item *modelsOrder.CartItem) (*modelsOrder.CartItem, error) {
	item.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "product_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"quantity":       gorm.Expr("cart_items.quantity + ?", item.Quantity),
				"unit_price":     gorm.Expr("excluded.unit_price"),
				"product_name":   gorm.Expr("excluded.product_name"),
				"currency":       gorm.Expr("excluded.currency"),
				"specifications": gorm.Expr("excluded.specifications"),
				"updated_at":     gorm.Expr("excluded.updated_at"),
			}),
		}).
		Create(item).Error
	if err != nil {
		return nil, err
	}
	return item, nil
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
