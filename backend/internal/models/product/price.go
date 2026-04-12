package product

import "time"

// PriceList represents a named pricing tier (e.g., "Standard", "VIP Wholesale").
type PriceList struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Currency    string    `json:"currency" gorm:"default:'USD'"`
	Status      string    `json:"status" gorm:"default:'active'"` // active, inactive
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PriceRule maps a product to a price list with optional volume tiers.
type PriceRule struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	ProductID   string    `json:"productId" gorm:"index;not null"`
	PriceListID string    `json:"priceListId" gorm:"index;not null"`
	MinQuantity int       `json:"minQuantity" gorm:"default:1"`
	UnitPrice   float64   `json:"unitPrice" gorm:"not null"`
	Currency    string    `json:"currency" gorm:"default:'USD'"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
