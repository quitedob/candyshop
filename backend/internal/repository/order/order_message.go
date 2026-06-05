package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"time"

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

// MarkRead marks unread messages sent by the opposite party as read.
func (r *OrderMessageRepository) MarkRead(ctx context.Context, orderID, readerType string, readAt time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Model(&modelsOrder.OrderMessage{}).
		Where("order_id = ? AND sender_type <> ? AND read_at IS NULL", orderID, readerType).
		Updates(map[string]interface{}{"read_at": readAt, "updated_at": readAt})
	return res.RowsAffected, res.Error
}

// CountUnread returns the unread message count for one order and reader type.
func (r *OrderMessageRepository) CountUnread(ctx context.Context, orderID, readerType string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&modelsOrder.OrderMessage{}).
		Where("order_id = ? AND sender_type <> ? AND read_at IS NULL", orderID, readerType).
		Count(&total).Error
	return total, err
}

// CountUnreadForCustomer returns unread admin messages across a customer's orders.
func (r *OrderMessageRepository) CountUnreadForCustomer(ctx context.Context, userID string) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&modelsOrder.OrderMessage{}).
		Joins("JOIN orders ON orders.id = order_messages.order_id").
		Where("orders.user_id = ? AND order_messages.sender_type = ? AND order_messages.read_at IS NULL", userID, "admin").
		Count(&total).Error
	return total, err
}

// CountUnreadForAdmin returns unread customer messages across all orders.
func (r *OrderMessageRepository) CountUnreadForAdmin(ctx context.Context) (int64, error) {
	var total int64
	err := r.db.WithContext(ctx).Model(&modelsOrder.OrderMessage{}).
		Where("sender_type = ? AND read_at IS NULL", "customer").
		Count(&total).Error
	return total, err
}
