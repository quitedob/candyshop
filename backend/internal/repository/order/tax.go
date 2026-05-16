package order

import (
	"context"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

type TaxRepository struct {
	db *gorm.DB
}

func NewTaxRepository(db *gorm.DB) *TaxRepository {
	return &TaxRepository{db: db}
}

// FindByCountry returns all active tax rates for a country, ordered by specificity (region-specific first).
func (r *TaxRepository) FindByCountry(ctx context.Context, country string) ([]modelsOrder.TaxRate, error) {
	var rates []modelsOrder.TaxRate
	err := r.db.WithContext(ctx).
		Where("country = ? AND is_active = true", country).
		Order("region DESC"). // region-specific comes before empty-region fallback
		Find(&rates).Error
	return rates, err
}

// FindBestRate returns the most specific active tax rate for a country and optional region.
func (r *TaxRepository) FindBestRate(ctx context.Context, country, region string) (*modelsOrder.TaxRate, error) {
	var rate modelsOrder.TaxRate
	q := r.db.WithContext(ctx).
		Where("country = ? AND is_active = true", country)
	if region != "" {
		q = q.Where("region = ? OR region = ''", region).
			Order("CASE WHEN region = '' THEN 1 ELSE 0 END ASC")
	}
	err := q.First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

// Standard CRUD

func (r *TaxRepository) FindAll(ctx context.Context) ([]modelsOrder.TaxRate, error) {
	var rates []modelsOrder.TaxRate
	err := r.db.WithContext(ctx).Order("country ASC, region ASC").Find(&rates).Error
	return rates, err
}

func (r *TaxRepository) FindByID(ctx context.Context, id string) (*modelsOrder.TaxRate, error) {
	var rate modelsOrder.TaxRate
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rate).Error
	if err != nil {
		return nil, err
	}
	return &rate, nil
}

func (r *TaxRepository) Create(ctx context.Context, rate *modelsOrder.TaxRate) error {
	return r.db.WithContext(ctx).Create(rate).Error
}

func (r *TaxRepository) Update(ctx context.Context, rate *modelsOrder.TaxRate) error {
	return r.db.WithContext(ctx).Save(rate).Error
}

func (r *TaxRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&modelsOrder.TaxRate{}).Error
}
