package common

import ()

import (
	"time"
)

// ActivityLog represents an audit log entry for user actions
type ActivityLog struct {
	ID         string    `json:"id" gorm:"primaryKey"`
	UserID     *string   `json:"userId" gorm:"index"`
	Action     string    `json:"action" gorm:"not null"` // login, update_inquiry, etc.
	EntityType string    `json:"entityType"`             // user, inquiry, product, order
	EntityID   string    `json:"entityId"`
	Details    string    `json:"details" gorm:"type:jsonb"`
	IPAddress  string    `json:"ipAddress"`
	UserAgent  string    `json:"userAgent"`
	CreatedAt  time.Time `json:"createdAt"`
}
