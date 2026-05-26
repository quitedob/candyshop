package product

import "time"

// Warehouse represents a physical storage location.
type Warehouse struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Code      string    `json:"code" gorm:"uniqueIndex;not null"`
	Type      string    `json:"type" gorm:"type:varchar(40)"` // factory, transit, overseas, bonded
	Address   string    `json:"address" gorm:"type:text"`
	Country   string    `json:"country" gorm:"type:varchar(100)"`
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	IsDefault bool      `json:"isDefault" gorm:"default:false;index"` // 默认发货仓（单仓回退）
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// WarehouseStock represents stock level of a product in a specific warehouse.
type WarehouseStock struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	WarehouseID string    `json:"warehouseId" gorm:"not null;uniqueIndex:idx_wh_prod"`
	ProductID   string    `json:"productId" gorm:"not null;uniqueIndex:idx_wh_prod;index"`
	VariantID   *string   `json:"variantId" gorm:"index"`
	Quantity    int       `json:"quantity" gorm:"default:0"`
	Reserved    int       `json:"reserved" gorm:"default:0"` // reserved for pending orders
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProductBatch represents a production batch with traceability info.
type ProductBatch struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	ProductID      string    `json:"productId" gorm:"not null;index:idx_batch_fefo,priority:1"`
	VariantID      *string   `json:"variantId" gorm:"index"`
	WarehouseID    string    `json:"warehouseId" gorm:"not null;index:idx_batch_fefo,priority:2"`
	BatchNumber    string    `json:"batchNumber" gorm:"uniqueIndex;not null"`
	Quantity       int       `json:"quantity" gorm:"default:0"`
	UnitCost       float64   `json:"unitCost" gorm:"default:0"` // purchase unit cost
	ProductionDate time.Time `json:"productionDate"`
	ExpiryDate     time.Time `json:"expiryDate" gorm:"index:idx_batch_fefo,priority:3"`
	IsExpired      bool      `json:"isExpired" gorm:"default:false"`
	Notes          string    `json:"notes" gorm:"type:text"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
