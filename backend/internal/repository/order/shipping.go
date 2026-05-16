package order

import (
	"context"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

type ShippingRepository struct {
	db *gorm.DB
}

func NewShippingRepository(db *gorm.DB) *ShippingRepository {
	return &ShippingRepository{db: db}
}

func (r *ShippingRepository) FindAll(ctx context.Context) ([]modelsOrder.ShippingRate, error) {
	var rates []modelsOrder.ShippingRate
	err := r.db.WithContext(ctx).Order("destination ASC, min_weight_kg ASC").Find(&rates).Error
	return rates, err
}

func (r *ShippingRepository) FindByID(ctx context.Context, id string) (*modelsOrder.ShippingRate, error) {
	var rate modelsOrder.ShippingRate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *ShippingRepository) Create(ctx context.Context, rate *modelsOrder.ShippingRate) error {
	return r.db.WithContext(ctx).Create(rate).Error
}

func (r *ShippingRepository) Update(ctx context.Context, rate *modelsOrder.ShippingRate) error {
	return r.db.WithContext(ctx).Save(rate).Error
}

func (r *ShippingRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&modelsOrder.ShippingRate{}).Error
}

func (r *ShippingRepository) FindByDestination(ctx context.Context, destination string) ([]modelsOrder.ShippingRate, error) {
	var rates []modelsOrder.ShippingRate
	err := r.db.WithContext(ctx).
		Where("destination = ? AND is_active = true", destination).
		Order("min_weight_kg ASC").
		Find(&rates).Error
	return rates, err
}

// FindBestRate finds the cheapest active shipping rate for a destination and weight.
func (r *ShippingRepository) FindBestRate(ctx context.Context, destination string, weightKg float64) (*modelsOrder.ShippingRate, error) {
	var rate modelsOrder.ShippingRate
	err := r.db.WithContext(ctx).
		Where("destination = ? AND is_active = true AND min_weight_kg <= ? AND max_weight_kg >= ?", destination, weightKg, weightKg).
		Order("base_cost ASC").
		First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}
