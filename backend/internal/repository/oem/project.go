package oem

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) FindAll(ctx context.Context, page, limit int) ([]modelsProduct.OEMProject, int64, error) {
	var projects []modelsProduct.OEMProject
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsProduct.OEMProject{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, 0, err
	}
	return projects, total, nil
}

func (r *ProjectRepository) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsProduct.OEMProject, int64, error) {
	var projects []modelsProduct.OEMProject
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsProduct.OEMProject{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&projects).Error; err != nil {
		return nil, 0, err
	}
	return projects, total, nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*modelsProduct.OEMProject, error) {
	var project modelsProduct.OEMProject
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) Create(ctx context.Context, project *modelsProduct.OEMProject) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *ProjectRepository) Update(ctx context.Context, project *modelsProduct.OEMProject) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.OEMProject{}, "id = ?", id).Error
}
