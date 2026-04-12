package user

import (
	modelsUser "candypro/api/internal/models/user"
	"context"
	"time"

	"gorm.io/gorm"
)

// CompanyRepository handles company data operations.
type CompanyRepository struct {
	db *gorm.DB
}

// NewCompanyRepository creates a new CompanyRepository.
func NewCompanyRepository(db *gorm.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

// FindAll returns paginated companies.
func (r *CompanyRepository) FindAll(ctx context.Context, page, limit int) ([]modelsUser.Company, int64, error) {
	var companies []modelsUser.Company
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsUser.Company{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&companies).Error; err != nil {
		return nil, 0, err
	}

	return companies, total, nil
}

// FindByID returns a company by ID.
func (r *CompanyRepository) FindByID(ctx context.Context, id string) (*modelsUser.Company, error) {
	var company modelsUser.Company
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&company).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

// Create creates a new company.
func (r *CompanyRepository) Create(ctx context.Context, company *modelsUser.Company) error {
	return r.db.WithContext(ctx).Create(company).Error
}

// Update updates a company.
func (r *CompanyRepository) Update(ctx context.Context, company *modelsUser.Company) error {
	return r.db.WithContext(ctx).Save(company).Error
}

// Delete deletes a company by ID.
func (r *CompanyRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsUser.Company{}, "id = ?", id).Error
}

// VerifyCompany sets company status to verified and records the timestamp.
func (r *CompanyRepository) VerifyCompany(ctx context.Context, id string, status string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": now,
	}
	if status == "verified" {
		updates["verified_at"] = now
	} else {
		updates["verified_at"] = nil
	}
	return r.db.WithContext(ctx).Model(&modelsUser.Company{}).Where("id = ?", id).Updates(updates).Error
}

// FindByUserID returns the company associated with a user.
func (r *CompanyRepository) FindByUserID(ctx context.Context, userID string) (*modelsUser.Company, error) {
	var user modelsUser.User
	if err := r.db.WithContext(ctx).Select("company_id").Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	if user.CompanyID == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.FindByID(ctx, *user.CompanyID)
}
