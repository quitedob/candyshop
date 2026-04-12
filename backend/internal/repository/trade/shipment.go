package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"

	"gorm.io/gorm"
)

// ShipmentRepository handles ShipmentTracking persistence.
type ShipmentRepository struct {
	db *gorm.DB
}

// NewShipmentRepository creates a new ShipmentRepository.
func NewShipmentRepository(db *gorm.DB) *ShipmentRepository {
	return &ShipmentRepository{db: db}
}

// FindAll returns paginated shipments with optional status filter.
func (r *ShipmentRepository) FindAll(ctx context.Context, page, pageSize int, status string) ([]modelsTrade.ShipmentTracking, int64, error) {
	var shipments []modelsTrade.ShipmentTracking
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsTrade.ShipmentTracking{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&shipments).Error; err != nil {
		return nil, 0, err
	}
	return shipments, total, nil
}

// FindByID returns a shipment by primary key.
func (r *ShipmentRepository) FindByID(ctx context.Context, id uint) (*modelsTrade.ShipmentTracking, error) {
	var shipment modelsTrade.ShipmentTracking
	if err := r.db.WithContext(ctx).First(&shipment, id).Error; err != nil {
		return nil, err
	}
	return &shipment, nil
}

// FindByTransactionID returns all shipments for a trade transaction.
func (r *ShipmentRepository) FindByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ShipmentTracking, error) {
	var shipments []modelsTrade.ShipmentTracking
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Order("created_at asc").Find(&shipments).Error; err != nil {
		return nil, err
	}
	return shipments, nil
}

// Create inserts a new shipment record.
func (r *ShipmentRepository) Create(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error {
	return r.db.WithContext(ctx).Create(shipment).Error
}

// Update saves changes to a shipment.
func (r *ShipmentRepository) Update(ctx context.Context, shipment *modelsTrade.ShipmentTracking) error {
	return r.db.WithContext(ctx).Save(shipment).Error
}

// Delete removes a shipment by ID.
func (r *ShipmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsTrade.ShipmentTracking{}, id).Error
}
