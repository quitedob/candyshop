package order

import "time"

// Payment represents a payment record for an order.
type Payment struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	OrderID     string     `json:"orderId" gorm:"index;not null"`
	Amount      float64    `json:"amount" gorm:"not null"`
	Currency    string     `json:"currency" gorm:"default:'USD'"`
	Method      string     `json:"method"`                                                         // bank_transfer, credit_card, letter_of_credit, wire
	Status      string     `json:"status" gorm:"default:'pending'"`                                // pending, confirmed, failed, refunded
	Reference   string     `json:"reference"`
	ProofURL    string     `json:"proofUrl"`
	Notes       string     `json:"notes"`
	ConfirmedBy *string    `json:"confirmedBy" gorm:"index"`
	ConfirmedAt *time.Time `json:"confirmedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
