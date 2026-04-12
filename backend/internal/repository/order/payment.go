package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
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

// Update updates a payment record.
func (r *PaymentRepository) Update(ctx context.Context, payment *modelsOrder.Payment) error {
	return r.db.WithContext(ctx).Save(payment).Error
}

// ConfirmPayment sets payment status to confirmed with admin user and timestamp.
func (r *PaymentRepository) ConfirmPayment(ctx context.Context, id, confirmedBy string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&modelsOrder.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       "confirmed",
			"confirmed_by": confirmedBy,
			"confirmed_at": now,
			"updated_at":   now,
		}).Error
}

// UpdateStatus updates payment status.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id, status string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&modelsOrder.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": now,
		}).Error
}
