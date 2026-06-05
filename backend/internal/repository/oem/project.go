package oem

import (
	modelsOrder "candypro/api/internal/models/order"
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

// UpdateIfVersionMatches persists a project only when no concurrent mutation
// has advanced its optimistic-lock version.
func (r *ProjectRepository) UpdateIfVersionMatches(ctx context.Context, project *modelsProduct.OEMProject, expectedVersion uint) (int64, error) {
	res := r.db.WithContext(ctx).Model(&modelsProduct.OEMProject{}).
		Where("id = ? AND version = ?", project.ID, expectedVersion).
		Updates(map[string]interface{}{
			"user_id":           project.UserID,
			"inquiry_id":        project.InquiryID,
			"order_id":          project.OrderID,
			"product_name":      project.ProductName,
			"quoted_unit_price": project.QuotedUnitPrice,
			"quoted_quantity":   project.QuotedQuantity,
			"product_id":        project.ProductID,
			"status":            project.Status,
			"current_step":      project.CurrentStep,
			"requirements":      project.Requirements,
			"samples":           project.Samples,
			"attachments":       project.Attachments,
			"assigned_to":       project.AssignedTo,
			"notes":             project.Notes,
			"admin_notes":       project.AdminNotes,
			"version":           gorm.Expr("version + 1"),
			"updated_at":        time.Now(),
		})
	return res.RowsAffected, res.Error
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
			"version":    gorm.Expr("version + 1"),
			"updated_at": time.Now(),
		})
	return res.RowsAffected, res.Error
}

// LinkOrderIfEmpty atomically records the first conversion order only.
func (r *ProjectRepository) LinkOrderIfEmpty(ctx context.Context, id, orderID string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&modelsProduct.OEMProject{}).
		Where("id = ? AND (order_id IS NULL OR order_id = '')", id).
		Updates(map[string]interface{}{
			"order_id":   orderID,
			"version":    gorm.Expr("version + 1"),
			"updated_at": time.Now(),
		})
	return res.RowsAffected, res.Error
}

// CreateLinkedOrderIfEmpty atomically claims an OEM project's conversion and
// creates its draft order. A concurrent loser creates no extra order.
func (r *ProjectRepository) CreateLinkedOrderIfEmpty(ctx context.Context, id string, order *modelsOrder.Order) (bool, error) {
	linked := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&modelsProduct.OEMProject{}).
			Where("id = ? AND (order_id IS NULL OR order_id = '')", id).
			Updates(map[string]interface{}{
				"order_id":   order.ID,
				"version":    gorm.Expr("version + 1"),
				"updated_at": time.Now(),
			})
		if res.Error != nil || res.RowsAffected == 0 {
			return res.Error
		}
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		linked = true
		return nil
	})
	return linked, err
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.OEMProject{}, "id = ?", id).Error
}
