package notification

import (
	modelsCommon "candypro/api/internal/models/common"
	"context"
	"time"

	"gorm.io/gorm"
)

// NotificationRepository handles notification data operations.
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new NotificationRepository.
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// FindByUserID returns notifications for a user, newest first.
func (r *NotificationRepository) FindByUserID(ctx context.Context, userID string, limit int) ([]modelsCommon.Notification, error) {
	var notifications []modelsCommon.Notification
	q := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if err := q.Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkRead marks a single notification as read.
func (r *NotificationRepository) MarkRead(ctx context.Context, id uint, userID string) error {
	return r.db.WithContext(ctx).Model(&modelsCommon.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"is_read": true, "updated_at": time.Now()}).Error
}

// MarkAllRead marks all notifications for a user as read.
func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&modelsCommon.Notification{}).
		Where("user_id = ? AND is_read = false", userID).
		Updates(map[string]interface{}{"is_read": true, "updated_at": time.Now()}).Error
}

// Create creates a new notification.
func (r *NotificationRepository) Create(ctx context.Context, n *modelsCommon.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}
