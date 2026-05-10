package order

import (
	"time"

	"gorm.io/datatypes"
)

// 发件箱事件类型：订单管理员确认为 confirmed 后异步创建贸易流水
const (
	OutboxEventOrderConfirmedCreateTrade = "order_confirmed_create_trade"
)

// 发件箱处理状态
const (
	OutboxStatusPending   = "pending"
	OutboxStatusProcessed = "processed"
	OutboxStatusFailed    = "failed"
)

// EventOutbox 事务性发件箱表，与订单状态更新同事务写入，由 Relay 异步消费
type EventOutbox struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	EventType    string         `gorm:"type:varchar(80);not null;uniqueIndex:idx_outbox_evt_agg" json:"eventType"`
	AggregateKey string         `gorm:"type:varchar(100);not null;uniqueIndex:idx_outbox_evt_agg" json:"aggregateKey"`
	Payload      datatypes.JSON `gorm:"type:jsonb" json:"payload"`
	Status       string         `gorm:"type:varchar(20);not null;default:'pending';index" json:"status"`
	Attempts     int            `gorm:"default:0" json:"attempts"`
	LastError    string         `gorm:"type:text" json:"lastError,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	ProcessedAt  *time.Time     `json:"processedAt,omitempty"`
}
