package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"time"

	"gorm.io/gorm"
)

// UpdateWithOutbox 在同一数据库事务内更新订单并写入发件箱（Outbox 模式）
func (r *OrderRepository) UpdateWithOutbox(ctx context.Context, order *modelsOrder.Order, outbox *modelsOrder.EventOutbox) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("User", "Inquiry").Save(order).Error; err != nil {
			return err
		}
		if outbox != nil {
			if err := tx.Create(outbox).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListPendingOutbox 拉取待处理发件箱记录供 Relay 消费
func (r *OrderRepository) ListPendingOutbox(ctx context.Context, eventType string, limit int) ([]modelsOrder.EventOutbox, error) {
	var rows []modelsOrder.EventOutbox
	q := r.db.WithContext(ctx).
		Where("status = ? AND event_type = ?", modelsOrder.OutboxStatusPending, eventType).
		Order("id ASC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// IncrementOutboxAttempt 记录消费重试次数
func (r *OrderRepository) IncrementOutboxAttempt(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&modelsOrder.EventOutbox{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + ?", 1)).Error
}

// UpdateOutboxResult 标记发件箱处理结果
func (r *OrderRepository) UpdateOutboxResult(ctx context.Context, id uint, status, lastErr string, processedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"last_error": lastErr,
	}
	if processedAt != nil {
		updates["processed_at"] = processedAt
	}
	return r.db.WithContext(ctx).Model(&modelsOrder.EventOutbox{}).Where("id = ?", id).Updates(updates).Error
}
