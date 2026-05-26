package activitylog

import (
	modelsCommon "candypro/api/internal/models/common"
	"context"
)

type activityLogRepository interface {
	Create(ctx context.Context, log *modelsCommon.ActivityLog) error
	FindByUser(ctx context.Context, userID string, page, limit int) ([]modelsCommon.ActivityLog, int64, error)
	FindAll(ctx context.Context, page, limit int) ([]modelsCommon.ActivityLog, int64, error)
	FindByEntity(ctx context.Context, entityType, entityID string) ([]modelsCommon.ActivityLog, error)
}

// ActivityLogService handles activity log business logic.
type ActivityLogService struct {
	repo activityLogRepository
}

// NewActivityLogService creates a new ActivityLogService.
func NewActivityLogService(repo activityLogRepository) *ActivityLogService {
	return &ActivityLogService{repo: repo}
}

// LogActivity creates a new activity log entry.
func (s *ActivityLogService) LogActivity(ctx context.Context, log *modelsCommon.ActivityLog) error {
	return s.repo.Create(ctx, log)
}

// GetUserActivity returns paginated activity logs for a specific user.
func (s *ActivityLogService) GetUserActivity(ctx context.Context, userID string, page, limit int) ([]modelsCommon.ActivityLog, int64, error) {
	return s.repo.FindByUser(ctx, userID, page, limit)
}

// GetAllActivity returns paginated activity logs for all users.
func (s *ActivityLogService) GetAllActivity(ctx context.Context, page, limit int) ([]modelsCommon.ActivityLog, int64, error) {
	return s.repo.FindAll(ctx, page, limit)
}

// FindByUser returns paginated activity logs for a specific user (alias for handler compatibility).
func (s *ActivityLogService) FindByUser(ctx context.Context, userID string, page, limit int) ([]modelsCommon.ActivityLog, int64, error) {
	return s.repo.FindByUser(ctx, userID, page, limit)
}

// FindAll returns paginated activity logs for all users (alias for handler compatibility).
func (s *ActivityLogService) FindAll(ctx context.Context, page, limit int) ([]modelsCommon.ActivityLog, int64, error) {
	return s.repo.FindAll(ctx, page, limit)
}

// FindByEntity returns chronological activity logs for an entity.
func (s *ActivityLogService) FindByEntity(ctx context.Context, entityType, entityID string) ([]modelsCommon.ActivityLog, error) {
	return s.repo.FindByEntity(ctx, entityType, entityID)
}
