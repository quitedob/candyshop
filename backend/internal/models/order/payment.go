package order

import (
	"fmt"
	"strings"
	"time"
)

// Payment record statuses
const (
	PaymentRecordStatusPending   = "pending"
	PaymentRecordStatusConfirmed = "confirmed"
	PaymentRecordStatusFailed    = "failed"
	PaymentRecordStatusRefunded  = "refunded"
)

// validPaymentStatusTransitions defines allowed payment status transitions.
var validPaymentStatusTransitions = map[string]map[string]bool{
	PaymentRecordStatusPending:   {PaymentRecordStatusConfirmed: true, PaymentRecordStatusFailed: true},
	PaymentRecordStatusConfirmed: {PaymentRecordStatusRefunded: true},
	PaymentRecordStatusFailed:    {},
	PaymentRecordStatusRefunded:  {},
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
	Method      string     `json:"method"` // bank_transfer, credit_card, letter_of_credit, wire
	Status      string     `json:"status" gorm:"default:'pending'"`
	Reference   string     `json:"reference"`
	ProofURL    string     `json:"proofUrl"`
	Notes       string     `json:"notes"`
	ConfirmedBy *string    `json:"confirmedBy" gorm:"index"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
