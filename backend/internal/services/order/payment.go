package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/money"
	orderrepo "candypro/api/internal/repository/order"
	"context"
	"errors"
	"fmt"
)

type paymentRepository interface {
	FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Payment, error)
	FindByID(ctx context.Context, id string) (*modelsOrder.Payment, error)
	FindByGatewayTransactionID(ctx context.Context, txID string) (*modelsOrder.Payment, error)
	Create(ctx context.Context, payment *modelsOrder.Payment) error
	CreateWithBalanceCheck(ctx context.Context, orderTotalAmount float64, payment *modelsOrder.Payment) error
	Update(ctx context.Context, payment *modelsOrder.Payment) error
	ConfirmPayment(ctx context.Context, id, confirmedBy string) error
	MarkPaymentRefunded(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id, currentStatus, newStatus string) error
	CountByStatus(ctx context.Context, status string) (int64, error)
	StatusBreakdown(ctx context.Context) (map[string]int64, error)
}

// PaymentService handles payment business logic.
type PaymentService struct {
	repo      paymentRepository
	orderRepo orderRepository
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(repo paymentRepository, orderRepo orderRepository) *PaymentService {
	return &PaymentService{repo: repo, orderRepo: orderRepo}
}

// GetPaymentsByOrder returns all payments for an order.
func (s *PaymentService) GetPaymentsByOrder(ctx context.Context, orderID string) ([]modelsOrder.Payment, error) {
	return s.repo.FindByOrderID(ctx, orderID)
}

// GetPayment returns a single payment by ID.
func (s *PaymentService) GetPayment(ctx context.Context, id string) (*modelsOrder.Payment, error) {
	return s.repo.FindByID(ctx, id)
}

// CreatePayment creates a new payment record.
func (s *PaymentService) CreatePayment(ctx context.Context, payment *modelsOrder.Payment) error {
	if payment.Status == "" {
		payment.Status = "pending"
	}
	return s.repo.Create(ctx, payment)
}

// CreatePaymentWithBalanceCheck creates a payment with atomic balance enforcement.
func (s *PaymentService) CreatePaymentWithBalanceCheck(ctx context.Context, orderTotalAmount float64, payment *modelsOrder.Payment) error {
	if payment.Status == "" {
		payment.Status = "pending"
	}
	return s.repo.CreateWithBalanceCheck(ctx, orderTotalAmount, payment)
}

// ConfirmPayment confirms a payment and updates the order's payment status atomically.
func (s *PaymentService) ConfirmPayment(ctx context.Context, paymentID, confirmedBy string) error {
	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}
	if payment.Status != "pending" {
		return fmt.Errorf("payment cannot be confirmed: current status is '%s'", payment.Status)
	}

	if err := s.repo.ConfirmPayment(ctx, paymentID, confirmedBy); err != nil {
		if errors.Is(err, orderrepo.ErrPaymentStateMismatch) {
			return fmt.Errorf("payment cannot be confirmed: not pending or already processed")
		}
		return err
	}

	return s.updateOrderPaymentStatus(ctx, payment.OrderID)
}

// RefundPayment marks a payment as refunded and updates order status.
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string) error {
	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}
	if payment.Status != "confirmed" {
		return fmt.Errorf("only confirmed payments can be refunded: current status is '%s'", payment.Status)
	}

	if err := s.repo.MarkPaymentRefunded(ctx, paymentID); err != nil {
		if errors.Is(err, orderrepo.ErrPaymentStateMismatch) {
			return fmt.Errorf("payment cannot be refunded: not in confirmed state or concurrent update")
		}
		return err
	}

	return s.updateOrderPaymentStatus(ctx, payment.OrderID)
}

// updateOrderPaymentStatus recalculates the order payment status based on its payments.
func (s *PaymentService) updateOrderPaymentStatus(ctx context.Context, orderID string) error {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return err
	}

	payments, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return err
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

	if order.PaymentStatus != newStatus {
		order.PaymentStatus = newStatus
		return s.orderRepo.Update(ctx, order)
	}
	return nil
}

// AuthorizePayment transitions a payment from pending to authorized (Stripe manual-capture flow).
func (s *PaymentService) AuthorizePayment(ctx context.Context, paymentID string) error {
	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}
	if payment.Status != modelsOrder.PaymentRecordStatusPending {
		return fmt.Errorf("cannot authorize payment with status '%s'", payment.Status)
	}
	return s.repo.UpdateStatus(ctx, paymentID, payment.Status, modelsOrder.PaymentRecordStatusAuthorized)
}

// FailPayment transitions a payment from pending/authorized to failed.
func (s *PaymentService) FailPayment(ctx context.Context, paymentID string) error {
	payment, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("payment not found: %w", err)
	}
	if payment.Status != modelsOrder.PaymentRecordStatusPending && payment.Status != modelsOrder.PaymentRecordStatusAuthorized {
		return fmt.Errorf("cannot fail payment with status '%s'", payment.Status)
	}
	return s.repo.UpdateStatus(ctx, paymentID, payment.Status, modelsOrder.PaymentRecordStatusFailed)
}

// FindByGatewayTransactionID looks up a payment by its Stripe/PayPal transaction ID.
func (s *PaymentService) FindByGatewayTransactionID(ctx context.Context, txID string) (*modelsOrder.Payment, error) {
	return s.repo.FindByGatewayTransactionID(ctx, txID)
}

// CountByStatus returns count of payments with given status.
func (s *PaymentService) CountByStatus(ctx context.Context, status string) (int64, error) {
	return s.repo.CountByStatus(ctx, status)
}

// StatusBreakdown returns payment counts grouped by status.
func (s *PaymentService) StatusBreakdown(ctx context.Context) (map[string]int64, error) {
	return s.repo.StatusBreakdown(ctx)
}
