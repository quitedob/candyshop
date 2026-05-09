package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"

	"gorm.io/gorm"
)

// DocumentAdjustmentRepository 持久化单证调整审计记录
type DocumentAdjustmentRepository struct {
	db *gorm.DB
}

// NewDocumentAdjustmentRepository 构造仓库
func NewDocumentAdjustmentRepository(db *gorm.DB) *DocumentAdjustmentRepository {
	return &DocumentAdjustmentRepository{db: db}
}

// Create 写入一条审计
func (r *DocumentAdjustmentRepository) Create(ctx context.Context, row *modelsOrder.DocumentAdjustment) error {
	return r.db.WithContext(ctx).Create(row).Error
}
