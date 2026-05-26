package order

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Payment method constants (enum-ish)
const (
	PaymentMethodBankTransfer   = "bank_transfer"
	PaymentMethodCreditCard     = "credit_card"
	PaymentMethodWire           = "wire"
	PaymentMethodLetterOfCredit = "letter_of_credit"
	PaymentMethodStripe         = "stripe"
	PaymentMethodPayPal         = "paypal"
)

// ValidPaymentMethods lists all recognized payment methods.
var ValidPaymentMethods = map[string]bool{
	PaymentMethodBankTransfer:   true,
	PaymentMethodCreditCard:     true,
	PaymentMethodWire:           true,
	PaymentMethodLetterOfCredit: true,
	PaymentMethodStripe:         true,
	PaymentMethodPayPal:         true,
}

// Payment record statuses
const (
	PaymentRecordStatusPending     = "pending"
	PaymentRecordStatusAuthorized  = "authorized"
	PaymentRecordStatusConfirmed   = "confirmed"
	PaymentRecordStatusFailed      = "failed"
	PaymentRecordStatusRefunded    = "refunded"
	PaymentRecordStatusPartialRefund = "partial_refund"
)

// validPaymentStatusTransitions defines allowed payment status transitions.
var validPaymentStatusTransitions = map[string]map[string]bool{
	PaymentRecordStatusPending:       {PaymentRecordStatusAuthorized: true, PaymentRecordStatusConfirmed: true, PaymentRecordStatusFailed: true},
	PaymentRecordStatusAuthorized:    {PaymentRecordStatusConfirmed: true, PaymentRecordStatusFailed: true},
	PaymentRecordStatusConfirmed:     {PaymentRecordStatusRefunded: true, PaymentRecordStatusPartialRefund: true},
	PaymentRecordStatusFailed:        {},
	PaymentRecordStatusRefunded:      {},
	PaymentRecordStatusPartialRefund: {PaymentRecordStatusRefunded: true},
}

// ValidatePaymentStatusTransition checks whether moving from current to target is allowed.
func ValidatePaymentStatusTransition(current, target string) error {
	cur := strings.ToLower(strings.TrimSpace(current))
	tgt := strings.ToLower(strings.TrimSpace(target))
	if cur == tgt {
		return nil
	}
	allowed, known := validPaymentStatusTransitions[cur]
	if !known {
		return fmt.Errorf("unknown current payment status '%s'", current)
	}
	if !allowed[tgt] {
		return fmt.Errorf("cannot transition payment from '%s' to '%s'", current, target)
	}
	return nil
}

// Payment represents a payment record for an order.
type Payment struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	OrderID     string     `json:"orderId" gorm:"index;not null"`
	Amount      float64    `json:"amount" gorm:"not null"`
	Currency    string     `json:"currency" gorm:"default:'USD'"`
	Method      string     `json:"method"` // bank_transfer, credit_card, letter_of_credit, wire, stripe, paypal
	Status      string     `json:"status" gorm:"default:'pending';index"`
	Reference   string     `json:"reference"`
	ProofURL    string     `json:"proofUrl"`
	Notes       string     `json:"notes"`
	ConfirmedBy *string    `json:"confirmedBy" gorm:"index"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	// Gateway fields for processor-linked payments (Stripe, PayPal, etc.)
	GatewayTransactionID *string `json:"gatewayTransactionId,omitempty" gorm:"index"`
	AuthorizedAmount     float64 `json:"authorizedAmount"`
	CapturedAmount       float64 `json:"capturedAmount"`
	RefundedAmount       float64 `json:"refundedAmount"`
	GatewayMetadata      string  `json:"gatewayMetadata,omitempty"` // JSONB: raw gateway response
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	// Version is the optimistic-lock counter (C-5).
	Version int64 `json:"version" gorm:"default:0"`
	// DeletedAt enables soft delete on payments (C-9). Refunded payments are
	// kept active; only operationally void rows are soft-deleted.
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
