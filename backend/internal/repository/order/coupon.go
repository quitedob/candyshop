package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CouponRepository handles coupon and gift card data operations.
type CouponRepository struct {
	db *gorm.DB
}

// NewCouponRepository creates a new CouponRepository.
func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{db: db}
}

// CreateCoupon creates a new coupon.
func (r *CouponRepository) CreateCoupon(ctx context.Context, coupon *modelsOrder.Coupon) error {
	return r.db.WithContext(ctx).Create(coupon).Error
}

// FindAllCoupons returns paginated coupons.
func (r *CouponRepository) FindAllCoupons(ctx context.Context, page, limit int) ([]modelsOrder.Coupon, int64, error) {
	var coupons []modelsOrder.Coupon
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.Coupon{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&coupons).Error; err != nil {
		return nil, 0, err
	}
	return coupons, total, nil
}

// FindCouponByID returns a coupon by ID.
func (r *CouponRepository) FindCouponByID(ctx context.Context, id string) (*modelsOrder.Coupon, error) {
	var coupon modelsOrder.Coupon
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// UpdateCoupon updates a coupon.
func (r *CouponRepository) UpdateCoupon(ctx context.Context, coupon *modelsOrder.Coupon) error {
	coupon.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(coupon).Error
}

// DeleteCoupon soft-deletes a coupon by setting it inactive.
func (r *CouponRepository) DeleteCoupon(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&modelsOrder.Coupon{}).Where("id = ?", id).
		Update("status", modelsOrder.CouponStatusInactive).Error
}

// FindCouponByCode returns an active coupon by code.
func (r *CouponRepository) FindCouponByCode(ctx context.Context, code string) (*modelsOrder.Coupon, error) {
	var coupon modelsOrder.Coupon
	if err := r.db.WithContext(ctx).Where("code = ? AND status = ?", code, modelsOrder.CouponStatusActive).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

// ValidateCoupon checks if a coupon is valid for the given order amount and user.
func (r *CouponRepository) ValidateCoupon(ctx context.Context, coupon *modelsOrder.Coupon, userID string, orderAmount float64) error {
	now := time.Now()
	if coupon.Status != modelsOrder.CouponStatusActive {
		return fmt.Errorf("coupon is not active")
	}
	if now.Before(coupon.StartsAt) {
		return fmt.Errorf("coupon is not yet valid")
	}
	if now.After(coupon.ExpiresAt) {
		return fmt.Errorf("coupon has expired")
	}
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		return fmt.Errorf("coupon usage limit reached")
	}
	if coupon.MinOrderAmount > 0 && orderAmount < coupon.MinOrderAmount {
		return fmt.Errorf("minimum order amount %.2f not met", coupon.MinOrderAmount)
	}
	if coupon.MaxUsesPerUser > 0 {
		var userCount int64
		r.db.WithContext(ctx).Model(&modelsOrder.OrderDiscount{}).
			Where("coupon_id = ?", coupon.ID).
			Joins("JOIN orders ON orders.id = order_discounts.order_id").
			Where("orders.user_id = ?", userID).
			Count(&userCount)
		if int(userCount) >= coupon.MaxUsesPerUser {
			return fmt.Errorf("coupon already used maximum times by this user")
		}
	}
	return nil
}

// CalculateDiscount computes the discount amount for a coupon.
func CalculateDiscount(coupon *modelsOrder.Coupon, orderAmount float64) float64 {
	switch coupon.Type {
	case modelsOrder.CouponTypePercentage:
		return math.Round(orderAmount*coupon.Value) / 100
	case modelsOrder.CouponTypeFixed:
		if coupon.Value > orderAmount {
			return orderAmount
		}
		return coupon.Value
	}
	return 0
}

// IncrementCouponUsage atomically increments the used count of a coupon.
func (r *CouponRepository) IncrementCouponUsage(ctx context.Context, couponID string) error {
	return r.db.WithContext(ctx).Model(&modelsOrder.Coupon{}).
		Where("id = ? AND (max_uses = 0 OR used_count < max_uses)", couponID).
		Update("used_count", gorm.Expr("used_count + 1")).Error
}

