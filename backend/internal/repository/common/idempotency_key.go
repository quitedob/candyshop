package common

import (
	"context"
	"errors"
	"time"

	modelsCommon "candypro/api/internal/models/common"

	"gorm.io/gorm"
)

// ErrIdempotencyKeyConflict 表示同一请求 key 但不同请求体（hash 不一致）。
var ErrIdempotencyKeyConflict = errors.New("idempotency key conflict")

// IdempotencyKeyRepository 操作 idempotency_keys 表（M-17）。
type IdempotencyKeyRepository struct {
	db *gorm.DB
}

// NewIdempotencyKeyRepository creates a new IdempotencyKeyRepository.
func NewIdempotencyKeyRepository(db *gorm.DB) *IdempotencyKeyRepository {
	return &IdempotencyKeyRepository{db: db}
}

// Find 返回匹配的活跃记录（未过期）。
func (r *IdempotencyKeyRepository) Find(ctx context.Context, userID, method, path, key string) (*modelsCommon.IdempotencyKey, error) {
	var rec modelsCommon.IdempotencyKey
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND method = ? AND path = ? AND key = ? AND expires_at > ?", userID, method, path, key, time.Now()).
		First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// Save 写入一条幂等记录；若同 key 已存在但 RequestHash 不同，返回 ErrIdempotencyKeyConflict。
func (r *IdempotencyKeyRepository) Save(ctx context.Context, rec *modelsCommon.IdempotencyKey) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing modelsCommon.IdempotencyKey
		err := tx.Where("user_id = ? AND method = ? AND path = ? AND key = ?", rec.UserID, rec.Method, rec.Path, rec.Key).
			First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(rec).Error
		}
		if err != nil {
			return err
		}
		// Same key, same payload: refresh existing row's response if needed.
		if existing.RequestHash != rec.RequestHash {
			return ErrIdempotencyKeyConflict
		}
		// Same payload, same key — update with the latest response (e.g. retry
		// while the first invocation was still in flight on another instance).
		return tx.Model(&existing).Updates(map[string]interface{}{
			"status_code":           rec.StatusCode,
			"response_body":         rec.ResponseBody,
			"response_c_type":       rec.ResponseCType,
			"expires_at":            rec.ExpiresAt,
		}).Error
	})
}

// CleanupExpired removes idempotency rows older than the given cutoff.
func (r *IdempotencyKeyRepository) CleanupExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("expires_at < ?", before).
		Delete(&modelsCommon.IdempotencyKey{})
	return res.RowsAffected, res.Error
}
