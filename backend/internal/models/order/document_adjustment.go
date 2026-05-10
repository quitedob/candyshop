package order

import (
	"time"

	"gorm.io/datatypes"
)

// 单证类型：订单 / 发票 / 贸易主单（审计用）
const (
	DocAdjustmentDocTypeOrder            = "order"
	DocAdjustmentDocTypeInvoice          = "invoice"
	DocAdjustmentDocTypeTradeTransaction = "trade_transaction"
	DocAdjustmentDocTypeTradeDocument    = "trade_document"
)

// 调整动作枚举
const (
	DocAdjustmentActionDerivedFromOrder     = "derived_from_order"
	DocAdjustmentActionManualInvoiceEdit    = "manual_invoice_edit"
	DocAdjustmentActionOrderFinancialChange = "order_financial_change"
	DocAdjustmentActionTradeSyncedFromOrder = "trade_synced_from_order"
)

// DocumentAdjustment 记录财务/行项目偏离「订单单一事实来源」时的审计（需人工填写原因）
type DocumentAdjustment struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	DocType        string         `gorm:"type:varchar(40);not null;index" json:"docType"`
	DocumentID     string         `gorm:"type:varchar(100);not null;index" json:"documentId"`
	RelatedOrderID string         `gorm:"type:varchar(100);index" json:"relatedOrderId,omitempty"`
	ActorUserID    string         `gorm:"type:varchar(100);index" json:"actorUserId"`
	Action         string         `gorm:"type:varchar(80);not null" json:"action"`
	Reason         string         `gorm:"type:text;not null" json:"reason"`
	BeforeSnapshot datatypes.JSON `gorm:"type:jsonb" json:"beforeSnapshot,omitempty"`
	AfterSnapshot  datatypes.JSON `gorm:"type:jsonb" json:"afterSnapshot,omitempty"`
	CreatedAt      time.Time      `gorm:"index" json:"createdAt"`
}
