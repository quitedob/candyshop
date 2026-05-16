package order

import "time"

// StockTransferStatus values
const (
	StockTransferStatusPending   = "pending"
	StockTransferStatusInTransit = "in_transit"
	StockTransferStatusCompleted = "completed"
	StockTransferStatusCancelled = "cancelled"
)

// StockTransfer represents a movement of stock from one warehouse to another.
type StockTransfer struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	TransferNumber  string    `json:"transferNumber" gorm:"uniqueIndex;not null"`
	FromWarehouseID string    `json:"fromWarehouseId" gorm:"not null;index"`
	ToWarehouseID   string    `json:"toWarehouseId" gorm:"not null;index"`
	Status          string    `json:"status" gorm:"default:'pending'"`
	Notes           string    `json:"notes" gorm:"type:text"`
	CreatedBy       string    `json:"createdBy" gorm:"not null"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// StockTransferItem represents a single product line in a stock transfer.
type StockTransferItem struct {
	ID              uint      `json:"id" gorm:"primaryKey"`
	TransferID      string    `json:"transferId" gorm:"not null;index"`
	ProductID       string    `json:"productId" gorm:"not null"`
	BatchID         *string   `json:"batchId"`
	Quantity        int       `json:"quantity" gorm:"not null"`
	ReceivedQty     int       `json:"receivedQty" gorm:"default:0"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
