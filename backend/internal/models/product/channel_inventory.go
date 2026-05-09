package product

import "time"

// ChannelWebstore 自建站渠道代码（OMS 可售量视图默认键）
const ChannelWebstore = "webstore"

// ChannelInventory 渠道库存与同步边界：用于平台店铺与 ERP 对齐，不替代 WMS 实物账
type ChannelInventory struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	ProductID          string     `gorm:"type:varchar(100);not null;uniqueIndex:idx_ch_prod_channel" json:"productId"`
	ChannelCode        string     `gorm:"type:varchar(40);not null;uniqueIndex:idx_ch_prod_channel" json:"channelCode"`
	WarehouseID        *string    `gorm:"type:varchar(100);index" json:"warehouseId,omitempty"`
	ListedQuantity     int        `gorm:"default:0" json:"listedQuantity"`     // 渠道上架可售上限；0=不单独限制
	ReservedForChannel int        `gorm:"default:0" json:"reservedForChannel"` // 渠道内预留（不上架可售）
	ExternalSKU        string     `gorm:"type:varchar(120)" json:"externalSku,omitempty"`
	SyncStatus         string     `gorm:"type:varchar(40);default:'local'" json:"syncStatus"` // local, pending_push, synced, error
	LastSyncedAt       *time.Time `json:"lastSyncedAt,omitempty"`
	SyncNotes          string     `gorm:"type:text" json:"syncNotes,omitempty"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}
