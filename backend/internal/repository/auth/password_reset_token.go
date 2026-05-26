package auth

import (
	"context"
	"errors"
	"time"

	modelsAuth "candypro/api/internal/models/auth"

	"gorm.io/gorm"
)

// ErrPasswordResetTokenNotFound 标识未命中或已使用/过期。
var ErrPasswordResetTokenNotFound = errors.New("password reset token not found")

// PasswordResetTokenRepository 操作 password_reset_tokens 表（M-22 / L-8）。
type PasswordResetTokenRepository struct {
	db *gorm.DB
}

// NewPasswordResetTokenRepository creates a new repository.
func NewPasswordResetTokenRepository(db *gorm.DB) *PasswordResetTokenRepository {
	return &PasswordResetTokenRepository{db: db}
}

// Create 写入一条新令牌。调用方应保证 TokenLookupKey 全局唯一。
func (r *PasswordResetTokenRepository) Create(ctx context.Context, token *modelsAuth.PasswordResetToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// FindActiveByLookupKey 通过 lookup key 查询活跃（未使用且未过期）令牌。
func (r *PasswordResetTokenRepository) FindActiveByLookupKey(ctx context.Context, lookupKey string) (*modelsAuth.PasswordResetToken, error) {
	var token modelsAuth.PasswordResetToken
	err := r.db.WithContext(ctx).
		Where("token_lookup_key = ? AND used_at IS NULL AND expires_at > ?", lookupKey, time.Now()).
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrPasswordResetTokenNotFound
	}
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// MarkUsed 标记令牌为已使用，幂等：再次调用不会改变状态。
func (r *PasswordResetTokenRepository) MarkUsed(ctx context.Context, id string) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&modelsAuth.PasswordResetToken{}).
		Where("id = ? AND used_at IS NULL", id).
		Update("used_at", now)
	if res.Error != nil {
		return res.Error
	}
	return nil
}

// InvalidateActiveForUser 把指定用户当前所有活跃令牌一次性标记为已使用，
// 用于 ResetPassword 成功后阻断同时签发的多个链接。
func (r *PasswordResetTokenRepository) InvalidateActiveForUser(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&modelsAuth.PasswordResetToken{}).
		Where("user_id = ? AND used_at IS NULL", userID).
		Update("used_at", now).Error
}

// CleanupExpired 删除指定截止时间之前的已过期或已使用记录，由后台任务调用。
func (r *PasswordResetTokenRepository) CleanupExpired(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("expires_at < ? OR used_at IS NOT NULL", before).
		Delete(&modelsAuth.PasswordResetToken{})
	return res.RowsAffected, res.Error
}
