package order

import "time"

// NegotiationOffer represents a counter-offer between buyer and seller during inquiry negotiation.
type NegotiationOffer struct {
	ID           string     `json:"id" gorm:"primaryKey"`
	InquiryID    string     `json:"inquiryId" gorm:"not null;index"`
	UserID       string     `json:"userId" gorm:"not null;index"`
	SenderType   string     `json:"senderType" gorm:"not null;default:'customer'"` // "customer" or "admin"
	Status       string     `json:"status" gorm:"index;default:'pending'"`          // pending, accepted, rejected, expired
	UnitPrice    float64    `json:"unitPrice"`
	Quantity     int        `json:"quantity"`
	TotalAmount  float64    `json:"totalAmount"`
	Currency     string     `json:"currency" gorm:"default:'USD'"`
	Incoterms    string     `json:"incoterms" gorm:"type:varchar(50)"`
	PaymentTerms string     `json:"paymentTerms" gorm:"type:varchar(255)"`
	DeliveryDate string     `json:"deliveryDate"`
	ValidUntil   *time.Time `json:"validUntil"`
	Message      string     `json:"message" gorm:"type:text"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}
