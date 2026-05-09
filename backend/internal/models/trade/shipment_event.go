package trade

import "time"

// Shipment event type constants.
const (
	EventDispatched     = "dispatched"
	EventPickedUp       = "picked_up"
	EventInTransit      = "in_transit"
	EventArrivedAtPort  = "arrived_at_port"
	EventCustomsCleared = "customs_cleared"
	EventOutForDelivery = "out_for_delivery"
	EventDelivered      = "delivered"
)

// ShipmentEvent represents a single timestamped tracking event on a shipment.
type ShipmentEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ShipmentID  uint      `gorm:"not null;index" json:"shipmentId"`
	EventType   string    `gorm:"type:varchar(50);not null" json:"eventType"`
	Location    string    `gorm:"type:varchar(255)" json:"location,omitempty"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	EventTime   time.Time `gorm:"not null;index" json:"eventTime"`
	OperatorID  string    `gorm:"type:varchar(100)" json:"operatorId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}
