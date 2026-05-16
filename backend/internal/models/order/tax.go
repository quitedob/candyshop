package order

import "time"

// TaxRate defines a tax rate for a destination country/region.
type TaxRate struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Country     string    `json:"country" gorm:"not null;index"` // ISO country code
	Region      string    `json:"region"`                        // state/province (optional)
	Rate        float64   `json:"rate"`                          // e.g. 0.08 = 8%
	Name        string    `json:"name"`                          // e.g. "VAT", "GST", "Sales Tax"
	IsActive    bool      `json:"isActive" gorm:"default:true"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
