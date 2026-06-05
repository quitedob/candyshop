package common

import (
	"context"

	modelsCommon "candypro/api/internal/models/common"

	"gorm.io/gorm"
)

// UploadedFileRepository stores upload metadata.
type UploadedFileRepository struct {
	db *gorm.DB
}

func NewUploadedFileRepository(db *gorm.DB) *UploadedFileRepository {
	return &UploadedFileRepository{db: db}
}

func (r *UploadedFileRepository) Create(ctx context.Context, file *modelsCommon.UploadedFile) error {
	return r.db.WithContext(ctx).Create(file).Error
}

func (r *UploadedFileRepository) Save(ctx context.Context, file *modelsCommon.UploadedFile) error {
	return r.db.WithContext(ctx).Save(file).Error
}

func (r *UploadedFileRepository) FindByID(ctx context.Context, id string) (*modelsCommon.UploadedFile, error) {
	var file modelsCommon.UploadedFile
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

func (r *UploadedFileRepository) FindByStorageKey(ctx context.Context, key string) (*modelsCommon.UploadedFile, error) {
	var file modelsCommon.UploadedFile
	if err := r.db.WithContext(ctx).Where("storage_key = ?", key).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}
