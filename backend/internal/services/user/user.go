package user

import (
	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type userRepository interface {
	FindByID(ctx context.Context, id string) (*modelsUser.User, error)
	FindByEmail(ctx context.Context, email string) (*modelsUser.User, error)
	FindByResetToken(ctx context.Context, token string) (*modelsUser.User, error)
	FindAll(ctx context.Context, page, limit int) ([]modelsUser.User, int64, error)
	FindRecent(ctx context.Context, limit int) ([]modelsUser.User, error)
	Create(ctx context.Context, user *modelsUser.User) error
	Update(ctx context.Context, user *modelsUser.User) error
	Delete(ctx context.Context, id string) error
	CountAll(ctx context.Context) (int64, error)
	CountCreatedSince(ctx context.Context, since time.Time) (int64, error)
	ActivateUsersByCompanyID(ctx context.Context, companyID string) error
	FindByRoleNames(ctx context.Context, names []string) ([]modelsUser.User, error)
}

type UserService struct {
	repo userRepository
}

func NewUserService(repo userRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetByID(ctx context.Context, id string) (*modelsUser.User, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*modelsUser.User, error) {
	return s.repo.FindByEmail(ctx, email)
}

func (s *UserService) GetByResetToken(ctx context.Context, token string) (*modelsUser.User, error) {
	return s.repo.FindByResetToken(ctx, token)
}

func (s *UserService) EmailExists(ctx context.Context, email string) (bool, error) {
	_, err := s.repo.FindByEmail(ctx, email)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func (s *UserService) UpdateLastLogin(ctx context.Context, id string) error {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// Using model field for LastLoginAt
	now := time.Now()
	user.LastLoginAt = &now
	return s.repo.Update(ctx, user)
}

func (s *UserService) GetUsers(ctx context.Context, page, limit int) ([]modelsUser.User, int64, error) {
	return s.repo.FindAll(ctx, page, limit)
}

// GetStaffUsers returns a paginated list of staff-role accounts (admin,
// superadmin) for the admin staff-management screen. The lookup is role-scoped
// (FindByRoleNames), so customer and supplier records — and their PII — never
// enter the /admin/staff response. There is no separate "staff" role in the
// platform, so admin + superadmin are the staff set.
func (s *UserService) GetStaffUsers(ctx context.Context, page, limit int) ([]modelsUser.User, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	staff, err := s.repo.FindByRoleNames(ctx, modelsAuth.AdminPortal())
	if err != nil {
		return nil, 0, err
	}
	total := int64(len(staff))
	start := (page - 1) * limit
	if start >= len(staff) {
		return []modelsUser.User{}, total, nil
	}
	end := start + limit
	if end > len(staff) {
		end = len(staff)
	}
	return staff[start:end], total, nil
}

func (s *UserService) UpdateUser(ctx context.Context, user *modelsUser.User) error {
	return s.repo.Update(ctx, user)
}

func (s *UserService) CreateUser(ctx context.Context, user *modelsUser.User) error {
	return s.repo.Create(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CountUsers returns total users.
func (s *UserService) CountUsers(ctx context.Context) (int64, error) {
	return s.repo.CountAll(ctx)
}

// CountUsersSince returns users created after a timestamp.
func (s *UserService) CountUsersSince(ctx context.Context, since time.Time) (int64, error) {
	return s.repo.CountCreatedSince(ctx, since)
}

// GetRecentUsers returns latest users.
func (s *UserService) GetRecentUsers(ctx context.Context, limit int) ([]modelsUser.User, error) {
	return s.repo.FindRecent(ctx, limit)
}

// FindAdminUsers returns all users with the "admin" or "superadmin" role.
func (s *UserService) FindAdminUsers(ctx context.Context) ([]modelsUser.User, error) {
	return s.repo.FindByRoleNames(ctx, []string{"admin", "superadmin"})
}

// ActivateUsersByCompanyID activates all pending users linked to a company.
func (s *UserService) ActivateUsersByCompanyID(ctx context.Context, companyID string) error {
	return s.repo.ActivateUsersByCompanyID(ctx, companyID)
}
