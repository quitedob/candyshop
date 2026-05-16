package order

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// WebhookEvent types
const (
	WebhookEventOrderCreated     = "order.created"
	WebhookEventOrderConfirmed   = "order.confirmed"
	WebhookEventOrderShipped     = "order.shipped"
	WebhookEventOrderDelivered   = "order.delivered"
	WebhookEventPaymentConfirmed = "payment.confirmed"
	WebhookEventReturnCreated    = "return.created"
)

// WebhookDelivery statuses
const (
	WebhookDeliveryPending = "pending"
	WebhookDeliverySuccess = "success"
	WebhookDeliveryFailed  = "failed"
)

// WebhookConfig holds a configured outbound webhook endpoint.
type WebhookConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"type:varchar(120);not null" json:"name"`
	URL       string    `gorm:"type:text;not null" json:"url"`
	Secret    string    `gorm:"type:varchar(255);not null" json:"-"`
	Events    JSONArray `gorm:"type:jsonb;not null" json:"events"`
	Status    string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// WebhookDelivery logs each outbound webhook attempt.
type WebhookDelivery struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	WebhookID    uint       `gorm:"not null;index" json:"webhookId"`
	EventType    string     `gorm:"type:varchar(80);not null;index" json:"eventType"`
	AggregateKey string     `gorm:"type:varchar(100);not null;index" json:"aggregateKey"`
	Payload      string     `gorm:"type:jsonb" json:"payload"`
	Status       string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ResponseCode *int       `json:"responseCode,omitempty"`
	ResponseBody string     `gorm:"type:text" json:"responseBody,omitempty"`
	Attempts     int        `gorm:"default:1" json:"attempts"`
	LastError    string     `gorm:"type:text" json:"lastError,omitempty"`
	DeliveredAt  *time.Time `json:"deliveredAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// JSONArray is a string slice stored as JSONB.
type JSONArray []string

func (a JSONArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *JSONArray) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, a)
}
