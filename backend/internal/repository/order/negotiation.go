package order

import (
	"context"

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
