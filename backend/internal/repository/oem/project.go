package oem

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) FindAll(ctx context.Context, page, limit int, search, status string) ([]modelsProduct.OEMProject, int64, error) {
	var projects []modelsProduct.OEMProject
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsProduct.OEMProject{})
	if s := strings.TrimSpace(search); s != "" {
		like := "%" + s + "%"
		query = query.Where("product_name ILIKE ? OR id ILIKE ?", like, like)
	}
	if st := strings.TrimSpace(status); st != "" {
		query = query.Where("status = ?", st)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Preload("User").Offset(offset).Limit(limit).Order("created_at DESC").Find(&projects).Error; err != nil {
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

// UpdateStatusIfMatches performs a conditional UPDATE that only succeeds when
// the row's current status equals expected. Returns the affected row count so
// the caller can detect a lost race (R2 A-4). This closes the OEM TOCTOU
// window without introducing a Version column or relying on row locks.
func (r *ProjectRepository) UpdateStatusIfMatches(ctx context.Context, id, expected, target string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&modelsProduct.OEMProject{}).
		Where("id = ? AND status = ?", id, expected).
		Updates(map[string]interface{}{
			"status":     target,
			"updated_at": time.Now(),
		})
	return res.RowsAffected, res.Error
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.OEMProject{}, "id = ?", id).Error
}
