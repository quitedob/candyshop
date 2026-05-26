package order

import "time"

// ReturnStatus values
const (
	ReturnStatusPending  = "pending"
	ReturnStatusApproved = "approved"
	ReturnStatusReceived = "received"
	ReturnStatusRefunded = "refunded"
	ReturnStatusRejected = "rejected"
)

// ReturnReason constants
const (
	ReturnReasonDamaged        = "damaged"
	ReturnReasonWrongItem      = "wrong_item"
	ReturnReasonDefective      = "defective"
	ReturnReasonNotAsDescribed = "not_as_described"
	ReturnReasonExpired        = "expired"
	ReturnReasonOther          = "other"
)

// ReturnRequest represents a customer's request to return items from an order.
type ReturnRequest struct {
	ID             string     `json:"id" gorm:"primaryKey"`
	OrderID        string     `json:"orderId" gorm:"not null;index"`
	UserID         string     `json:"userId" gorm:"not null;index:idx_return_user_status,priority:1"`
	Status         string     `json:"status" gorm:"default:'pending';index:idx_return_user_status,priority:2"`
	Reason         string     `json:"reason" gorm:"not null"`
	Notes          string     `json:"notes" gorm:"type:text"`
	ApprovedBy     *string    `json:"approvedBy"`
	ApprovedAt     *time.Time `json:"approvedAt"`
	ReceivedBy     *string    `json:"receivedBy"`
	ReceivedAt     *time.Time `json:"receivedAt"`
	RefundedBy     *string    `json:"refundedBy"`
	RefundedAt     *time.Time `json:"refundedAt"`
	RejectedBy     *string    `json:"rejectedBy"`
	RejectedAt     *time.Time `json:"rejectedAt"`
	RejectReason   string     `json:"rejectReason"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	// Items is a has-many association used by repository preload paths so list
	// views can render line counts/products without a follow-up query per row (H-6).
	Items []ReturnItem `json:"items,omitempty" gorm:"foreignKey:ReturnID;references:ID"`
}

// ReturnItem represents a single line item within a return request.
type ReturnItem struct {
	ID             uint    `json:"id" gorm:"primaryKey"`
	ReturnID       string  `json:"returnId" gorm:"not null;index"`
	OrderItemIdx   int     `json:"orderItemIdx" gorm:"not null"`
	ProductID      string  `json:"productId" gorm:"not null"`
	Quantity       int     `json:"quantity" gorm:"not null"`
	ReasonCode     string  `json:"reasonCode" gorm:"not null"`
	Condition      string  `json:"condition" gorm:"default:'opened'"`
	RefundAmount   float64 `json:"refundAmount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
