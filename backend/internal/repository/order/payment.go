package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/money"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PaymentRepository handles payment data operations.
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// FindByOrderID returns all payments for an order.
func (r *PaymentRepository) FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Payment, error) {
	var payments []modelsOrder.Payment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at DESC").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// FindByID returns a payment by ID.
func (r *PaymentRepository) FindByID(ctx context.Context, id string) (*modelsOrder.Payment, error) {
	var payment modelsOrder.Payment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// Create creates a new payment record.
func (r *PaymentRepository) Create(ctx context.Context, payment *modelsOrder.Payment) error {
	return r.db.WithContext(ctx).Create(payment).Error
}

// CreateWithBalanceCheck atomically checks remaining balance and creates a payment within a transaction.
// Prevents TOCTOU race where concurrent requests both pass balance validation (M8).
func (r *PaymentRepository) CreateWithBalanceCheck(ctx context.Context, orderTotalAmount float64, payment *modelsOrder.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the order row (SELECT ... FOR UPDATE) so concurrent checkouts —
		// including across multiple app instances — serialize on the row: the
		// second transaction observes the first's committed payment and fails the
		// balance check. The in-process mutex in PaymentService remains as a cheap
		// single-process pre-serializer.
		var order modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", payment.OrderID).First(&order).Error; err != nil {
			return err
		}
		var payments []modelsOrder.Payment
		if err := tx.Where("order_id = ?", payment.OrderID).Find(&payments).Error; err != nil {
			return err
		}
		confirmedTotal := 0.0
		pendingTotal := 0.0
		for _, p := range payments {
			if p.Status == "confirmed" {
				confirmedTotal += p.Amount
			}
			if p.Status == "pending" {
				pendingTotal += p.Amount
			}
		}
		allocated := confirmedTotal + pendingTotal
		remaining := orderTotalAmount - allocated
		if remaining < 0 {
			remaining = 0
		}
		// Reject any payment above the remaining balance. This must also fire when
		// remaining == 0 (fully-paid order): the previous `&& remaining > 0` clause
		// disabled the guard exactly in the double-submit case, letting a second
		// full payment land on an already-covered order (H7).
		if payment.Amount > remaining {
			return fmt.Errorf("payment amount %.2f exceeds remaining balance %.2f", payment.Amount, remaining)
		}
		return tx.Create(payment).Error
	})
}

// Update updates a payment record. M-9: bumps Version on every save so
// any concurrent guarded write fails loudly instead of silently overwriting.
func (r *PaymentRepository) Update(ctx context.Context, payment *modelsOrder.Payment) error {
	if payment != nil {
		payment.Version++
	}
	return r.db.WithContext(ctx).Save(payment).Error
}

// ConfirmPayment sets payment status to confirmed with admin user and timestamp.
// 仅当当前为 pending 时更新成功，避免并发双确认。
func (r *PaymentRepository) ConfirmPayment(ctx context.Context, id, confirmedBy string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&modelsOrder.Payment{}).
		Where("id = ? AND status = ?", id, "pending").
		Updates(map[string]interface{}{
			"status":       "confirmed",
			"confirmed_by": confirmedBy,
			"confirmed_at": now,
			"updated_at":   now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPaymentStateMismatch
	}
	return nil
}

// ConfirmPaymentAndRecomputeOrderStatus atomically confirms a pending payment and
// recomputes the parent order's payment_status (paid / partial / unpaid / refunded)
// inside a single transaction. Replaces the previous two-step pattern where
// ConfirmPayment committed first and then a separate updateOrderPaymentStatus call
// could fail and leave the order out of sync (H-13).
func (r *PaymentRepository) ConfirmPaymentAndRecomputeOrderStatus(ctx context.Context, id, confirmedBy string) (orderID, newStatus string, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&modelsOrder.Payment{}).
			Where("id = ? AND status = ?", id, "pending").
			Updates(map[string]interface{}{
				"status":       "confirmed",
				"confirmed_by": confirmedBy,
				"confirmed_at": now,
				"updated_at":   now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrPaymentStateMismatch
		}
		var p modelsOrder.Payment
		if err := tx.Where("id = ?", id).First(&p).Error; err != nil {
			return err
		}
		recomputed, err := recomputeOrderPaymentStatusTx(tx, p.OrderID)
		if err != nil {
			return err
		}
		orderID = p.OrderID
		newStatus = recomputed
		return nil
	})
	return
}

// MarkPaymentRefundedAndRecomputeOrderStatus atomically refunds a confirmed payment
// and recomputes the parent order's payment_status. See H-13.
func (r *PaymentRepository) MarkPaymentRefundedAndRecomputeOrderStatus(ctx context.Context, id string) (orderID, newStatus string, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&modelsOrder.Payment{}).
			Where("id = ? AND status = ?", id, "confirmed").
			Updates(map[string]interface{}{
				"status":     "refunded",
				"updated_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrPaymentStateMismatch
		}
		var p modelsOrder.Payment
		if err := tx.Where("id = ?", id).First(&p).Error; err != nil {
			return err
		}
		recomputed, err := recomputeOrderPaymentStatusTx(tx, p.OrderID)
		if err != nil {
			return err
		}
		orderID = p.OrderID
		newStatus = recomputed
		return nil
	})
	return
}

