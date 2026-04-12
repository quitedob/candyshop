package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"
)

type shipmentRepository interface {
	FindAll(ctx context.Context, page, pageSize int, status string) ([]modelsTrade.ShipmentTracking, int64, error)
	FindByID(ctx context.Context, id uint) (*modelsTrade.ShipmentTracking, error)
	FindByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ShipmentTracking, error)
	Create(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error
	Update(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error
	Delete(ctx context.Context, id uint) error
}

// ShipmentService provides shipment business logic.
type ShipmentService struct {
	repo shipmentRepository
}

// NewShipmentService creates a ShipmentService.
func NewShipmentService(repo shipmentRepository) *ShipmentService {
	return &ShipmentService{repo: repo}
}

// ListShipments returns paginated shipments with optional status filter.
func (s *ShipmentService) ListShipments(ctx context.Context, page, limit int, status string) ([]modelsTrade.ShipmentTracking, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.FindAll(ctx, page, limit, status)
}

// GetShipment returns a single shipment by ID.
func (s *ShipmentService) GetShipment(ctx context.Context, id uint) (*modelsTrade.ShipmentTracking, error) {
	return s.repo.FindByID(ctx, id)
}

// GetByTransactionID returns all shipments for a trade transaction.
func (s *ShipmentService) GetByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ShipmentTracking, error) {
	return s.repo.FindByTransactionID(ctx, transactionID)
}

// CreateShipment creates a new shipment record with a default status.
func (s *ShipmentService) CreateShipment(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error {
	if shipment.Status == "" {
		shipment.Status = "PENDING"
	}
	return s.repo.Create(ctx, shipment)
}

// UpdateShipment saves changes to a shipment.
func (s *ShipmentService) UpdateShipment(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error {
	return s.repo.Update(ctx, shipment)
}

// DeleteShipment removes a shipment.
func (s *ShipmentService) DeleteShipment(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
