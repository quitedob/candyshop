package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/money"
	"context"
	"errors"
	"fmt"
	"log"
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

// Sentinel errors returned by coupon apply/remove so handlers can map the
// outcome to the appropriate HTTP status (M3). The caller must use errors.Is
// to distinguish them from generic failures.
var (
	// ErrCouponOrderForbidden indicates the target order does not exist or does
	// not belong to the caller. Returned instead of gorm.ErrRecordNotFound so a
	// cross-tenant attempt cannot distinguish "missing" from "not yours".
	ErrCouponOrderForbidden = errors.New("coupon: order not found or not owned by the caller")
	// ErrCouponOrderInvalidState indicates the order has moved past the mutable
	// draft states (pending / pending_confirmation / pending_approval), so its
	// booked totals can no longer be changed by the customer.
	ErrCouponOrderInvalidState = errors.New("coupon: order status does not allow coupon changes")
)

// isCouponMutableOrderStatus reports whether an order is still in a state where
// coupon changes are permitted. Once an order is paid or leaves the draft-like
// states its booked subtotal/tax/total are treated as immutable, otherwise a
// coupon removal after payment could inflate the payable amount (M3).
func isCouponMutableOrderStatus(status string) bool {
	switch status {
	case modelsOrder.OrderStatusPending,
		modelsOrder.OrderStatusPendingConfirm,
		modelsOrder.OrderStatusPendingApproval:
		return true
	}
	return false
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
//
// db is the connection to query against. Callers inside a Transaction MUST
// pass the transaction's tx so the validation reads see (and lock) the same
// rows that the subsequent usage increment writes — see ApplyCouponToCart for
// the SELECT … FOR UPDATE coupling that closes the TOCTOU window (R2 A-2).
func (r *CouponRepository) ValidateCoupon(db *gorm.DB, coupon *modelsOrder.Coupon, userID string, orderAmount float64) error {
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
		if err := db.Model(&modelsOrder.OrderDiscount{}).
			Where("coupon_id = ?", coupon.ID).
			Joins("JOIN orders ON orders.id = order_discounts.order_id").
			Where("orders.user_id = ?", userID).
			Count(&userCount).Error; err != nil {
			return fmt.Errorf("validate per-user coupon usage: %w", err)
		}
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
//
// db must be the same handle (typically a Transaction tx) used to write the
// OrderDiscount row, so a rollback also rolls back the increment — otherwise
// failed orders permanently consume coupon uses (R2 A-1). The WHERE clause
// makes this the atomic guard for the usage-cap race: callers must check that
// RowsAffected == 1 to detect a lost race against a concurrent applier.
func (r *CouponRepository) IncrementCouponUsage(db *gorm.DB, couponID string) (int64, error) {
	res := db.Model(&modelsOrder.Coupon{}).
		Where("id = ? AND (max_uses = 0 OR used_count < max_uses)", couponID).
		Update("used_count", gorm.Expr("used_count + 1"))
	return res.RowsAffected, res.Error
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
	// Mark as exhausted if balance is zero. R2 A-9: previously this UPDATE's
	// error was discarded, leaving zero-balance cards still flagged "active".
	// Failures are logged so the result endpoint stays simple, but the row
	// drift is now visible in observability instead of silently invisible.
	if upErr := r.db.WithContext(ctx).Model(&modelsOrder.GiftCard{}).
		Where("id = ? AND current_balance <= 0", gcID).
		Update("status", modelsOrder.GiftCardStatusExhausted).Error; upErr != nil {
		// Best-effort — the deduction itself already succeeded above.
		log.Printf("coupon: mark gift card exhausted failed for %s: %v", gcID, upErr)
	}
	return nil
}

// PreviewCouponDiscount validates a coupon and returns the discount amount without applying it.
func (r *CouponRepository) PreviewCouponDiscount(ctx context.Context, code, userID string, orderAmount float64) (float64, *modelsOrder.Coupon, error) {
	coupon, err := r.FindCouponByCode(ctx, code)
	if err != nil {
		return 0, nil, fmt.Errorf("coupon not found")
	}
	if err := r.ValidateCoupon(r.db.WithContext(ctx), coupon, userID, orderAmount); err != nil {
		return 0, nil, err
	}
	return CalculateDiscount(coupon, orderAmount), coupon, nil
}

// ApplyCouponToCart applies a coupon to an order and records the discount.
//
// B2B tax compliance (H-20): discounts must apply to the order subtotal — never
// the gross total — because tax is calculated on the post-discount net amount.
// Reducing TotalAmount directly silently lowers the booked tax base, which can
// be challenged by tax authorities. Validation and computation therefore use
// Subtotal; TotalAmount is rebuilt from (Subtotal - Discount + Tax + Shipping).
//
// Transaction integrity (R2 A-1, A-2): the coupon row is locked with
// SELECT … FOR UPDATE inside the transaction, validated, and the usage
// increment is performed against the same tx. If any subsequent step fails
// the rollback also reverts the increment, so failed orders no longer waste a
// coupon use. The lock + RowsAffected check on the increment closes the race
// where two concurrent applications could both consume the coupon's last
// remaining use.
func (r *CouponRepository) ApplyCouponToCart(ctx context.Context, orderID, userID string, code string) (*modelsOrder.OrderDiscount, error) {
	var discount *modelsOrder.OrderDiscount
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var coupon modelsOrder.Coupon
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("code = ? AND status = ?", code, modelsOrder.CouponStatusActive).
			First(&coupon).Error; err != nil {
			return fmt.Errorf("coupon not found")
		}

		var order modelsOrder.Order
		if err := tx.Where("id = ? AND user_id = ?", orderID, userID).First(&order).Error; err != nil {
			return ErrCouponOrderForbidden
		}
		if !isCouponMutableOrderStatus(order.Status) {
			return ErrCouponOrderInvalidState
		}

		if err := r.ValidateCoupon(tx, &coupon, userID, order.Subtotal); err != nil {
			return err
		}

		discountAmount := CalculateDiscount(&coupon, order.Subtotal)
		if discountAmount > order.Subtotal {
			discountAmount = order.Subtotal
		}

		rows, err := r.IncrementCouponUsage(tx, coupon.ID)
		if err != nil {
			return err
		}
		if rows == 0 {
			// Lost the race against a concurrent applier or hit the cap between
			// the locked validate and the conditional update — treat as exhausted.
			return fmt.Errorf("coupon usage limit reached")
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
		// M7: tax must be booked on the post-discount net subtotal, never the
		// gross subtotal — the booked tax base is (subtotal - discount). Tax is
		// proportional to its base (tax = base × rate), so the discounted base
		// recomputes the already-booked tax proportionally. Subtotal and
		// ShippingAmount are left untouched so the invariant
		// total = subtotal - discount + tax + shipping holds exactly.
		netSubtotal := order.Subtotal - discountAmount
		if netSubtotal < 0 {
			netSubtotal = 0
		}
		newTax := order.TaxAmount
		if order.Subtotal > 0 {
			newTax = money.RoundMoney(order.TaxAmount * netSubtotal / order.Subtotal)
		}
		newTotal := netSubtotal + newTax + order.ShippingAmount
		if newTotal < 0 {
			newTotal = 0
		}
		return tx.Model(&order).Updates(map[string]interface{}{
			"tax_amount":   newTax,
			"total_amount": newTotal,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return discount, nil
}

// RemoveCouponFromCart removes a coupon discount from an order, refunds the
// coupon's global usage count, and restores the order totals (M3).
//
// Security (M3): the order row is locked and scoped to the owning user
// (WHERE id = ? AND user_id = ?), so one tenant can never remove another
// tenant's coupon or mutate another order's total_amount. A status guard
// refuses removal once the order has moved past the mutable draft states, so a
// paid/confirmed order's booked total cannot be inflated after payment.
//
// Usage symmetry (M3): IncrementCouponUsage is an atomic conditional increment;
// removal refunds the use with the symmetric conditional decrement (never below
// 0) in the same transaction. apply+remove therefore no longer permanently
// burns the global MaxUses cap.
func (r *CouponRepository) RemoveCouponFromCart(ctx context.Context, orderID, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var order modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND user_id = ?", orderID, userID).
			First(&order).Error; err != nil {
			return ErrCouponOrderForbidden
		}
		if !isCouponMutableOrderStatus(order.Status) {
			return ErrCouponOrderInvalidState
		}

		var discount modelsOrder.OrderDiscount
		if err := tx.Where("order_id = ? AND type = ?", orderID, "coupon").First(&discount).Error; err != nil {
			return fmt.Errorf("no coupon applied")
		}

		// Refund the coupon's global usage count (never below 0). The WHERE
		// guard mirrors IncrementCouponUsage's conditional UPDATE, so a
		// concurrent increment on the same coupon can't drive used_count
		// negative and this can't refund a use that was never charged.
		if discount.CouponID != nil {
			res := tx.Model(&modelsOrder.Coupon{}).
				Where("id = ? AND used_count > 0", *discount.CouponID).
				Update("used_count", gorm.Expr("used_count - 1"))
			if res.Error != nil {
				return res.Error
			}
		}

		// Restore totals to the pre-coupon snapshot. Subtotal is the pre-coupon
		// base (never mutated by ApplyCouponToCart), so net = subtotal - discount.
		// ApplyCouponToCart booked tax on the discounted net (M7); restoring the
		// full-subtotal tax uses the inverse proportional recompute so the
		// invariant total = subtotal + tax + shipping holds again.
		netSubtotal := order.Subtotal - discount.Amount
		if netSubtotal < 0 {
			netSubtotal = 0
		}
		restoredTax := order.TaxAmount
		if order.Subtotal > 0 && netSubtotal > 0 {
			restoredTax = money.RoundMoney(order.TaxAmount * order.Subtotal / netSubtotal)
		}
		newTotal := money.RoundMoney(order.Subtotal + restoredTax + order.ShippingAmount)
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"tax_amount":   restoredTax,
			"total_amount": newTotal,
		}).Error; err != nil {
			return err
		}
		return tx.Delete(&discount).Error
	})
}
