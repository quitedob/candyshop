package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
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
// Prevents TOCTOU race where concurrent requests both pass balance validation.
func (r *PaymentRepository) CreateWithBalanceCheck(ctx context.Context, orderTotalAmount float64, payment *modelsOrder.Payment) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		if payment.Amount > remaining && remaining > 0 {
			return fmt.Errorf("payment amount %.2f exceeds remaining balance %.2f", payment.Amount, remaining)
		}
		return tx.Create(payment).Error
	})
}

// Update updates a payment record.
func (r *PaymentRepository) Update(ctx context.Context, payment *modelsOrder.Payment) error {
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
