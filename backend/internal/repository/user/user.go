package user

import (
	modelsUser "candypro/api/internal/models/user"
	"context"
	"time"

	"gorm.io/gorm"
)

// UserRepository handles user data operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID returns a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*modelsUser.User, error) {
	var user modelsUser.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail returns a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*modelsUser.User, error) {
	var user modelsUser.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByResetToken returns a user by reset token
func (r *UserRepository) FindByResetToken(ctx context.Context, token string) (*modelsUser.User, error) {
	var user modelsUser.User
	if err := r.db.WithContext(ctx).Preload("Role").Where("reset_token = ?", token).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *modelsUser.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *modelsUser.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsUser.User{}, "id = ?", id).Error
}

// FindAll returns paginated users
func (r *UserRepository) FindAll(ctx context.Context, page, limit int) ([]modelsUser.User, int64, error) {
	var users []modelsUser.User
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsUser.User{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("Role").Offset(offset).Limit(limit).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// CountAll returns total user count.
func (r *UserRepository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&modelsUser.User{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountCreatedSince returns user count created after given time.
func (r *UserRepository) CountCreatedSince(ctx context.Context, since time.Time) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&modelsUser.User{}).
		Where("created_at >= ?", since).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindRecent returns latest users.
func (r *UserRepository) FindRecent(ctx context.Context, limit int) ([]modelsUser.User, error) {
	var users []modelsUser.User
	if err := r.db.WithContext(ctx).
		Preload("Role").
		Order("created_at DESC").
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindByRoleNames returns users whose role name is in the provided list.
func (r *UserRepository) FindByRoleNames(ctx context.Context, names []string) ([]modelsUser.User, error) {
	var users []modelsUser.User
	err := r.db.WithContext(ctx).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name IN ?", names).
		Preload("Role").
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

// ActivateUsersByCompanyID sets all pending users of a company to active.
func (r *UserRepository) ActivateUsersByCompanyID(ctx context.Context, companyID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&modelsUser.User{}).
		Where("company_id = ? AND status = ?", companyID, "pending").
		Updates(map[string]interface{}{
			"status":     "active",
			"updated_at": now,
		}).Error
}
