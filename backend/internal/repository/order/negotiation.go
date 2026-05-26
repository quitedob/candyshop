package order

import (
	"context"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

type NegotiationRepository struct {
	db *gorm.DB
}

func NewNegotiationRepository(db *gorm.DB) *NegotiationRepository {
	return &NegotiationRepository{db: db}
}

func (r *NegotiationRepository) FindByInquiryID(ctx context.Context, inquiryID string) ([]modelsOrder.NegotiationOffer, error) {
	var offers []modelsOrder.NegotiationOffer
	err := r.db.WithContext(ctx).
		Where("inquiry_id = ?", inquiryID).
		Order("created_at ASC").
		Find(&offers).Error
	return offers, err
}

func (r *NegotiationRepository) FindByID(ctx context.Context, id string) (*modelsOrder.NegotiationOffer, error) {
	var offer modelsOrder.NegotiationOffer
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&offer).Error
	if err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *NegotiationRepository) Create(ctx context.Context, offer *modelsOrder.NegotiationOffer) error {
	return r.db.WithContext(ctx).Create(offer).Error
}

func (r *NegotiationRepository) Update(ctx context.Context, offer *modelsOrder.NegotiationOffer) error {
	return r.db.WithContext(ctx).Save(offer).Error
}

// TransitionStatus 原子地把指定 offer 从 fromStatus 改为 toStatus。
// 仅当数据库当前 status = fromStatus 时才会写入；返回 RowsAffected。
//
// A-4: AcceptOffer / RejectOffer 不能再用 read→check→Save 的 TOCTOU 模式，
// 必须依赖此方法的条件 WHERE 防止两个管理员同时接受同一 pending offer。
func (r *NegotiationRepository) TransitionStatus(ctx context.Context, id, fromStatus, toStatus string, updatedAt time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Model(&modelsOrder.NegotiationOffer{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(map[string]any{
			"status":     toStatus,
			"updated_at": updatedAt,
		})
	if res.Error != nil {
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

func (r *NegotiationRepository) FindPendingByInquiryID(ctx context.Context, inquiryID string) (*modelsOrder.NegotiationOffer, error) {
	var offer modelsOrder.NegotiationOffer
	err := r.db.WithContext(ctx).
		Where("inquiry_id = ? AND status = ?", inquiryID, "pending").
		Order("created_at DESC").
		First(&offer).Error
	if err != nil {
		return nil, err
	}
	return &offer, nil
}
