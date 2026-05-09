package product

import "time"

// OEMProjectInventoryHold OEM 项目对成品库存的预留（与普通订单可售隔离）
type OEMProjectInventoryHold struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProjectID string    `gorm:"type:varchar(100);not null;index" json:"projectId"`
	ProductID string    `gorm:"type:varchar(100);not null;index" json:"productId"`
	Quantity  int       `gorm:"not null" json:"quantity"`
	Status    string    `gorm:"type:varchar(30);default:'active';index" json:"status"` // active, released
	Notes     string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
