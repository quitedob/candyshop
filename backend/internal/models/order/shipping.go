package order

import "time"

// ShippingRate defines shipping cost by destination country and weight range.
type ShippingRate struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	Destination    string    `json:"destination" gorm:"not null;index"` // country code or region
	MinWeightKg    float64   `json:"minWeightKg"`
	MaxWeightKg    float64   `json:"maxWeightKg"`
	BaseCost       float64   `json:"baseCost"`
	CostPerKg      float64   `json:"costPerKg"`
	Currency       string    `json:"currency" gorm:"default:'USD'"`
	Carrier        string    `json:"carrier" gorm:"type:varchar(100)"`  // sea, air, express
	EstimatedDays  int       `json:"estimatedDays"`
	IsActive       bool      `json:"isActive" gorm:"default:true"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
