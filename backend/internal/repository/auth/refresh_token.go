package auth

import (
	modelsAuth "candypro/api/internal/models/auth"
	"context"

	"gorm.io/gorm"
)

// RefreshTokenRepository handles refresh token data operations
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create saves a new refresh token
func (r *RefreshTokenRepository) Create(ctx context.Context, token *modelsAuth.RefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

// FindByToken retrieves a refresh token by its token string
func (r *RefreshTokenRepository) FindByToken(ctx context.Context, tokenStr string) (*modelsAuth.RefreshToken, error) {
	var token modelsAuth.RefreshToken
	if err := r.db.WithContext(ctx).Where("token = ?", tokenStr).First(&token).Error; err != nil {
		return nil, err
	}
	return &token, nil
}

// Update updates an existing refresh token (e.g., for revocation)
func (r *RefreshTokenRepository) Update(ctx context.Context, token *modelsAuth.RefreshToken) error {
	return r.db.WithContext(ctx).Save(token).Error
}

// DeleteAllForUser deletes or revokes all refresh tokens for a specific user
func (r *RefreshTokenRepository) DeleteAllForUser(ctx context.Context, userId string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userId).Delete(&modelsAuth.RefreshToken{}).Error
}

// DeleteByToken deletes a specific token
func (r *RefreshTokenRepository) DeleteByToken(ctx context.Context, tokenStr string) error {
	return r.db.WithContext(ctx).Where("token = ?", tokenStr).Delete(&modelsAuth.RefreshToken{}).Error
}
