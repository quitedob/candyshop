package auth

import ()

import (
	"time"
)

// RefreshToken represents a user's JWT refresh token
type RefreshToken struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	UserID    string     `json:"userId" gorm:"not null;index"`
	Token     string     `json:"token" gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time  `json:"expiresAt" gorm:"not null"`
	RevokedAt *time.Time `json:"revokedAt"`
	RevokedBy *string    `json:"revokedBy"`
	IPAddress string     `json:"ipAddress"`
	UserAgent string     `json:"userAgent"`
	CreatedAt time.Time  `json:"createdAt"`
}
