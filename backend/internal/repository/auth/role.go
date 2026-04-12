package auth

import (
	modelsAuth "candypro/api/internal/models/auth"
)

import (
	"context"

	"gorm.io/gorm"
)

// RoleRepository handles role data operations
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new RoleRepository
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// FindByID returns a role by ID
func (r *RoleRepository) FindByID(ctx context.Context, id string) (*modelsAuth.Role, error) {
	var role modelsAuth.Role
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// FindByName returns a role by name
func (r *RoleRepository) FindByName(ctx context.Context, name string) (*modelsAuth.Role, error) {
	var role modelsAuth.Role
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *modelsAuth.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// Update updates a role
func (r *RoleRepository) Update(ctx context.Context, role *modelsAuth.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete deletes a role
func (r *RoleRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsAuth.Role{}, "id = ?", id).Error
}

// FindAll returns all roles
func (r *RoleRepository) FindAll(ctx context.Context) ([]modelsAuth.Role, error) {
	var roles []modelsAuth.Role
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
