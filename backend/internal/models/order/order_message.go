package order

import (
	modelsCommon "candypro/api/internal/models/common"
	"time"
)

// OrderMessage represents a message in an order conversation thread between customer and admin.
type OrderMessage struct {
	ID          string                    `json:"id" gorm:"primaryKey"`
	OrderID     string                    `json:"orderId" gorm:"not null;index"`
	UserID      string                    `json:"userId" gorm:"not null;index"`
	SenderType  string                    `json:"senderType" gorm:"not null;default:'customer'"` // "customer" or "admin"
	Message     string                    `json:"message" gorm:"type:text;not null"`
	Attachments modelsCommon.StringArray  `json:"attachments" gorm:"type:jsonb"`
	CreatedAt   time.Time                 `json:"createdAt"`
	UpdatedAt   time.Time                 `json:"updatedAt"`
}
