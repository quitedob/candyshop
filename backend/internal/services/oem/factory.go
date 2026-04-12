package oem

import (
	modelsProduct "candypro/api/internal/models/product"
	"context"
)

type factoryRepository interface {
	GetInfo(ctx context.Context) (*modelsProduct.FactoryInfo, error)
	UpdateInfo(ctx context.Context, info *modelsProduct.FactoryInfo) error
	FindAllCertifications(ctx context.Context) ([]modelsProduct.Certification, error)
	FindCertificationByID(ctx context.Context, id string) (*modelsProduct.Certification, error)
	FindAllProcessControls(ctx context.Context) ([]modelsProduct.ProcessControl, error)
	FindProcessControlsByStage(ctx context.Context, stage string) ([]modelsProduct.ProcessControl, error)
	CreateCertification(ctx context.Context, cert *modelsProduct.Certification) error
	CreateProcessControl(ctx context.Context, control *modelsProduct.ProcessControl) error
	UpdateCertification(ctx context.Context, cert *modelsProduct.Certification) error
	DeleteCertification(ctx context.Context, id string) error
	DeleteProcessControl(ctx context.Context, id string) error
}

// FactoryService handles factory business logic.
type FactoryService struct {
	repo factoryRepository
}

// NewFactoryService creates a new FactoryService.
func NewFactoryService(repo factoryRepository) *FactoryService {
	return &FactoryService{repo: repo}
}

// GetInfo returns factory information.
func (s *FactoryService) GetInfo(ctx context.Context) (*modelsProduct.FactoryInfo, error) {
	return s.repo.GetInfo(ctx)
}

// UpdateInfo updates factory information.
func (s *FactoryService) UpdateInfo(ctx context.Context, info *modelsProduct.FactoryInfo) error {
	return s.repo.UpdateInfo(ctx, info)
}

// GetCertifications returns all certifications.
func (s *FactoryService) GetCertifications(ctx context.Context) ([]modelsProduct.Certification, error) {
	return s.repo.FindAllCertifications(ctx)
}

// GetCertificationByID returns a certification by ID.
func (s *FactoryService) GetCertificationByID(ctx context.Context, id string) (*modelsProduct.Certification, error) {
	return s.repo.FindCertificationByID(ctx, id)
}

// GetProcessControls returns all process controls.
func (s *FactoryService) GetProcessControls(ctx context.Context) ([]modelsProduct.ProcessControl, error) {
	return s.repo.FindAllProcessControls(ctx)
}

// GetProcessControlsByStage returns process controls by stage.
func (s *FactoryService) GetProcessControlsByStage(ctx context.Context, stage string) ([]modelsProduct.ProcessControl, error) {
	return s.repo.FindProcessControlsByStage(ctx, stage)
}

// CreateCertification creates a new certification.
func (s *FactoryService) CreateCertification(ctx context.Context, cert *modelsProduct.Certification) error {
	return s.repo.CreateCertification(ctx, cert)
}

// CreateProcessControl creates a new process control.
func (s *FactoryService) CreateProcessControl(ctx context.Context, control *modelsProduct.ProcessControl) error {
	return s.repo.CreateProcessControl(ctx, control)
}

// UpdateCertification updates a certification.
func (s *FactoryService) UpdateCertification(ctx context.Context, cert *modelsProduct.Certification) error {
	return s.repo.UpdateCertification(ctx, cert)
}

// DeleteCertification deletes a certification.
func (s *FactoryService) DeleteCertification(ctx context.Context, id string) error {
	return s.repo.DeleteCertification(ctx, id)
}

// DeleteProcessControl deletes a process control.
func (s *FactoryService) DeleteProcessControl(ctx context.Context, id string) error {
	return s.repo.DeleteProcessControl(ctx, id)
}
