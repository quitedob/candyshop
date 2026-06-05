package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"fmt"
	"time"
)

type orderMessageRepository interface {
	FindByOrderID(ctx context.Context, orderID string, page, limit int) ([]modelsOrder.OrderMessage, int64, error)
	Create(ctx context.Context, msg *modelsOrder.OrderMessage) error
	MarkRead(ctx context.Context, orderID, readerType string, readAt time.Time) (int64, error)
	CountUnread(ctx context.Context, orderID, readerType string) (int64, error)
	CountUnreadForCustomer(ctx context.Context, userID string) (int64, error)
	CountUnreadForAdmin(ctx context.Context) (int64, error)
}

// OrderMessageService handles order message business logic.
type OrderMessageService struct {
	repo orderMessageRepository
}

// NewOrderMessageService creates a new OrderMessageService.
func NewOrderMessageService(repo orderMessageRepository) *OrderMessageService {
	return &OrderMessageService{repo: repo}
}

// GetMessages returns paginated messages for an order.
func (s *OrderMessageService) GetMessages(ctx context.Context, orderID string, page, limit int) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	messages, total, err := s.repo.FindByOrderID(ctx, orderID, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: messages,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// SendMessage creates a new order message.
func (s *OrderMessageService) SendMessage(ctx context.Context, msg *modelsOrder.OrderMessage) error {
	return s.repo.Create(ctx, msg)
}

// MarkRead marks the opposite party's unread messages as read.
func (s *OrderMessageService) MarkRead(ctx context.Context, orderID, readerType string) (int64, time.Time, error) {
	if readerType != "customer" && readerType != "admin" {
		return 0, time.Time{}, fmt.Errorf("invalid reader type %q", readerType)
	}
	readAt := time.Now()
	n, err := s.repo.MarkRead(ctx, orderID, readerType, readAt)
	return n, readAt, err
}

// CountUnread returns unread messages for a single order.
func (s *OrderMessageService) CountUnread(ctx context.Context, orderID, readerType string) (int64, error) {
	if readerType != "customer" && readerType != "admin" {
		return 0, fmt.Errorf("invalid reader type %q", readerType)
	}
	return s.repo.CountUnread(ctx, orderID, readerType)
}

// CountUnreadForCustomer returns unread admin messages across the customer's orders.
func (s *OrderMessageService) CountUnreadForCustomer(ctx context.Context, userID string) (int64, error) {
	return s.repo.CountUnreadForCustomer(ctx, userID)
}

// CountUnreadForAdmin returns unread customer messages across all orders.
func (s *OrderMessageService) CountUnreadForAdmin(ctx context.Context) (int64, error) {
	return s.repo.CountUnreadForAdmin(ctx)
}