// CreateGiftCard creates a new gift card.
func (r *CouponRepository) CreateGiftCard(ctx context.Context, gc *modelsOrder.GiftCard) error {
	return r.db.WithContext(ctx).Create(gc).Error
}

// FindAllGiftCards returns paginated gift cards.
func (r *CouponRepository) FindAllGiftCards(ctx context.Context, page, limit int) ([]modelsOrder.GiftCard, int64, error) {
	var cards []modelsOrder.GiftCard
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.GiftCard{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&cards).Error; err != nil {
		return nil, 0, err
	}
	return cards, total, nil
}

// FindGiftCardByCode returns an active gift card by code.
func (r *CouponRepository) FindGiftCardByCode(ctx context.Context, code string) (*modelsOrder.GiftCard, error) {
	var gc modelsOrder.GiftCard
	if err := r.db.WithContext(ctx).
		Where("code = ? AND status = ?", code, modelsOrder.GiftCardStatusActive).
		First(&gc).Error; err != nil {
		return nil, err
	}
	return &gc, nil
}

// DeductGiftCardBalance atomically deducts from a gift card balance.
func (r *CouponRepository) DeductGiftCardBalance(ctx context.Context, gcID string, amount float64) error {
	res := r.db.WithContext(ctx).Model(&modelsOrder.GiftCard{}).
		Where("id = ? AND current_balance >= ? AND status = ?", gcID, amount, modelsOrder.GiftCardStatusActive).
		Update("current_balance", gorm.Expr("current_balance - ?", amount))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("insufficient gift card balance")
	}
	// Mark as exhausted if balance is zero
	r.db.WithContext(ctx).Model(&modelsOrder.GiftCard{}).
		Where("id = ? AND current_balance <= 0", gcID).
		Update("status", modelsOrder.GiftCardStatusExhausted)
	return nil
}

// ApplyCouponToCart applies a coupon to an order and records the discount.
func (r *CouponRepository) ApplyCouponToCart(ctx context.Context, orderID, userID string, code string) (*modelsOrder.OrderDiscount, error) {
	coupon, err := r.FindCouponByCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("coupon not found")
	}

	var order modelsOrder.Order
	if err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
		return nil, fmt.Errorf("order not found")
	}

	if err := r.ValidateCoupon(ctx, coupon, userID, order.TotalAmount); err != nil {
		return nil, err
	}

	discountAmount := CalculateDiscount(coupon, order.TotalAmount)

	var discount *modelsOrder.OrderDiscount
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.IncrementCouponUsage(ctx, coupon.ID); err != nil {
			return err
		}
		discount = &modelsOrder.OrderDiscount{
			OrderID:  orderID,
			CouponID: &coupon.ID,
			Type:     "coupon",
			Amount:   discountAmount,
			Label:    fmt.Sprintf("Coupon: %s", coupon.Code),
		}
		if err := tx.Create(discount).Error; err != nil {
			return err
		}
		newTotal := order.TotalAmount - discountAmount
		if newTotal < 0 {
			newTotal = 0
		}
		return tx.Model(&order).Update("total_amount", newTotal).Error
	})
	if err != nil {
		return nil, err
	}
	return discount, nil
}

// RemoveCouponFromCart removes a coupon discount from an order.
func (r *CouponRepository) RemoveCouponFromCart(ctx context.Context, orderID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var discount modelsOrder.OrderDiscount
		if err := tx.Where("order_id = ? AND type = ?", orderID, "coupon").First(&discount).Error; err != nil {
			return fmt.Errorf("no coupon applied")
		}
		var order modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", orderID).First(&order).Error; err != nil {
			return err
		}
		// Restore order total
		newTotal := order.TotalAmount + discount.Amount
		if err := tx.Model(&order).Update("total_amount", newTotal).Error; err != nil {
			return err
		}
		return tx.Delete(&discount).Error
	})
}
