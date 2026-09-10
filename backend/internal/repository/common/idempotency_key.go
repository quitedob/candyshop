package common

import (
	"context"
	"errors"
	"time"

	modelsCommon "candypro/api/internal/models/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrIdempotencyKeyConflict 表示同一请求 key 但不同请求体（hash 不一致）。
var ErrIdempotencyKeyConflict = errors.New("idempotency key conflict")

// ErrIdempotencyInProgress prevents concurrent execution or unsafe retries after
// a worker dies before recording the outcome of a mutating operation.
var ErrIdempotencyInProgress = errors.New("idempotency request already in progress")

// IdempotencyKeyRepository 操作 idempotency_keys 表（M-17）。
type IdempotencyKeyRepository struct {
	db *gorm.DB
}

// NewIdempotencyKeyRepository creates a new IdempotencyKeyRepository.
func NewIdempotencyKeyRepository(db *gorm.DB) *IdempotencyKeyRepository {
	return &IdempotencyKeyRepository{db: db}
}

// Reserve atomically claims a key. Existing complete responses are returned for
// replay; pending reservations never expire automatically because their effects
// may already have committed even if the worker failed to record its response.
func (repository *IdempotencyKeyRepository) Reserve(ctx context.Context, rec *modelsCommon.IdempotencyKey) (*modelsCommon.IdempotencyKey, error) {
	var existing *modelsCommon.IdempotencyKey
	err := repository.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ? AND method = ? AND path = ? AND key = ? AND expires_at <= ? AND status_code <> ?",
			rec.UserID, rec.Method, rec.Path, rec.Key, time.Now(), modelsCommon.IdempotencyStatusPending).
			Delete(&modelsCommon.IdempotencyKey{}).Error; err != nil {
			return err
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(rec)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 0 {
			return nil
		}
		existing = &modelsCommon.IdempotencyKey{}
		if err := tx.Where("user_id = ? AND method = ? AND path = ? AND key = ?", rec.UserID, rec.Method, rec.Path, rec.Key).
			First(existing).Error; err != nil {
			return err
		}
		if existing.RequestHash != rec.RequestHash {
			return ErrIdempotencyKeyConflict
		}
		if existing.StatusCode == modelsCommon.IdempotencyStatusPending {
			return ErrIdempotencyInProgress
		}
		return nil
	})
	return existing, err
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
		// The reservation owner records the completed outcome. Other instances
		// cannot run this operation while its pending reservation exists.
		return tx.Model(&existing).Updates(map[string]interface{}{
			"status_code":     rec.StatusCode,
			"response_body":   rec.ResponseBody,
			"response_c_type": rec.ResponseCType,
			"expires_at":      rec.ExpiresAt,
		}).Error
	})
}

// CleanupExpired removes completed responses older than the cutoff, retaining
// pending requests whose side-effect outcome is unknown.
func (r *IdempotencyKeyRepository) CleanupExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("expires_at < ? AND status_code <> ?", before, modelsCommon.IdempotencyStatusPending).
		Delete(&modelsCommon.IdempotencyKey{})
	return res.RowsAffected, res.Error
}
