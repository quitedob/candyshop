package activitylog

import (
	modelsCommon "candypro/api/internal/models/common"
	"context"

	"gorm.io/gorm"
)

// ActivityLogRepository handles activity log persistence.
type ActivityLogRepository struct {
	db *gorm.DB
}

// NewActivityLogRepository creates a new ActivityLogRepository.
func NewActivityLogRepository(db *gorm.DB) *ActivityLogRepository {
	return &ActivityLogRepository{db: db}
}

// Create inserts a new activity log entry.
func (r *ActivityLogRepository) Create(ctx context.Context, log *modelsCommon.ActivityLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// FindByUser returns paginated activity logs for a specific user.
func (r *ActivityLogRepository) FindByUser(ctx context.Context, userID string, page, limit int) ([]modelsCommon.ActivityLog, int64, error) {
	var logs []modelsCommon.ActivityLog
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsCommon.ActivityLog{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// FindByEntity returns activity logs for a specific entity ordered chronologically.
func (r *ActivityLogRepository) FindByEntity(ctx context.Context, entityType, entityID string) ([]modelsCommon.ActivityLog, error) {
	var logs []modelsCommon.ActivityLog
	err := r.db.WithContext(ctx).
		Where("entity_type = ? AND entity_id = ?", entityType, entityID).
		Order("created_at ASC").
		Find(&logs).Error
	return logs, err
}

// FindAll returns paginated activity logs for all users.
func (r *ActivityLogRepository) FindAll(ctx context.Context, page, limit int) ([]modelsCommon.ActivityLog, int64, error) {
	var logs []modelsCommon.ActivityLog
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsCommon.ActivityLog{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}
