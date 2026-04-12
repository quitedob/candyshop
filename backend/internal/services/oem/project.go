package oem

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type projectRepository interface {
	FindAll(ctx context.Context, page, limit int) ([]modelsProduct.OEMProject, int64, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsProduct.OEMProject, int64, error)
	FindByID(ctx context.Context, id string) (*modelsProduct.OEMProject, error)
	Create(ctx context.Context, project *modelsProduct.OEMProject) error
	Update(ctx context.Context, project *modelsProduct.OEMProject) error
	Delete(ctx context.Context, id string) error
}

type ProjectService struct {
	repo projectRepository
}

func NewProjectService(repo projectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) GetProjects(ctx context.Context, page, limit int) ([]modelsProduct.OEMProject, int64, error) {
	return s.repo.FindAll(ctx, page, limit)
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
	project.Status = status
	return s.repo.Update(ctx, project)
}
