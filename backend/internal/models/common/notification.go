package common

import "time"

// Notification represents a persistent notification for a user.
type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"type:varchar(100);not null;index" json:"userId"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"` // order, inquiry, trade, system
	Reference string    `gorm:"type:varchar(100)" json:"reference"`    // related entity ID
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Message   string    `gorm:"type:text" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
