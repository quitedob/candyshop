package auth

import "time"

// PasswordResetToken decouples password reset state from the user row.
//
// M-22: tokens used to be of the form `<userID>.<random>`, leaking the user ID
// in the email link. With a dedicated table the token sent to the user is a
// pure random opaque string and the user ID is resolved server-side via an
// indexed lookup on TokenLookupKey (a non-reversible 16-char prefix of the
// token). The token's bcrypt hash is verified after the lookup so DB dumps
// remain useless without the original token.
type PasswordResetToken struct {
	ID             string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID         string    `gorm:"type:varchar(36);not null;index" json:"userId"`
	TokenLookupKey string    `gorm:"type:varchar(64);not null;uniqueIndex" json:"-"`
	TokenHash      string    `gorm:"type:varchar(255);not null" json:"-"`
	ExpiresAt      time.Time `gorm:"not null;index" json:"expiresAt"`
	UsedAt         *time.Time `json:"usedAt,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}
