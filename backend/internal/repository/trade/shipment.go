package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"time"

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

// UpdateStatus transitions a shipment's status with a guarded conditional UPDATE
// keyed on the currently-loaded status (M3). extra holds additional fields to set
// in the same guarded write.
func (r *ShipmentRepository) UpdateStatus(ctx context.Context, id uint, fromStatus, toStatus string, extra map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":     toStatus,
		"updated_at": time.Now(),
	}
	for k, v := range extra {
		updates[k] = v
	}
	res := r.db.WithContext(ctx).Model(&modelsTrade.ShipmentTracking{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrShipmentStateMismatch
	}
	return nil
}

// Dispatch atomically transitions a PENDING shipment to DISPATCHED using a
// conditional UPDATE ... WHERE id=? AND status='PENDING' (mirrors the optimistic
// lock in QuotationReviewRepository.UpdateStatus). It runs against the provided db
// handle (the caller's transaction) so the transition is atomic with stock
// deduction and dispatch-event creation. Returns true when this call won the
// transition; false when the shipment was already advanced by a concurrent
// dispatch or no longer exists. RowsAffected==0 is the guard that prevents a
// concurrent dispatch from double-deducting stock or emitting duplicate events.
func (r *ShipmentRepository) Dispatch(ctx context.Context, db *gorm.DB, id uint, dispatchedAt time.Time) (bool, error) {
	res := db.WithContext(ctx).
		Model(&modelsTrade.ShipmentTracking{}).
		Where("id = ? AND status = ?", id, "PENDING").
		Updates(map[string]any{
			"status":     "DISPATCHED",
			"updated_at": dispatchedAt,
		})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// Delete removes a shipment by ID.
func (r *ShipmentRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsTrade.ShipmentTracking{}, id).Error
}

// FindTrackable 返回待同步追踪状态的发货记录（非终态且有提单号）。
func (r *ShipmentRepository) FindTrackable(ctx context.Context, limit int) ([]modelsTrade.ShipmentTracking, error) {
	if limit < 1 {
		limit = 50
	}
	var shipments []modelsTrade.ShipmentTracking
	err := r.db.WithContext(ctx).
		Where("status IN ?", []string{"PENDING", "DISPATCHED", "IN_TRANSIT"}).
		Where("bill_of_lading_no <> ''").
		Order("updated_at asc").
		Limit(limit).
		Find(&shipments).Error
	return shipments, err
}

// FindByTransactionIDs returns shipments for multiple transaction IDs (batch preload).
func (r *ShipmentRepository) FindByTransactionIDs(ctx context.Context, transactionIDs []uint) ([]modelsTrade.ShipmentTracking, error) {
	if len(transactionIDs) == 0 {
		return nil, nil
	}
	var shipments []modelsTrade.ShipmentTracking
	if err := r.db.WithContext(ctx).Where("transaction_id IN ?", transactionIDs).Order("created_at asc").Find(&shipments).Error; err != nil {
		return nil, err
	}
	return shipments, nil
}
