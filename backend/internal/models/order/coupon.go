package order

import "time"

// CouponType values
const (
	CouponTypePercentage = "percentage"
	CouponTypeFixed      = "fixed"
)

// CouponStatus values
const (
	CouponStatusActive    = "active"
	CouponStatusInactive  = "inactive"
	CouponStatusExhausted = "exhausted"
	CouponStatusExpired   = "expired"
)

// Coupon represents a discount coupon that customers can apply to orders.
type Coupon struct {
	ID             string     `json:"id" gorm:"primaryKey"`
	Code           string     `json:"code" gorm:"uniqueIndex;not null"`
	Type           string     `json:"type" gorm:"not null"` // percentage, fixed
	Value          float64    `json:"value" gorm:"not null"`
	MinOrderAmount float64    `json:"minOrderAmount"`
	MaxUses        int        `json:"maxUses" gorm:"default:0"`
	UsedCount      int        `json:"usedCount" gorm:"default:0"`
	MaxUsesPerUser int        `json:"maxUsesPerUser" gorm:"default:1"`
	StartsAt       time.Time  `json:"startsAt"`
	ExpiresAt      time.Time  `json:"expiresAt"`
	Status         string     `json:"status" gorm:"default:'active'"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

// GiftCardStatus values
const (
	GiftCardStatusActive   = "active"
	GiftCardStatusInactive = "inactive"
	GiftCardStatusExhausted = "exhausted"
)

// GiftCard represents a prepaid store credit card.
type GiftCard struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	Code            string    `json:"code" gorm:"uniqueIndex;not null"`
	InitialBalance  float64   `json:"initialBalance" gorm:"not null"`
	CurrentBalance  float64   `json:"currentBalance" gorm:"not null"`
	Currency        string    `json:"currency" gorm:"default:'USD'"`
	ExpiresAt       *time.Time `json:"expiresAt"`
	Status          string    `json:"status" gorm:"default:'active'"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// OrderDiscount records a discount applied to an order.
type OrderDiscount struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	OrderID   string  `json:"orderId" gorm:"not null;index"`
	CouponID  *string `json:"couponId"`
	GiftCardID *string `json:"giftCardId"`
	Type      string  `json:"type" gorm:"not null"` // coupon, gift_card, manual
	Amount    float64 `json:"amount" gorm:"not null"`
	Label     string  `json:"label"`
}
