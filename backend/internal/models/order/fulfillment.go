package order

import "time"

// FulfillmentStatus values
const (
	FulfillmentStatusPending   = "pending"
	FulfillmentStatusPicked    = "picked"
	FulfillmentStatusPacked    = "packed"
	FulfillmentStatusShipped   = "shipped"
	FulfillmentStatusDelivered = "delivered"
	FulfillmentStatusCancelled = "cancelled"
)

// Fulfillment represents a shipment of one or more order items from a warehouse.
type Fulfillment struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OrderID        string    `json:"orderId" gorm:"not null;index"`
	WarehouseID    string    `json:"warehouseId" gorm:"not null;index"`
	TrackingNumber string    `json:"trackingNumber"`
	Carrier        string    `json:"carrier"`
	Status         string    `json:"status" gorm:"default:'pending'"`
	Notes          string    `json:"notes" gorm:"type:text"`
	ShippedAt      *time.Time `json:"shippedAt"`
	DeliveredAt    *time.Time `json:"deliveredAt"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// FulfillmentItem represents a single line item within a fulfillment.
type FulfillmentItem struct {
	ID            uint    `json:"id" gorm:"primaryKey"`
	FulfillmentID string  `json:"fulfillmentId" gorm:"not null;index"`
	OrderItemIdx  int     `json:"orderItemIdx" gorm:"not null"` // index into Order.Items array
	ProductID     string  `json:"productId" gorm:"not null"`
	Quantity      int     `json:"quantity" gorm:"not null"`
}
