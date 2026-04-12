package notification

import (
	modelsCommon "candypro/api/internal/models/common"
	"context"
)

type notificationRepository interface {
	FindByUserID(ctx context.Context, userID string, limit int) ([]modelsCommon.Notification, error)
	MarkRead(ctx context.Context, id uint, userID string) error
	MarkAllRead(ctx context.Context, userID string) error
	Create(ctx context.Context, n *modelsCommon.Notification) error
}

// NotificationService handles notification business logic.
type NotificationService struct {
	repo notificationRepository
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(repo notificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// GetUserNotifications returns notifications for a user.
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID string, limit int) ([]modelsCommon.Notification, error) {
	return s.repo.FindByUserID(ctx, userID, limit)
}

// MarkRead marks a single notification as read.
func (s *NotificationService) MarkRead(ctx context.Context, id uint, userID string) error {
	return s.repo.MarkRead(ctx, id, userID)
}

// MarkAllRead marks all notifications for a user as read.
func (s *NotificationService) MarkAllRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllRead(ctx, userID)
}

// Create creates a new notification.
func (s *NotificationService) Create(ctx context.Context, n *modelsCommon.Notification) error {
	return s.repo.Create(ctx, n)
}
