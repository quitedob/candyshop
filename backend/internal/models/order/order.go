package order

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Order Statuses
const (
	OrderStatusPending            = "pending"
	OrderStatusPendingConfirm     = "pending_confirmation"
	OrderStatusPendingApproval    = "pending_approval"
	OrderStatusConfirmed          = "confirmed"
	OrderStatusProduction         = "production"
	OrderStatusShipped            = "shipped"
	OrderStatusPartiallyShipped   = "partially_shipped"
	OrderStatusDelivered          = "delivered"
	OrderStatusPartiallyDelivered = "partially_delivered"
	OrderStatusPartiallyReturned  = "partially_returned"
	OrderStatusReturned           = "returned"
	OrderStatusCancelled          = "cancelled"
	OrderStatusExpired            = "expired"
)

// Payment Statuses
const (
	PaymentStatusUnpaid   = "unpaid"
	PaymentStatusPartial  = "partial"
	PaymentStatusPaid     = "paid"
	PaymentStatusRefunded = "refunded"
)

// ValidOrderStatusTransitions defines the allowed order status flow.
var ValidOrderStatusTransitions = map[string]map[string]bool{
	OrderStatusPending:            {OrderStatusConfirmed: true, OrderStatusCancelled: true},
	OrderStatusPendingConfirm:     {OrderStatusPending: true, OrderStatusConfirmed: true, OrderStatusCancelled: true, OrderStatusExpired: true},
	OrderStatusPendingApproval:    {OrderStatusPendingConfirm: true, OrderStatusCancelled: true},
	OrderStatusConfirmed:          {OrderStatusProduction: true, OrderStatusCancelled: true},
	OrderStatusProduction:         {OrderStatusShipped: true, OrderStatusPartiallyShipped: true, OrderStatusCancelled: true},
	OrderStatusPartiallyShipped:   {OrderStatusShipped: true, OrderStatusPartiallyDelivered: true, OrderStatusDelivered: true, OrderStatusPartiallyReturned: true, OrderStatusReturned: true},
	OrderStatusShipped:            {OrderStatusDelivered: true, OrderStatusPartiallyDelivered: true, OrderStatusPartiallyReturned: true, OrderStatusReturned: true},
	OrderStatusPartiallyDelivered: {OrderStatusDelivered: true, OrderStatusPartiallyReturned: true, OrderStatusReturned: true},
	OrderStatusDelivered:          {OrderStatusReturned: true, OrderStatusPartiallyReturned: true},
	// partially_returned must be able to escalate to fully returned — otherwise
	// orders that started with a partial return get stuck and can never be
	// finalized, blocking COGS reversal and refund accounting (R2 A-6).
	OrderStatusPartiallyReturned:  {OrderStatusReturned: true},
	OrderStatusReturned:           {},
	OrderStatusCancelled:          {},
	OrderStatusExpired:            {},
}

// ValidateOrderStatusTransition checks whether moving from current to target is allowed.
func ValidateOrderStatusTransition(current, target string) error {
	cur := strings.ToLower(strings.TrimSpace(current))
	tgt := strings.ToLower(strings.TrimSpace(target))
	if cur == tgt {
		return nil
	}
	allowed, known := ValidOrderStatusTransitions[cur]
	if !known {
		return fmt.Errorf("unknown current order status '%s'", current)
	}
	if !allowed[tgt] {
		return fmt.Errorf("cannot transition order from '%s' to '%s'", current, target)
	}
	return nil
}

// OrderItem represents a single product in an order
type OrderItem struct {
	ProductID         string  `json:"productId"`
	Quantity          int     `json:"quantity"`
	UnitPrice         float64 `json:"unitPrice"`
	FulfilledQuantity int     `json:"fulfilledQuantity,omitempty"`
	ShippedQuantity   int     `json:"shippedQuantity,omitempty"`
	Specifications    string  `json:"specifications,omitempty"`
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

// UnmarshalJSON accepts zipCode as well as common aliases zip / postalCode.
func (a *Address) UnmarshalJSON(data []byte) error {
	type addressAlias Address
	var raw struct {
		addressAlias
		Zip        string `json:"zip"`
		PostalCode string `json:"postalCode"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	*a = Address(raw.addressAlias)
	if strings.TrimSpace(a.ZipCode) == "" {
		if z := strings.TrimSpace(raw.Zip); z != "" {
			a.ZipCode = z
		} else if z := strings.TrimSpace(raw.PostalCode); z != "" {
			a.ZipCode = z
		}
	}
	return nil
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

// Order Source values — 订单创建来源
const (
	OrderSourceCart     = "cart"
	OrderSourceAIAssist = "ai_assist"
	OrderSourceBulk     = "bulk"
	OrderSourceInquiry  = "inquiry"
)

// Order represents a customer order
type Order struct {
	ID                         string                `json:"id" gorm:"primaryKey"`
	OrderNumber                string                `json:"orderNumber" gorm:"uniqueIndex;not null"`
	UserID                     string                `json:"userId" gorm:"not null;index;index:idx_orders_user_status,priority:1"`
	User                       *OrderUserSnapshot    `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	InquiryID                  *string               `json:"inquiryId" gorm:"index"`
	Inquiry                    *OrderInquirySnapshot `json:"inquiry,omitempty" gorm:"foreignKey:InquiryID;references:ID"`
	Source                     string                `json:"source" gorm:"type:varchar(32);default:'';index"`
	Status                     string                `json:"status" gorm:"default:'pending';index;index:idx_orders_user_status,priority:2"`
	// PaymentStatus indexed for financial dashboards / payment-status filters
	// that previously triggered full table scans (R2 E-7).
	PaymentStatus              string                `json:"paymentStatus" gorm:"default:'unpaid';index"`
	Items                      OrderItemArray        `json:"items" gorm:"type:jsonb;not null"`
	WarehouseID                *string               `json:"warehouseId"`
	StockReserved              bool                  `json:"stockReserved" gorm:"default:false"`
	ComplianceOfficialEvidence bool                  `json:"complianceOfficialEvidence" gorm:"default:false"`
	Subtotal                   float64               `json:"subtotal"`
	TaxAmount                  float64               `json:"taxAmount" gorm:"default:0"`
	ShippingAmount             float64               `json:"shippingAmount" gorm:"default:0"`
	COGS                       float64               `json:"cogs" gorm:"default:0"`
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
	CreatedAt                  time.Time             `json:"createdAt" gorm:"index"`
	UpdatedAt                  time.Time             `json:"updatedAt"`
	// Version supports optimistic concurrency. GORM increments this on Updates(map)
	// or via the gorm.io/plugin/optimisticlock plugin. We hand-roll the conflict
	// guard in repository helpers (see UpdateWithVersionGuard) so non-state writes
	// can't silently overwrite each other (C-5).
	Version int64 `json:"version" gorm:"default:0"`
	// DeletedAt enables GORM soft-delete on orders. Cancellation already covers
	// the user-facing "remove" semantics; DeletedAt is reserved for compliance/
	// retention scrubbing where the row must be hidden but auditable history
	// preserved (C-9).
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
