package common

import "time"

// NotificationOutbox status constants.
const (
	NotificationOutboxStatusPending   = "pending"
	NotificationOutboxStatusProcessed = "processed"
	NotificationOutboxStatusFailed    = "failed"
)

// NotificationOutbox channel constants.
const (
	NotificationChannelEmail = "email"
	NotificationChannelSMS   = "sms"
)

// NotificationOutbox is a transactional outbox row used to dispatch external
// notifications (email, SMS) without coupling the originating DB transaction
// to the network call (C-8).
//
// Producers insert a row inside the same transaction that mutates business
// state. A separate relay drains pending rows and calls the appropriate
// adapter; failures are retried with exponential backoff.
type NotificationOutbox struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Channel     string    `gorm:"type:varchar(20);not null;index" json:"channel"`
	Recipient   string    `gorm:"type:varchar(320);not null" json:"recipient"`
	Subject     string    `gorm:"type:varchar(500)" json:"subject"`
	Body        string    `gorm:"type:text;not null" json:"body"`
	Status      string    `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	Attempts    int       `gorm:"default:0" json:"attempts"`
	LastError   string    `gorm:"type:text" json:"lastError,omitempty"`
	NextAttempt time.Time `gorm:"index" json:"nextAttempt"`
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
}
