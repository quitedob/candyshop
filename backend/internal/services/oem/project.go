package oem

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"errors"
)

// ErrOEMStatusConflict signals that an OEM project status update lost a race
// against a concurrent writer — the row's current status no longer matches
// the version we validated against (R2 A-4).
var ErrOEMStatusConflict = errors.New("oem project status changed concurrently")

type projectRepository interface {
	FindAll(ctx context.Context, page, limit int, search, status string) ([]modelsProduct.OEMProject, int64, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsProduct.OEMProject, int64, error)
	FindByID(ctx context.Context, id string) (*modelsProduct.OEMProject, error)
	Create(ctx context.Context, project *modelsProduct.OEMProject) error
	Update(ctx context.Context, project *modelsProduct.OEMProject) error
	UpdateStatusIfMatches(ctx context.Context, id, expected, target string) (int64, error)
	Delete(ctx context.Context, id string) error
}

type ProjectService struct {
	repo projectRepository
}

func NewProjectService(repo projectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetProjects(ctx context.Context, page, limit int, search, status string) ([]modelsProduct.OEMProject, int64, error) {
	return s.repo.FindAll(ctx, page, limit, search, status)
}

func (s *ProjectService) GetUserProjects(ctx context.Context, userID string, page, limit int) ([]modelsProduct.OEMProject, int64, error) {
	return s.repo.FindByUserID(ctx, userID, page, limit)
}

func (s *ProjectService) GetProject(ctx context.Context, id string) (*modelsProduct.OEMProject, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProjectService) CreateProject(ctx context.Context, project *modelsProduct.OEMProject) error {
	if project.Status == "" {
		project.Status = "inquiry"
	}
	return s.repo.Create(ctx, project)
}

func (s *ProjectService) UpdateProject(ctx context.Context, project *modelsProduct.OEMProject) error {
	return s.repo.Update(ctx, project)
}

func (s *ProjectService) UpdateProjectStatus(ctx context.Context, id string, status string) error {
	project, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	// R4-10: Validate status transition
	if err := modelsProduct.ValidateOEMStatusTransition(project.Status, status); err != nil {
		return err
	}
	// R2 A-4: previously this read the row, validated in-memory, then issued a
	// blind Save() — two concurrent admins could each see the same source state
	// and silently overwrite each other. The conditional UPDATE makes the
	// original status part of the WHERE clause; if a competing writer changed
	// it first, RowsAffected is 0 and we surface the conflict.
	rows, uerr := s.repo.UpdateStatusIfMatches(ctx, id, project.Status, status)
	if uerr != nil {
		return uerr
	}
	if rows == 0 {
		return ErrOEMStatusConflict
	}
	return nil
}