// ConfirmAuthorizedPaymentAndRecomputeOrderStatus is the atomic variant of
// ConfirmAuthorizedPayment that also recomputes the parent order status. See H-13.
func (r *PaymentRepository) ConfirmAuthorizedPaymentAndRecomputeOrderStatus(ctx context.Context, id, confirmedBy string, capturedAmount float64) (orderID, newStatus string, err error) {
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		res := tx.Model(&modelsOrder.Payment{}).
			Where("id = ? AND status = ?", id, modelsOrder.PaymentRecordStatusAuthorized).
			Updates(map[string]interface{}{
				"status":          modelsOrder.PaymentRecordStatusConfirmed,
				"confirmed_by":    confirmedBy,
				"confirmed_at":    now,
				"captured_amount": capturedAmount,
				"updated_at":      now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrPaymentStateMismatch
		}
		var p modelsOrder.Payment
		if err := tx.Where("id = ?", id).First(&p).Error; err != nil {
			return err
		}
		recomputed, err := recomputeOrderPaymentStatusTx(tx, p.OrderID)
		if err != nil {
			return err
		}
		orderID = p.OrderID
		newStatus = recomputed
		return nil
	})
	return
}

// recomputeOrderPaymentStatusTx loads the order and its payments inside the given
// transaction, applies the same business rules as PaymentService.updateOrderPaymentStatus,
// and persists the result. Returns the computed status (or the unchanged status if no
// update was needed).
func recomputeOrderPaymentStatusTx(tx *gorm.DB, orderID string) (string, error) {
	var order modelsOrder.Order
	if err := tx.Where("id = ?", orderID).First(&order).Error; err != nil {
		return "", err
	}
	var payments []modelsOrder.Payment
	if err := tx.Where("order_id = ?", orderID).Find(&payments).Error; err != nil {
		return "", err
	}
	confirmedTotal := 0.0
	hasRefunded := false
	for _, p := range payments {
		if p.Status == "confirmed" {
			confirmedTotal += p.Amount
		}
		if p.Status == "refunded" {
			hasRefunded = true
		}
	}
	newStatus := "unpaid"
	if hasRefunded && confirmedTotal == 0 {
		newStatus = "refunded"
	} else if money.MoneyCoversTotal(confirmedTotal, order.TotalAmount) {
		newStatus = "paid"
	} else if confirmedTotal > 0 {
		newStatus = "partial"
	}
	if order.PaymentStatus == newStatus {
		return newStatus, nil
	}
	res := tx.Model(&modelsOrder.Order{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"payment_status": newStatus,
			"updated_at":     time.Now(),
		})
	if res.Error != nil {
		return "", res.Error
	}
	return newStatus, nil
}

// ConfirmAuthorizedPayment sets an authorized payment to confirmed with capture amount.
func (r *PaymentRepository) ConfirmAuthorizedPayment(ctx context.Context, id, confirmedBy string, capturedAmount float64) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&modelsOrder.Payment{}).
		Where("id = ? AND status = ?", id, modelsOrder.PaymentRecordStatusAuthorized).
		Updates(map[string]interface{}{
			"status":          modelsOrder.PaymentRecordStatusConfirmed,
			"confirmed_by":    confirmedBy,
			"confirmed_at":    now,
			"captured_amount": capturedAmount,
			"updated_at":      now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPaymentStateMismatch
	}
	return nil
}

// MarkPaymentRefunded 将已确认付款标为已退款（仅 status=confirmed 时生效，避免并发双退）
func (r *PaymentRepository) MarkPaymentRefunded(ctx context.Context, id string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&modelsOrder.Payment{}).
		Where("id = ? AND status = ?", id, "confirmed").
		Updates(map[string]interface{}{
			"status":     "refunded",
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPaymentStateMismatch
	}
	return nil
}

// UpdateStatus updates payment status with state transition validation.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, currentStatus, newStatus string) error {
	if err := modelsOrder.ValidatePaymentStatusTransition(currentStatus, newStatus); err != nil {
		return err
	}
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&modelsOrder.Payment{}).
		Where("id = ? AND status = ?", id, currentStatus).
		Updates(map[string]interface{}{
			"status":     newStatus,
			"updated_at": now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrPaymentStateMismatch
	}
	return nil
}

// CountByStatus returns count of payments with given status.
func (r *PaymentRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var result struct {
		Count int64
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Payment{}).
		Where("status = ?", status).
		Select("COUNT(*) AS count").
		Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Count, nil
}

// StatusBreakdown returns payment counts grouped by status.
func (r *PaymentRepository) StatusBreakdown(ctx context.Context) (map[string]int64, error) {
	type row struct {
		Status string
		Count  int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Payment{}).
		Select("status, COUNT(*) AS count").
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, r := range rows {
		m[r.Status] = r.Count
	}
	return m, nil
}

// FindByGatewayTransactionID finds a payment by its gateway (Stripe/PayPal) transaction ID.
func (r *PaymentRepository) FindByGatewayTransactionID(ctx context.Context, txID string) (*modelsOrder.Payment, error) {
	var payment modelsOrder.Payment
	if err := r.db.WithContext(ctx).Where("gateway_transaction_id = ?", txID).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}
