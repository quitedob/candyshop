package systemsetting

import (
	modelsCommon "candypro/api/internal/models/common"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SystemSettingRepository handles system setting persistence.
type SystemSettingRepository struct {
	db *gorm.DB
}

// NewSystemSettingRepository creates a new SystemSettingRepository.
func NewSystemSettingRepository(db *gorm.DB) *SystemSettingRepository {
	return &SystemSettingRepository{db: db}
}

// FindAll returns all system settings.
func (r *SystemSettingRepository) FindAll(ctx context.Context) ([]modelsCommon.SystemSetting, error) {
	var settings []modelsCommon.SystemSetting
	if err := r.db.WithContext(ctx).Order("category, key").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

// FindByCategory returns system settings filtered by category.
func (r *SystemSettingRepository) FindByCategory(ctx context.Context, category string) ([]modelsCommon.SystemSetting, error) {
	var settings []modelsCommon.SystemSetting
	if err := r.db.WithContext(ctx).Where("category = ?", category).Order("key").Find(&settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

// Upsert creates or updates a system setting.
func (r *SystemSettingRepository) Upsert(ctx context.Context, setting *modelsCommon.SystemSetting) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "category", "updated_by", "updated_at"}),
	}).Create(setting).Error
}
