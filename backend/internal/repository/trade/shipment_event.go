package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"

	"gorm.io/gorm"
)

// ShipmentEventRepository handles ShipmentEvent persistence.
type ShipmentEventRepository struct {
	db *gorm.DB
}

// NewShipmentEventRepository creates a new ShipmentEventRepository.
func NewShipmentEventRepository(db *gorm.DB) *ShipmentEventRepository {
	return &ShipmentEventRepository{db: db}
}

// FindByShipmentID returns all events for a shipment ordered by eventTime.
func (r *ShipmentEventRepository) FindByShipmentID(ctx context.Context, shipmentID uint) ([]modelsTrade.ShipmentEvent, error) {
	var events []modelsTrade.ShipmentEvent
	if err := r.db.WithContext(ctx).Where("shipment_id = ?", shipmentID).Order("event_time asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// FindByShipmentIDs returns all events for multiple shipments (batch preload).
func (r *ShipmentEventRepository) FindByShipmentIDs(ctx context.Context, shipmentIDs []uint) ([]modelsTrade.ShipmentEvent, error) {
	if len(shipmentIDs) == 0 {
		return nil, nil
	}
	var events []modelsTrade.ShipmentEvent
	if err := r.db.WithContext(ctx).Where("shipment_id IN ?", shipmentIDs).Order("event_time asc").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

// Create inserts a new shipment event.
func (r *ShipmentEventRepository) Create(ctx context.Context, event *modelsTrade.ShipmentEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}
