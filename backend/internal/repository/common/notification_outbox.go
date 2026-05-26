package common

import (
	"context"
	"time"

	modelsCommon "candypro/api/internal/models/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// NotificationOutboxRepository operates on the notification_outbox table (C-8).
type NotificationOutboxRepository struct {
	db *gorm.DB
}

// NewNotificationOutboxRepository creates a new NotificationOutboxRepository.
func NewNotificationOutboxRepository(db *gorm.DB) *NotificationOutboxRepository {
	return &NotificationOutboxRepository{db: db}
}

// Create inserts a new pending outbox row. The caller is expected to be inside
// a transaction so the row commits atomically with the originating change.
func (r *NotificationOutboxRepository) Create(ctx context.Context, row *modelsCommon.NotificationOutbox) error {
	return r.db.WithContext(ctx).Create(row).Error
}

// CreateInTx 同事务写入。tx 必须由调用方传入（与业务变更同 tx）。
func (r *NotificationOutboxRepository) CreateInTx(tx *gorm.DB, row *modelsCommon.NotificationOutbox) error {
	return tx.Create(row).Error
}

// ClaimPending uses SELECT ... FOR UPDATE SKIP LOCKED to fetch up to limit
// pending rows due for delivery, atomically incrementing attempts so concurrent
// relay instances cannot grab the same row.
func (r *NotificationOutboxRepository) ClaimPending(ctx context.Context, limit int) ([]modelsCommon.NotificationOutbox, error) {
	var rows []modelsCommon.NotificationOutbox
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		q := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND next_attempt <= ?", modelsCommon.NotificationOutboxStatusPending, time.Now()).
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
		return tx.Model(&modelsCommon.NotificationOutbox{}).
			Where("id IN ?", ids).
			UpdateColumn("attempts", gorm.Expr("attempts + ?", 1)).Error
	})
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].Attempts++
	}
	return rows, nil
}

// MarkProcessed sets a row to processed.
func (r *NotificationOutboxRepository) MarkProcessed(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&modelsCommon.NotificationOutbox{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       modelsCommon.NotificationOutboxStatusProcessed,
			"processed_at": &now,
		}).Error
}

// MarkFailed updates a row with a last_error, scheduling a retry via NextAttempt.
// When attempts exceeds maxAttempts the row is left as failed and not re-attempted.
func (r *NotificationOutboxRepository) MarkFailed(ctx context.Context, id uint, msg string, attempts int, maxAttempts int, backoff time.Duration) error {
	updates := map[string]interface{}{
		"last_error":   msg,
		"next_attempt": time.Now().Add(backoff),
	}
	if attempts >= maxAttempts {
		updates["status"] = modelsCommon.NotificationOutboxStatusFailed
	}
	return r.db.WithContext(ctx).Model(&modelsCommon.NotificationOutbox{}).
		Where("id = ?", id).Updates(updates).Error
}

// CleanupProcessed deletes processed rows older than `before`.
func (r *NotificationOutboxRepository) CleanupProcessed(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("status = ? AND processed_at < ?", modelsCommon.NotificationOutboxStatusProcessed, before).
		Delete(&modelsCommon.NotificationOutbox{})
	return res.RowsAffected, res.Error
}
