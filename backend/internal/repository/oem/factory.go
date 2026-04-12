package oem

import (
	modelsProduct "candypro/api/internal/models/product"
)

import (
	"context"

	"gorm.io/gorm"
)

// FactoryRepository handles factory data operations
type FactoryRepository struct {
	db *gorm.DB
}

// NewFactoryRepository creates a new FactoryRepository
func NewFactoryRepository(db *gorm.DB) *FactoryRepository {
	return &FactoryRepository{db: db}
}

// GetInfo returns factory information (singleton pattern - always ID "1")
func (r *FactoryRepository) GetInfo(ctx context.Context) (*modelsProduct.FactoryInfo, error) {
	var info modelsProduct.FactoryInfo
	if err := r.db.WithContext(ctx).Where("id = ?", "1").First(&info).Error; err != nil {
		// Return empty info if not found
		if err == gorm.ErrRecordNotFound {
			return &modelsProduct.FactoryInfo{
				ID:          "1",
				Name:        "CandyPro Factory",
				Founded:     2010,
				Description: "Leading OEM candy manufacturer",
			}, nil
		}
		return nil, err
	}
	return &info, nil
}

// UpdateInfo updates factory information
func (r *FactoryRepository) UpdateInfo(ctx context.Context, info *modelsProduct.FactoryInfo) error {
	info.ID = "1" // Ensure singleton
	return r.db.WithContext(ctx).Save(info).Error
}

// FindAllCertifications returns all certifications
func (r *FactoryRepository) FindAllCertifications(ctx context.Context) ([]modelsProduct.Certification, error) {
	var certifications []modelsProduct.Certification
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&certifications).Error; err != nil {
		return nil, err
	}
	return certifications, nil
}

// FindCertificationByID returns a certification by ID
func (r *FactoryRepository) FindCertificationByID(ctx context.Context, id string) (*modelsProduct.Certification, error) {
	var certification modelsProduct.Certification
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&certification).Error; err != nil {
		return nil, err
	}
	return &certification, nil
}

// FindAllProcessControls returns all process controls
func (r *FactoryRepository) FindAllProcessControls(ctx context.Context) ([]modelsProduct.ProcessControl, error) {
	var controls []modelsProduct.ProcessControl
	if err := r.db.WithContext(ctx).Order("stage, name ASC").Find(&controls).Error; err != nil {
		return nil, err
	}
	return controls, nil
}

// FindProcessControlsByStage returns process controls by stage
func (r *FactoryRepository) FindProcessControlsByStage(ctx context.Context, stage string) ([]modelsProduct.ProcessControl, error) {
	var controls []modelsProduct.ProcessControl
	if err := r.db.WithContext(ctx).Where("stage = ?", stage).Order("name ASC").Find(&controls).Error; err != nil {
		return nil, err
	}
	return controls, nil
}

// CreateCertification creates a new certification
func (r *FactoryRepository) CreateCertification(ctx context.Context, cert *modelsProduct.Certification) error {
	return r.db.WithContext(ctx).Create(cert).Error
}

// CreateProcessControl creates a new process control
func (r *FactoryRepository) CreateProcessControl(ctx context.Context, control *modelsProduct.ProcessControl) error {
	return r.db.WithContext(ctx).Create(control).Error
}

// UpdateCertification updates a certification
func (r *FactoryRepository) UpdateCertification(ctx context.Context, cert *modelsProduct.Certification) error {
	return r.db.WithContext(ctx).Save(cert).Error
}

// DeleteCertification deletes a certification
func (r *FactoryRepository) DeleteCertification(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.Certification{}, "id = ?", id).Error
}

// DeleteProcessControl deletes a process control
func (r *FactoryRepository) DeleteProcessControl(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.ProcessControl{}, "id = ?", id).Error
}
