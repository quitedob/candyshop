package systemsetting

import (
	modelsCommon "candypro/api/internal/models/common"
	"context"
	"time"
)

type systemSettingRepository interface {
	FindAll(ctx context.Context) ([]modelsCommon.SystemSetting, error)
	FindByCategory(ctx context.Context, category string) ([]modelsCommon.SystemSetting, error)
	Upsert(ctx context.Context, setting *modelsCommon.SystemSetting) error
}

// SystemSettingService handles system setting business logic.
type SystemSettingService struct {
	repo systemSettingRepository
}

// NewSystemSettingService creates a new SystemSettingService.
func NewSystemSettingService(repo systemSettingRepository) *SystemSettingService {
	return &SystemSettingService{repo: repo}
}

// GetAllSettings returns all system settings.
func (s *SystemSettingService) GetAllSettings(ctx context.Context) ([]modelsCommon.SystemSetting, error) {
	return s.repo.FindAll(ctx)
}

// GetSettingsByCategory returns system settings for a category.
func (s *SystemSettingService) GetSettingsByCategory(ctx context.Context, category string) ([]modelsCommon.SystemSetting, error) {
	return s.repo.FindByCategory(ctx, category)
}

// UpsertSetting creates or updates a system setting.
func (s *SystemSettingService) UpsertSetting(ctx context.Context, setting *modelsCommon.SystemSetting) error {
	setting.UpdatedAt = time.Now()
	return s.repo.Upsert(ctx, setting)
}

// FindAll returns all system settings (alias for handler compatibility).
func (s *SystemSettingService) FindAll(ctx context.Context) ([]modelsCommon.SystemSetting, error) {
	return s.repo.FindAll(ctx)
}

// FindByCategory returns system settings filtered by category (alias for handler compatibility).
func (s *SystemSettingService) FindByCategory(ctx context.Context, category string) ([]modelsCommon.SystemSetting, error) {
	return s.repo.FindByCategory(ctx, category)
}

// Upsert creates or updates a system setting by key/value/category (alias for handler compatibility).
func (s *SystemSettingService) Upsert(ctx context.Context, key, value, category string) error {
	setting := &modelsCommon.SystemSetting{
		Key:       key,
		Value:     value,
		Category:  category,
		UpdatedAt: time.Now(),
	}
	return s.repo.Upsert(ctx, setting)
}
