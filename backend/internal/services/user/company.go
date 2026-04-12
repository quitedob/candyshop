package user

import (
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	"context"
	"fmt"
)

type companyRepository interface {
	FindAll(ctx context.Context, page, limit int) ([]modelsUser.Company, int64, error)
	FindByID(ctx context.Context, id string) (*modelsUser.Company, error)
	Create(ctx context.Context, company *modelsUser.Company) error
	Update(ctx context.Context, company *modelsUser.Company) error
	Delete(ctx context.Context, id string) error
	VerifyCompany(ctx context.Context, id string, status string) error
}

// CompanyService handles company business logic.
type CompanyService struct {
	repo companyRepository
	user *UserService
}

// NewCompanyService creates a new CompanyService.
func NewCompanyService(repo companyRepository, userSvc *UserService) *CompanyService {
	return &CompanyService{repo: repo, user: userSvc}
}

// GetCompanies returns paginated companies.
func (s *CompanyService) GetCompanies(ctx context.Context, page, limit int) (*modelsProduct.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}

	companies, total, err := s.repo.FindAll(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &modelsProduct.PaginatedResponse{
		Data: companies,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// GetCompany returns a single company by ID.
func (s *CompanyService) GetCompany(ctx context.Context, id string) (*modelsUser.Company, error) {
	return s.repo.FindByID(ctx, id)
}

// CreateCompany creates a new company.
func (s *CompanyService) CreateCompany(ctx context.Context, company *modelsUser.Company) error {
	if company.Status == "" {
		company.Status = "pending"
	}
	return s.repo.Create(ctx, company)
}

// UpdateCompany updates a company.
func (s *CompanyService) UpdateCompany(ctx context.Context, company *modelsUser.Company) error {
	return s.repo.Update(ctx, company)
}

// DeleteCompany deletes a company.
func (s *CompanyService) DeleteCompany(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// VerifyCompany verifies or rejects a company and updates linked user status.
func (s *CompanyService) VerifyCompany(ctx context.Context, id string, status string) error {
	if status != "verified" && status != "rejected" {
		return fmt.Errorf("invalid verification status: %s", status)
	}

	if err := s.repo.VerifyCompany(ctx, id, status); err != nil {
		return err
	}

	// When company is verified, activate all pending users linked to it
	if status == "verified" && s.user != nil {
		// R4-11: Propagate activation error — don't silently swallow it
		if err := s.user.ActivateUsersByCompanyID(ctx, id); err != nil {
			return fmt.Errorf("company verified but user activation failed: %w", err)
		}
	}

	return nil
}
