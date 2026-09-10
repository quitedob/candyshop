package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ReturnRepository handles return/refund data operations.
type ReturnRepository struct {
	db *gorm.DB
}

// NewReturnRepository creates a new ReturnRepository.
func NewReturnRepository(db *gorm.DB) *ReturnRepository {
	return &ReturnRepository{db: db}
}

// Create creates a return request with items.
func (r *ReturnRepository) Create(ctx context.Context, ret *modelsOrder.ReturnRequest, items []modelsOrder.ReturnItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		ret.CreatedAt = now
		ret.UpdatedAt = now
		if err := tx.Create(ret).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}
		// R2 E-6: previously this was `for i := range items { tx.Create(&items[i]) }`,
		// producing N individual INSERTs. CreateInBatches keeps the parent/items
		// in the same transaction while collapsing the round-trips. The batch
		// size of 100 is conservative — well under PostgreSQL's parameter limit
		// at ~10 cols × 100 rows = 1k bind params.
		for i := range items {
			items[i].ReturnID = ret.ID
			items[i].CreatedAt = now
			items[i].UpdatedAt = now
		}
		return tx.CreateInBatches(items, 100).Error
	})
}

// FindAll returns paginated returns with optional status filter.
//
// Preloads Items so admin list views can show line counts/totals without a
// per-row follow-up query (H-6).
func (r *ReturnRepository) FindAll(ctx context.Context, status string, page, limit int) ([]modelsOrder.ReturnRequest, int64, error) {
	var returns []modelsOrder.ReturnRequest
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.ReturnRequest{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Preload("Items").Offset(offset).Limit(limit).Order("created_at DESC").Find(&returns).Error; err != nil {
		return nil, 0, err
	}
	return returns, total, nil
}

// FindByID returns a return request with its items.
func (r *ReturnRepository) FindByID(ctx context.Context, id string) (*modelsOrder.ReturnRequest, []modelsOrder.ReturnItem, error) {
	var ret modelsOrder.ReturnRequest
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&ret).Error; err != nil {
		return nil, nil, err
	}
	var items []modelsOrder.ReturnItem
	if err := r.db.WithContext(ctx).Where("return_id = ?", id).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	return &ret, items, nil
}

// FindByUserID returns all return requests for a user.
//
// Preloads Items for the same reason as FindAll (H-6).
func (r *ReturnRepository) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsOrder.ReturnRequest, int64, error) {
	var returns []modelsOrder.ReturnRequest
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.ReturnRequest{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Preload("Items").Offset(offset).Limit(limit).Order("created_at DESC").Find(&returns).Error; err != nil {
		return nil, 0, err
	}
	return returns, total, nil
}

// UpdateStatus transitions a return to a new status with audit fields.
//
// The transition is a guarded conditional UPDATE (WHERE id AND status=<expected>) so two
// concurrent requests cannot both win the same transition — the loser gets
// ErrReturnStateMismatch. Stock restore happens only after the transition succeeds, inside the
// same transaction, so it can never run twice.
func (r *ReturnRepository) UpdateStatus(ctx context.Context, id, status, operatorID, rejectReason string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ret modelsOrder.ReturnRequest
		if err := tx.Where("id = ?", id).First(&ret).Error; err != nil {
			return err
		}

		now := time.Now()
		updates := map[string]interface{}{
			"status":     status,
			"updated_at": now,
		}
		var expected string

		switch status {
		case modelsOrder.ReturnStatusApproved:
			expected = modelsOrder.ReturnStatusPending
			updates["approved_by"] = operatorID
			updates["approved_at"] = now

		case modelsOrder.ReturnStatusReceived:
			expected = modelsOrder.ReturnStatusApproved
			updates["received_by"] = operatorID
			updates["received_at"] = now

		case modelsOrder.ReturnStatusRefunded:
			expected = modelsOrder.ReturnStatusReceived
			updates["refunded_by"] = operatorID
			updates["refunded_at"] = now

		case modelsOrder.ReturnStatusRejected:
			expected = modelsOrder.ReturnStatusPending
			updates["rejected_by"] = operatorID
			updates["rejected_at"] = now
			updates["reject_reason"] = rejectReason

		default:
			return fmt.Errorf("unknown return status: %s", status)
		}

		res := tx.Model(&modelsOrder.ReturnRequest{}).
			Where("id = ? AND status = ?", id, expected).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrReturnStateMismatch
		}

		if status != modelsOrder.ReturnStatusRefunded {
			return nil
		}

		// Restore stock for returned items to the order's warehouse. Runs only after the
		// refund transition above succeeded, so a concurrent second refund cannot double-restore.
		var order modelsOrder.Order
		if err := tx.Where("id = ?", ret.OrderID).First(&order).Error; err != nil {
			return err
		}
		warehouseID := ""
		if order.WarehouseID != nil {
			warehouseID = *order.WarehouseID
		}
		if warehouseID == "" {
			var werr error
			warehouseID, werr = resolveDefaultWarehouseID(tx, "")
			if werr != nil {
				warehouseID = ""
			}
		}
		var items []modelsOrder.ReturnItem
		if err := tx.Where("return_id = ?", id).Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			var err error
			if warehouseID != "" {
				err = restoreLegacyProductStockWithWarehouse(tx, warehouseID, item.ProductID, item.Quantity)
			} else {
				err = restoreLegacyProductStock(tx, item.ProductID, item.Quantity)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
}
