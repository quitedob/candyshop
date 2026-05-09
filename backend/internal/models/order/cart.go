package order

import "time"

// CartItem represents a single product line in a user's shopping cart.
type CartItem struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	UserID         string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_cart_user_product" json:"userId"`
	ProductID      string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_cart_user_product" json:"productId"`
	ProductName    string    `gorm:"type:varchar(255)" json:"productName"`
	Quantity       int       `gorm:"not null;default:1" json:"quantity"`
	UnitPrice      float64   `json:"unitPrice"`
	Currency       string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	Specifications string    `gorm:"type:text" json:"specifications"` // JSON or free-text OEM specs
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
