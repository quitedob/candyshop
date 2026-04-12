package product

import "time"

// Warehouse represents a physical storage location.
type Warehouse struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Code      string    `json:"code" gorm:"uniqueIndex;not null"`
	Address   string    `json:"address" gorm:"type:text"`
	Country   string    `json:"country" gorm:"type:varchar(100)"`
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// WarehouseStock represents stock level of a product in a specific warehouse.
type WarehouseStock struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	WarehouseID string    `json:"warehouseId" gorm:"not null;index"`
	ProductID   string    `json:"productId" gorm:"not null;index"`
	VariantID   *string   `json:"variantId" gorm:"index"`
	Quantity    int       `json:"quantity" gorm:"default:0"`
	Reserved    int       `json:"reserved" gorm:"default:0"` // reserved for pending orders
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProductBatch represents a production batch with traceability info.
type ProductBatch struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ProductID      string    `json:"productId" gorm:"not null;index"`
	VariantID      *string   `json:"variantId" gorm:"index"`
	WarehouseID    string    `json:"warehouseId" gorm:"not null;index"`
	BatchNumber    string    `json:"batchNumber" gorm:"uniqueIndex;not null"`
	Quantity       int       `json:"quantity" gorm:"default:0"`
	ProductionDate time.Time `json:"productionDate"`
	ExpiryDate     time.Time `json:"expiryDate"`
	IsExpired      bool      `json:"isExpired" gorm:"default:false"`
	Notes          string    `json:"notes" gorm:"type:text"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
