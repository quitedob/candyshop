package oem

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type oemRepository interface {
	FindAllFlows(ctx context.Context) ([]modelsProduct.OEMFlow, error)
	FindFlowByID(ctx context.Context, id string) (*modelsProduct.OEMFlow, error)
	FindFlowsByType(ctx context.Context, flowType string) ([]modelsProduct.OEMFlow, error)
	FindAllSolutions(ctx context.Context) ([]modelsProduct.OEMSolution, error)
	FindSolutionBySlug(ctx context.Context, slug string) (*modelsProduct.OEMSolution, error)
	FindSolutionsByCategory(ctx context.Context, category string) ([]modelsProduct.OEMSolution, error)
	CreateFlow(ctx context.Context, flow *modelsProduct.OEMFlow) error
	CreateSolution(ctx context.Context, solution *modelsProduct.OEMSolution) error
	UpdateFlow(ctx context.Context, flow *modelsProduct.OEMFlow) error
	UpdateSolution(ctx context.Context, solution *modelsProduct.OEMSolution) error
	DeleteFlow(ctx context.Context, id string) error
	DeleteSolution(ctx context.Context, id string) error
}

// OEMService handles OEM business logic.
type OEMService struct {
	repo oemRepository
}

// NewOEMService creates a new OEMService.
func NewOEMService(repo oemRepository) *OEMService {
	return &OEMService{repo: repo}
}

// GetFlows returns all OEM flows.
func (s *OEMService) GetFlows(ctx context.Context) ([]modelsProduct.OEMFlow, error) {
	return s.repo.FindAllFlows(ctx)
}

// GetFlowByID returns an OEM flow by ID.
func (s *OEMService) GetFlowByID(ctx context.Context, id string) (*modelsProduct.OEMFlow, error) {
	return s.repo.FindFlowByID(ctx, id)
}

// GetFlowsByType returns OEM flows by type.
func (s *OEMService) GetFlowsByType(ctx context.Context, flowType string) ([]modelsProduct.OEMFlow, error) {
	return s.repo.FindFlowsByType(ctx, flowType)
}

// GetSolutions returns all OEM solutions.
func (s *OEMService) GetSolutions(ctx context.Context) ([]modelsProduct.OEMSolution, error) {
	return s.repo.FindAllSolutions(ctx)
}

// GetSolutionBySlug returns an OEM solution by slug.
func (s *OEMService) GetSolutionBySlug(ctx context.Context, slug string) (*modelsProduct.OEMSolution, error) {
	return s.repo.FindSolutionBySlug(ctx, slug)
}

// GetSolutionsByCategory returns OEM solutions by category.
func (s *OEMService) GetSolutionsByCategory(ctx context.Context, category string) ([]modelsProduct.OEMSolution, error) {
	return s.repo.FindSolutionsByCategory(ctx, category)
}

// CreateFlow creates a new OEM flow.
func (s *OEMService) CreateFlow(ctx context.Context, flow *modelsProduct.OEMFlow) error {
	return s.repo.CreateFlow(ctx, flow)
}

// CreateSolution creates a new OEM solution.
func (s *OEMService) CreateSolution(ctx context.Context, solution *modelsProduct.OEMSolution) error {
	return s.repo.CreateSolution(ctx, solution)
}

// UpdateFlow updates an OEM flow.
func (s *OEMService) UpdateFlow(ctx context.Context, flow *modelsProduct.OEMFlow) error {
	return s.repo.UpdateFlow(ctx, flow)
}

// UpdateSolution updates an OEM solution.
func (s *OEMService) UpdateSolution(ctx context.Context, solution *modelsProduct.OEMSolution) error {
	return s.repo.UpdateSolution(ctx, solution)
}

// DeleteFlow deletes an OEM flow.
func (s *OEMService) DeleteFlow(ctx context.Context, id string) error {
	return s.repo.DeleteFlow(ctx, id)
}

// DeleteSolution deletes an OEM solution.
func (s *OEMService) DeleteSolution(ctx context.Context, id string) error {
	return s.repo.DeleteSolution(ctx, id)
}
