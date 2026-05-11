package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type orderMessageRepository interface {
	FindByOrderID(ctx context.Context, orderID string, page, limit int) ([]modelsOrder.OrderMessage, int64, error)
	Create(ctx context.Context, msg *modelsOrder.OrderMessage) error
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
