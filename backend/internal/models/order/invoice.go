package order

import "time"

// Invoice type constants.
const (
	InvoiceTypeProforma   = "proforma"
	InvoiceTypeCommercial = "commercial"
	InvoiceTypeCreditNote = "credit_note"
)

// Invoice status constants.
const (
	InvoiceStatusDraft   = "draft"
	InvoiceStatusSent    = "sent"
	InvoiceStatusPaid    = "paid"
	InvoiceStatusOverdue = "overdue"
	InvoiceStatusVoided  = "voided"
)

// Invoice represents a billing document linked to an order.
type Invoice struct {
	ID          string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	OrderID     string     `gorm:"type:varchar(255);index" json:"orderId"`
	Type        string     `gorm:"type:varchar(50);default:'commercial'" json:"type"`
	Status      string     `gorm:"type:varchar(50);default:'draft'" json:"status"`
	InvoiceNo   string     `gorm:"type:varchar(100);uniqueIndex" json:"invoiceNo"`
	Amount      float64    `json:"amount"`
	TaxAmount   float64    `json:"taxAmount"`
	TotalAmount float64    `json:"totalAmount"`
	Currency    string     `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	DueDate     *time.Time `json:"dueDate"`
	SentAt      *time.Time `json:"sentAt"`
	PaidAt      *time.Time `json:"paidAt"`
	Notes       string     `gorm:"type:text" json:"notes"`
	// JSON blob for line items: [{productId, name, qty, unitPrice, total}]
	Items     string    `gorm:"type:text" json:"items"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
