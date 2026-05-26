package order

import "time"

// StockTransactionReason describes why the stock changed.
const (
	StockReasonOrderCreated     = "order_created"
	StockReasonOrderConfirmed   = "order_confirmed"
	StockReasonOrderCancelled   = "order_cancelled"
	StockReasonOrderDeleted     = "order_deleted"
	StockReasonOrderUpdated     = "order_updated"
	StockReasonDraftExpired     = "draft_expired"
	StockReasonManualAdjustment = "manual_adjustment"
	StockReasonDispatched       = "dispatched"
	StockReasonGoodsIssued      = "goods_issued"
	StockReasonGoodsReceived    = "goods_received"
	StockReasonStockReserved    = "stock_reserved"
	StockReasonStockReleased    = "stock_released"
	StockReasonStockDeducted    = "stock_deducted"
	StockReasonStockTransfer    = "stock_transfer"
)

// StockTransaction records every stock quantity change for audit and traceability.
type StockTransaction struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProductID   string    `gorm:"type:varchar(100);index;not null" json:"productId"`
	Change      int       `gorm:"not null" json:"change"` // negative = deduction, positive = restoration
	StockBefore int       `gorm:"not null" json:"stockBefore"`
	StockAfter  int       `gorm:"not null" json:"stockAfter"`
	Reason      string    `gorm:"type:varchar(50);not null" json:"reason"`
	ReferenceID string    `gorm:"type:varchar(100);index" json:"referenceId,omitempty"` // order ID or manual ref
	OperatorID  string    `gorm:"type:varchar(100)" json:"operatorId,omitempty"`        // user or system
	BatchID     *string   `gorm:"type:varchar(100);index" json:"batchId,omitempty"`     // ProductBatch.id（FEFO 扣减）
	LotNumber   string    `gorm:"type:varchar(100)" json:"lotNumber,omitempty"`         // 批次号 / 溯源
	WarehouseID *string   `gorm:"type:varchar(100);index" json:"warehouseId,omitempty"` // 多仓扣减时记录仓库
	CreatedAt   time.Time `gorm:"index" json:"createdAt"`
}
