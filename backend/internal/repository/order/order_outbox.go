package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UpdateWithOptionalStockReservationAndOutbox atomically reserves stock (when needed), saves the order, and writes outbox.
func (r *OrderRepository) UpdateWithOptionalStockReservationAndOutbox(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int, reserve bool, outbox *modelsOrder.EventOutbox) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", order.ID).First(&locked).Error; err != nil {
			return err
		}

		now := time.Now()
		if reserve && !locked.StockReserved {
			var all []*modelsOrder.StockTransaction
			for productID, qty := range stockDeltas {
				if qty <= 0 {
					continue
				}
				recs, err := reserveStockForOrderLine(tx, warehouseIDFromOrder(&locked), productID, qty, modelsOrder.StockReasonStockReserved, locked.ID, locked.UserID, now)
				if err != nil {
					return err
				}
				all = append(all, recs...)
			}
			if err := writeStockAuditEntries(tx, all); err != nil {
				return err
			}
			order.StockReserved = true
		}

		// M-9: bump Version inside the locked tx so concurrent guarded updates fail loudly.
		order.Version = locked.Version + 1
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

// UpdateWithOutbox 在同一数据库事务内更新订单并写入发件箱（Outbox 模式）
func (r *OrderRepository) UpdateWithOutbox(ctx context.Context, order *modelsOrder.Order, outbox *modelsOrder.EventOutbox) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if order != nil {
			// M-9: bump version on every save (see OrderRepository.Update).
			order.Version++
		}
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
//
// 在单个事务内使用 SELECT ... FOR UPDATE SKIP LOCKED 选取 pending 行并立即递增
// attempts 字段，以此「认领」事件。其他并发 Relay 实例会跳过被锁定的行，避免
// 多实例部署下重复抓取同一事件 (H-14)。事务提交后行锁释放，但 attempts 已变更，
// 后续幂等性由下游 (HasTransactionForOrder) 保证。
func (r *OrderRepository) ListPendingOutbox(ctx context.Context, eventType string, limit int) ([]modelsOrder.EventOutbox, error) {
	var rows []modelsOrder.EventOutbox
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Model(&modelsOrder.EventOutbox{}).
			Where("status = ? AND event_type = ?", modelsOrder.OutboxStatusPending, eventType).
			Order("id ASC")
		if limit > 0 {
			q = q.Limit(limit)
		}
		if err := q.Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		ids := make([]uint, 0, len(rows))
		for _, r := range rows {
			ids = append(ids, r.ID)
		}
		// Bump attempts so concurrent workers (running their own tx) that didn't
		// observe the lock still see a different state and naturally back off if
		// they re-query the same window.
		return tx.Model(&modelsOrder.EventOutbox{}).
			Where("id IN ?", ids).
			UpdateColumn("attempts", gorm.Expr("attempts + ?", 1)).Error
	})
	if err != nil {
		return nil, err
	}
	// Reflect the bumped counter on the returned slice so downstream attempt-cap
	// logic in the relay matches the persisted value.
	for i := range rows {
		rows[i].Attempts++
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
