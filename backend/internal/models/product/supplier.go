package product

import "time"

// Supplier represents a vendor or manufacturer that supplies products.
type Supplier struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	Name          string    `json:"name" gorm:"not null"`
	ContactPerson string    `json:"contactPerson"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Address       string    `json:"address" gorm:"type:text"`
	Country       string    `json:"country" gorm:"type:varchar(100)"`
	PaymentTerms  string    `json:"paymentTerms"` // Net 30, Net 60, etc.
	Rating        float64   `json:"rating" gorm:"default:0"`
	IsActive      bool      `json:"isActive" gorm:"default:true"`
	APIKey        string    `json:"apiKey,omitempty" gorm:"type:varchar(64)"` // self-service auth key; disclosed only to the owning supplier (registration/profile) and admins
	Notes         string    `json:"notes" gorm:"type:text"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// PurchaseOrderStatus values
const (
	POStatusDraft     = "draft"
	POStatusSent      = "sent"
	POStatusPartial   = "partially_received"
	POStatusReceived  = "received"
	POStatusCancelled = "cancelled"
)

// PurchaseOrder represents a procurement order to a supplier.
type PurchaseOrder struct {
	ID               string     `json:"id" gorm:"primaryKey"`
	PONumber         string     `json:"poNumber" gorm:"uniqueIndex;not null"`
	SupplierID       string     `json:"supplierId" gorm:"not null;index"`
	WarehouseID      string     `json:"warehouseId"`
	Status           string     `json:"status" gorm:"default:'draft'"`
	ExpectedDate     *time.Time `json:"expectedDate"`
	ReceivedDate     *time.Time `json:"receivedDate"`
	Notes            string     `json:"notes" gorm:"type:text"`
	CreatedBy        string     `json:"createdBy" gorm:"not null"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// PurchaseOrderItem represents a line item in a purchase order.
type PurchaseOrderItem struct {
	ID              uint    `json:"id" gorm:"primaryKey"`
	POID            string  `json:"poId" gorm:"not null;index"`
	ProductID       string  `json:"productId" gorm:"not null"`
	Quantity        int     `json:"quantity" gorm:"not null"`
	UnitCost        float64 `json:"unitCost"`
	ReceivedQty     int     `json:"receivedQty" gorm:"default:0"`
}
