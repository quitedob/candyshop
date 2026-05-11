package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"

	"gorm.io/gorm"
)

// OrderMessageRepository handles order message data operations.
type OrderMessageRepository struct {
	db *gorm.DB
}

// NewOrderMessageRepository creates a new OrderMessageRepository.
func NewOrderMessageRepository(db *gorm.DB) *OrderMessageRepository {
	return &OrderMessageRepository{db: db}
}

// FindByOrderID returns paginated messages for an order, ordered oldest-first.
func (r *OrderMessageRepository) FindByOrderID(ctx context.Context, orderID string, page, limit int) ([]modelsOrder.OrderMessage, int64, error) {
	var messages []modelsOrder.OrderMessage
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsOrder.OrderMessage{}).Where("order_id = ?", orderID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at ASC").Find(&messages).Error; err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// Create creates a new order message.
func (r *OrderMessageRepository) Create(ctx context.Context, msg *modelsOrder.OrderMessage) error {
	return r.db.WithContext(ctx).Create(msg).Error
}
