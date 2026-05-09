package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"errors"
)

type cartRepository interface {
	FindByUserID(ctx context.Context, userID string) ([]modelsOrder.CartItem, error)
	FindByID(ctx context.Context, id uint) (*modelsOrder.CartItem, error)
	FindByUserIDAndProductID(ctx context.Context, userID, productID string) (*modelsOrder.CartItem, error)
	Create(ctx context.Context, item *modelsOrder.CartItem) error
	Update(ctx context.Context, item *modelsOrder.CartItem) error
	Delete(ctx context.Context, id uint) error
	ClearByUserID(ctx context.Context, userID string) error
	CountByUserID(ctx context.Context, userID string) (int64, error)
	UpsertItem(ctx context.Context, item *modelsOrder.CartItem) (*modelsOrder.CartItem, error)
}

// CartService provides cart business logic.
type CartService struct {
	repo cartRepository
}

// NewCartService creates a CartService.
func NewCartService(repo cartRepository) *CartService {
	return &CartService{repo: repo}
}

// GetCart returns all items in a user's cart.
func (s *CartService) GetCart(ctx context.Context, userID string) ([]modelsOrder.CartItem, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// AddItem adds a product to the cart, merging quantity atomically if the product already exists.
func (s *CartService) AddItem(ctx context.Context, userID string, item *modelsOrder.CartItem) (*modelsOrder.CartItem, error) {
	if item.Quantity < 1 {
		item.Quantity = 1
	}
	item.UserID = userID

	return s.repo.UpsertItem(ctx, item)
}

// UpdateItem changes the quantity of a cart item.
func (s *CartService) UpdateItem(ctx context.Context, userID string, itemID uint, quantity int) (*modelsOrder.CartItem, error) {
	if quantity < 1 {
		return nil, errors.New("quantity must be at least 1")
	}
	item, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item.UserID != userID {
		return nil, errors.New("cart item not found")
	}
	item.Quantity = quantity
	if err := s.repo.Update(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

// RemoveItem deletes a single cart item, enforcing ownership.
func (s *CartService) RemoveItem(ctx context.Context, userID string, itemID uint) error {
	item, err := s.repo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item.UserID != userID {
		return errors.New("cart item not found")
	}
	return s.repo.Delete(ctx, itemID)
}

// ClearCart removes all items for a user.
func (s *CartService) ClearCart(ctx context.Context, userID string) error {
	return s.repo.ClearByUserID(ctx, userID)
}

// ItemCount returns the total number of cart items for a user.
func (s *CartService) ItemCount(ctx context.Context, userID string) (int64, error) {
	return s.repo.CountByUserID(ctx, userID)
}
