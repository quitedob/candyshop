package oem

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"

	"gorm.io/gorm"
)

// OEMRepository handles OEM data operations
type OEMRepository struct {
	db *gorm.DB
}

// NewOEMRepository creates a new OEMRepository
func NewOEMRepository(db *gorm.DB) *OEMRepository {
	return &OEMRepository{db: db}
}

// FindAllFlows returns all OEM flows
func (r *OEMRepository) FindAllFlows(ctx context.Context) ([]modelsProduct.OEMFlow, error) {
	var flows []modelsProduct.OEMFlow
	if err := r.db.WithContext(ctx).Order("created_at ASC").Find(&flows).Error; err != nil {
		return nil, err
	}
	return flows, nil
}

// FindFlowByID returns an OEM flow by ID
func (r *OEMRepository) FindFlowByID(ctx context.Context, id string) (*modelsProduct.OEMFlow, error) {
	var flow modelsProduct.OEMFlow
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&flow).Error; err != nil {
		return nil, err
	}
	return &flow, nil
}

// FindFlowsByType returns OEM flows by type
func (r *OEMRepository) FindFlowsByType(ctx context.Context, flowType string) ([]modelsProduct.OEMFlow, error) {
	var flows []modelsProduct.OEMFlow
	if err := r.db.WithContext(ctx).Where("type = ?", flowType).Order("created_at ASC").Find(&flows).Error; err != nil {
		return nil, err
	}
	return flows, nil
}

// FindAllSolutions returns all OEM solutions
func (r *OEMRepository) FindAllSolutions(ctx context.Context) ([]modelsProduct.OEMSolution, error) {
	var solutions []modelsProduct.OEMSolution
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&solutions).Error; err != nil {
		return nil, err
	}
	return solutions, nil
}

// FindSolutionBySlug returns an OEM solution by slug
func (r *OEMRepository) FindSolutionBySlug(ctx context.Context, slug string) (*modelsProduct.OEMSolution, error) {
	var solution modelsProduct.OEMSolution
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&solution).Error; err != nil {
		return nil, err
	}
	return &solution, nil
}

// FindSolutionsByCategory returns OEM solutions by category
func (r *OEMRepository) FindSolutionsByCategory(ctx context.Context, category string) ([]modelsProduct.OEMSolution, error) {
	var solutions []modelsProduct.OEMSolution
	if err := r.db.WithContext(ctx).Where("category = ?", category).Order("created_at DESC").Find(&solutions).Error; err != nil {
		return nil, err
	}
	return solutions, nil
}

// CreateFlow creates a new OEM flow
func (r *OEMRepository) CreateFlow(ctx context.Context, flow *modelsProduct.OEMFlow) error {
	return r.db.WithContext(ctx).Create(flow).Error
}

// CreateSolution creates a new OEM solution
func (r *OEMRepository) CreateSolution(ctx context.Context, solution *modelsProduct.OEMSolution) error {
	return r.db.WithContext(ctx).Create(solution).Error
}

// UpdateFlow updates an OEM flow
func (r *OEMRepository) UpdateFlow(ctx context.Context, flow *modelsProduct.OEMFlow) error {
	return r.db.WithContext(ctx).Save(flow).Error
}

// UpdateSolution updates an OEM solution
func (r *OEMRepository) UpdateSolution(ctx context.Context, solution *modelsProduct.OEMSolution) error {
	return r.db.WithContext(ctx).Save(solution).Error
}

// DeleteFlow deletes an OEM flow
func (r *OEMRepository) DeleteFlow(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.OEMFlow{}, "id = ?", id).Error
}

// DeleteSolution deletes an OEM solution
func (r *OEMRepository) DeleteSolution(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.OEMSolution{}, "id = ?", id).Error
}
