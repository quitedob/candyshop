package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"log"
)

// StockTransferRepository handles stock transfer data operations.
type StockTransferRepository struct {
	db *gorm.DB
}

// NewStockTransferRepository creates a new StockTransferRepository.
func NewStockTransferRepository(db *gorm.DB) *StockTransferRepository {
	return &StockTransferRepository{db: db}
}

// Create creates a stock transfer with items and executes the transfer in a transaction.
func (r *StockTransferRepository) Create(ctx context.Context, transfer *modelsOrder.StockTransfer, items []modelsOrder.StockTransferItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		transfer.CreatedAt = now
		transfer.UpdatedAt = now
		if err := tx.Create(transfer).Error; err != nil {
			return err
		}

		var allAudit []*modelsOrder.StockTransaction
		for i := range items {
			items[i].TransferID = transfer.ID
			items[i].CreatedAt = now
			items[i].UpdatedAt = now
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
			if transfer.Status == modelsOrder.StockTransferStatusCompleted {
				recs, err := TransferStockBetweenWarehouses(tx,
					transfer.FromWarehouseID, transfer.ToWarehouseID,
					items[i].ProductID, items[i].Quantity,
					modelsOrder.StockReasonStockTransfer, transfer.ID, transfer.CreatedBy, now,
				)
				if err != nil {
					return err
				}
				allAudit = append(allAudit, recs...)
			}
		}
		return writeStockAuditEntries(tx, allAudit)
	})
}

// FindAll returns paginated stock transfers.
func (r *StockTransferRepository) FindAll(ctx context.Context, page, limit int) ([]modelsOrder.StockTransfer, int64, error) {
	var transfers []modelsOrder.StockTransfer
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.StockTransfer{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&transfers).Error; err != nil {
		return nil, 0, err
	}
	return transfers, total, nil
}

// FindByID returns a stock transfer with its items.
func (r *StockTransferRepository) FindByID(ctx context.Context, id string) (*modelsOrder.StockTransfer, []modelsOrder.StockTransferItem, error) {
	var transfer modelsOrder.StockTransfer
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&transfer).Error; err != nil {
		return nil, nil, err
	}
	var items []modelsOrder.StockTransferItem
	if err := r.db.WithContext(ctx).Where("transfer_id = ?", id).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	return &transfer, items, nil
}

// UpdateStatus updates a transfer's status and executes or reverses stock movement.
func (r *StockTransferRepository) UpdateStatus(ctx context.Context, id, status, operatorID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var transfer modelsOrder.StockTransfer
		if err := tx.Where("id = ?", id).First(&transfer).Error; err != nil {
			return err
		}
		if transfer.Status == modelsOrder.StockTransferStatusCompleted || transfer.Status == modelsOrder.StockTransferStatusCancelled {
			return fmt.Errorf("transfer %s is already in terminal status %s", id, transfer.Status)
		}
		now := time.Now()

		if status == modelsOrder.StockTransferStatusCompleted {
			var items []modelsOrder.StockTransferItem
			if err := tx.Where("transfer_id = ?", id).Find(&items).Error; err != nil {
				return err
			}
			var allAudit []*modelsOrder.StockTransaction
			for _, item := range items {
				recs, err := TransferStockBetweenWarehouses(tx,
					transfer.FromWarehouseID, transfer.ToWarehouseID,
					item.ProductID, item.Quantity,
					modelsOrder.StockReasonStockTransfer, transfer.ID, operatorID, now,
				)
				if err != nil {
					return err
				}
				allAudit = append(allAudit, recs...)
				// Update received quantity
				if ue := tx.Model(&item).Update("received_qty", item.Quantity).Error; ue != nil {
					log.Printf("stock_transfer: received_qty update failed for item %d: %v", item.ID, ue)
				}
			}
			if err := writeStockAuditEntries(tx, allAudit); err != nil {
				return err
			}
		}

		return tx.Model(&transfer).Updates(map[string]interface{}{
			"status":     status,
			"updated_at": now,
		}).Error
	})
}

// addStockToWarehouse is a helper that adds quantity to a warehouse stock row.
func addStockToWarehouse(tx *gorm.DB, warehouseID, productID string, qty int) error {
	return tx.Model(&modelsProduct.WarehouseStock{}).
		Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
		Update("quantity", gorm.Expr("quantity + ?", qty)).Error
}
