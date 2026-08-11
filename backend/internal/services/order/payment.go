package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/money"
	orderrepo "candypro/api/internal/repository/order"
	"context"
	"errors"
	"fmt"
	"sync"
)

type paymentRepository interface {
	FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Payment, error)
	FindByID(ctx context.Context, id string) (*modelsOrder.Payment, error)
	FindByGatewayTransactionID(ctx context.Context, txID string) (*modelsOrder.Payment, error)
	Create(ctx context.Context, payment *modelsOrder.Payment) error
	CreateWithBalanceCheck(ctx context.Context, orderTotalAmount float64, payment *modelsOrder.Payment) error
	Update(ctx context.Context, payment *modelsOrder.Payment) error
	ConfirmPayment(ctx context.Context, id, confirmedBy string) error
	ConfirmPaymentAndRecomputeOrderStatus(ctx context.Context, id, confirmedBy string) (string, string, error)
	MarkPaymentRefunded(ctx context.Context, id string) error
	MarkPaymentRefundedAndRecomputeOrderStatus(ctx context.Context, id string) (string, string, error)
	UpdateStatus(ctx context.Context, id, currentStatus, newStatus string) error
	ConfirmAuthorizedPayment(ctx context.Context, id, confirmedBy string, capturedAmount float64) error
	ConfirmAuthorizedPaymentAndRecomputeOrderStatus(ctx context.Context, id, confirmedBy string, capturedAmount float64) (string, string, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	StatusBreakdown(ctx context.Context) (map[string]int64, error)
}

// PaymentService handles payment business logic.
type PaymentService struct {
	repo      paymentRepository
	orderRepo orderRepository
	// checkoutMu serializes CreatePaymentWithBalanceCheck within this process.
	// The repo's CreateWithBalanceCheck now runs under SELECT ... FOR UPDATE on
	// the order row, so multi-instance safety is guaranteed at the DB level (M8).
	// This mutex remains as a cheap single-process pre-serializer that bounds DB
	// lock wait for the common case.
	checkoutMu sync.Mutex
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
// The mutex serializes the balance-check + insert so concurrent checkout calls for
// the same order cannot both pass when their combined amount exceeds the remaining
// balance (M8). The amount is committed to the order's allocated balance before the
// mutex is released.
func (s *PaymentService) CreatePaymentWithBalanceCheck(ctx context.Context, orderTotalAmount float64, payment *modelsOrder.Payment) error {
	if payment.Status == "" {
		payment.Status = "pending"
	}
	s.checkoutMu.Lock()
	defer s.checkoutMu.Unlock()
	return s.repo.CreateWithBalanceCheck(ctx, orderTotalAmount, payment)
}

// AttachGatewayResult records the gateway transaction ID + raw response on a
// persisted payment row after the gateway authorize call succeeds. The row is
// created (status pending) BEFORE the gateway call (M8) so a gateway failure can
// mark it failed instead of orphaning a collectible intent; the transaction ID is
// back-filled here once the gateway returns.
func (s *PaymentService) AttachGatewayResult(ctx context.Context, paymentID, txID, raw string) error {
	pay, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}
	if pay.Status != modelsOrder.PaymentRecordStatusPending && pay.Status != modelsOrder.PaymentRecordStatusAuthorized {
		return fmt.Errorf("cannot attach gateway result to payment with status '%s'", pay.Status)
	}
	pay.GatewayTransactionID = &txID
	pay.GatewayMetadata = raw
	return s.repo.Update(ctx, pay)
}

// ConfirmPayment confirms a payment and updates the order's payment status atomically.
// Both the payment update and the order payment_status recomputation happen in a single
// DB transaction so the two records cannot drift if one of the writes fails (H-13).
func (s *PaymentService) ConfirmPayment(ctx context.Context, paymentID, confirmedBy string) error {
	if _, _, err := s.repo.ConfirmPaymentAndRecomputeOrderStatus(ctx, paymentID, confirmedBy); err != nil {
		if errors.Is(err, orderrepo.ErrPaymentStateMismatch) {
			return fmt.Errorf("payment cannot be confirmed: not pending or already processed")
		}
		return err
	}
	return nil
}

// RefundPayment marks a payment as refunded and updates order status atomically (H-13).
func (s *PaymentService) RefundPayment(ctx context.Context, paymentID string) error {
	if _, _, err := s.repo.MarkPaymentRefundedAndRecomputeOrderStatus(ctx, paymentID); err != nil {
		if errors.Is(err, orderrepo.ErrPaymentStateMismatch) {
			return fmt.Errorf("payment cannot be refunded: not in confirmed state or concurrent update")
		}
		return err
	}
	return nil
}

// CountConfirmedPaymentsForOrder reports how many payments on the order are still
// in "confirmed" state. Used by cancel flows so admins can be warned when an order
// is being cancelled while money has already been received (H-22).
func (s *PaymentService) CountConfirmedPaymentsForOrder(ctx context.Context, orderID string) (int, float64, error) {
	payments, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return 0, 0, err
	}
	count := 0
	total := 0.0
	for _, p := range payments {
		if p.Status == "confirmed" {
			count++
			total += p.Amount
		}
	}
	return count, total, nil
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

// HasLivePaymentForOrder reports whether the order still has a payment in
// pending / authorized / confirmed state. Used by the compensation hook
// (H-21) so we only release stock for orders that no longer have any
// in-flight money.
func (s *PaymentService) HasLivePaymentForOrder(ctx context.Context, orderID string) (bool, error) {
	payments, err := s.repo.FindByOrderID(ctx, orderID)
	if err != nil {
		return false, err
	}
	for _, p := range payments {
		switch p.Status {
		case modelsOrder.PaymentRecordStatusPending,
			modelsOrder.PaymentRecordStatusAuthorized,
			modelsOrder.PaymentRecordStatusConfirmed:
			return true, nil
		}
	}
	return false, nil
}

// ConfirmAuthorizedPayment confirms a previously authorized gateway payment.
// Atomic with the parent order's payment_status recomputation (H-13).
func (s *PaymentService) ConfirmAuthorizedPayment(ctx context.Context, paymentID, confirmedBy string, capturedAmount float64) error {
	if _, _, err := s.repo.ConfirmAuthorizedPaymentAndRecomputeOrderStatus(ctx, paymentID, confirmedBy, capturedAmount); err != nil {
		if errors.Is(err, orderrepo.ErrPaymentStateMismatch) {
			return fmt.Errorf("payment cannot be confirmed: not in authorized state or concurrent update")
		}
		return err
	}
	return nil
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
