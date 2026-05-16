package product

import (
	"context"
	"time"

	modelsProduct "candypro/api/internal/models/product"

	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) List(ctx context.Context) ([]modelsProduct.Channel, error) {
	var channels []modelsProduct.Channel
	err := r.db.WithContext(ctx).Order("id ASC").Find(&channels).Error
	return channels, err
}

func (r *ChannelRepository) Get(ctx context.Context, id uint) (*modelsProduct.Channel, error) {
	var ch modelsProduct.Channel
	err := r.db.WithContext(ctx).First(&ch, id).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) GetByCode(ctx context.Context, code string) (*modelsProduct.Channel, error) {
	var ch modelsProduct.Channel
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&ch).Error
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

func (r *ChannelRepository) Create(ctx context.Context, ch *modelsProduct.Channel) error {
	return r.db.WithContext(ctx).Create(ch).Error
}

func (r *ChannelRepository) Update(ctx context.Context, ch *modelsProduct.Channel) error {
	ch.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(ch).Error
}

func (r *ChannelRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.Channel{}, id).Error
}
