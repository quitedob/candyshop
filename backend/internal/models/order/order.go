package order

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// OrderItem represents a single product in an order
type OrderItem struct {
	ProductID      string  `json:"productId"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	Specifications string  `json:"specifications,omitempty"`
}

// OrderItemArray is a custom type for storing OrderItem arrays in PostgreSQL as JSON
type OrderItemArray []OrderItem

// Value implements driver.Valuer interface
func (s OrderItemArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface
func (s *OrderItemArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan OrderItemArray: expected []byte")
	}
	return json.Unmarshal(bytes, s)
}

// Address represents a standard shipping/billing address
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zipCode"`
	Country string `json:"country"`
}

// Value implements driver.Valuer interface
func (a Address) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner interface
func (a *Address) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Address: expected []byte")
	}
	return json.Unmarshal(bytes, a)
}

// OrderUserSnapshot is a lightweight projection of user fields used by order views.
type OrderUserSnapshot struct {
	ID        string `json:"id" gorm:"primaryKey"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// TableName keeps snapshot bound to users table.
func (OrderUserSnapshot) TableName() string {
	return "users"
}

// OrderInquirySnapshot is a lightweight projection of inquiry fields used by order views.
type OrderInquirySnapshot struct {
	ID            string `json:"id" gorm:"primaryKey"`
	CompanyName   string `json:"companyName"`
	ContactPerson string `json:"contactPerson"`
	Email         string `json:"email"`
	Status        string `json:"status"`
}

// TableName keeps snapshot bound to inquiries table.
func (OrderInquirySnapshot) TableName() string {
	return "inquiries"
}

// Order represents a customer order
type Order struct {
	ID                         string                `json:"id" gorm:"primaryKey"`
	OrderNumber                string                `json:"orderNumber" gorm:"uniqueIndex;not null"`
	UserID                     string                `json:"userId" gorm:"not null;index"`
	User                       *OrderUserSnapshot    `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	InquiryID                  *string               `json:"inquiryId" gorm:"index"`
	Inquiry                    *OrderInquirySnapshot `json:"inquiry,omitempty" gorm:"foreignKey:InquiryID;references:ID"`
	Status                     string                `json:"status" gorm:"default:'pending'"`       // pending, confirmed, production, shipped, delivered, cancelled
	PaymentStatus              string                `json:"paymentStatus" gorm:"default:'unpaid'"` // unpaid, partial, paid, refunded
	Items                      OrderItemArray        `json:"items" gorm:"type:jsonb;not null"`
	StockReserved              bool                  `json:"stockReserved" gorm:"default:false"`
	ComplianceOfficialEvidence bool                  `json:"complianceOfficialEvidence" gorm:"default:false"`
	Subtotal                   float64               `json:"subtotal"`
	TaxAmount                  float64               `json:"taxAmount" gorm:"default:0"`
	ShippingAmount             float64               `json:"shippingAmount" gorm:"default:0"`
	TotalAmount                float64               `json:"totalAmount"`
	Currency                   string                `json:"currency" gorm:"default:'USD'"`
	ProductionStartDate        *time.Time            `json:"productionStartDate"`
	EstimatedCompletion        *time.Time            `json:"estimatedCompletion"`
	ActualCompletion           *time.Time            `json:"actualCompletion"`
	ShippingAddress            Address               `json:"shippingAddress" gorm:"type:jsonb"`
	TrackingNumber             string                `json:"trackingNumber"`
	ConfirmedAt                *time.Time            `json:"confirmedAt"`
	ShippedAt                  *time.Time            `json:"shippedAt"`
	DeliveredAt                *time.Time            `json:"deliveredAt"`
	CreatedAt                  time.Time             `json:"createdAt"`
	UpdatedAt                  time.Time             `json:"updatedAt"`
}
