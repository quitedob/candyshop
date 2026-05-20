package product

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"log"
)

// SupplierRepository handles supplier and purchase order operations.
type SupplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new SupplierRepository.
func NewSupplierRepository(db *gorm.DB) *SupplierRepository {
	return &SupplierRepository{db: db}
}

// ---- Suppliers ----

func (r *SupplierRepository) CreateSupplier(ctx context.Context, s *modelsProduct.Supplier) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SupplierRepository) FindAllSuppliers(ctx context.Context, page, limit int) ([]modelsProduct.Supplier, int64, error) {
	var suppliers []modelsProduct.Supplier
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsProduct.Supplier{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&suppliers).Error; err != nil {
		return nil, 0, err
	}
	return suppliers, total, nil
}

func (r *SupplierRepository) FindSupplierByID(ctx context.Context, id string) (*modelsProduct.Supplier, error) {
	var s modelsProduct.Supplier
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&s).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SupplierRepository) UpdateSupplier(ctx context.Context, s *modelsProduct.Supplier) error {
	s.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *SupplierRepository) DeleteSupplier(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&modelsProduct.Supplier{}).Where("id = ?", id).
		Update("is_active", false).Error
}

func (r *SupplierRepository) FindByAPIKey(ctx context.Context, key string) (*modelsProduct.Supplier, error) {
	var s modelsProduct.Supplier
	err := r.db.WithContext(ctx).Where("api_key = ? AND is_active = true", key).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SupplierRepository) CountSuppliers(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&modelsProduct.Supplier{}).Count(&count).Error
	return count, err
}

func (r *SupplierRepository) FindPOsBySupplierID(ctx context.Context, supplierID string) ([]modelsProduct.PurchaseOrder, error) {
	var pos []modelsProduct.PurchaseOrder
	err := r.db.WithContext(ctx).Where("supplier_id = ?", supplierID).Order("created_at DESC").Find(&pos).Error
	return pos, err
}

func (r *SupplierRepository) UpdatePOStatus(ctx context.Context, poID, supplierID, status string) error {
	return r.db.WithContext(ctx).Model(&modelsProduct.PurchaseOrder{}).
		Where("id = ? AND supplier_id = ?", poID, supplierID).
		Updates(map[string]interface{}{"status": status, "updated_at": time.Now()}).Error
}

// ---- Purchase Orders ----

func (r *SupplierRepository) CreatePO(ctx context.Context, po *modelsProduct.PurchaseOrder, items []modelsProduct.PurchaseOrderItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(po).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].POID = po.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *SupplierRepository) FindAllPOs(ctx context.Context, page, limit int) ([]modelsProduct.PurchaseOrder, int64, error) {
	var pos []modelsProduct.PurchaseOrder
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsProduct.PurchaseOrder{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&pos).Error; err != nil {
		return nil, 0, err
	}
	return pos, total, nil
}

func (r *SupplierRepository) FindPOByID(ctx context.Context, id string) (*modelsProduct.PurchaseOrder, []modelsProduct.PurchaseOrderItem, error) {
	var po modelsProduct.PurchaseOrder
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&po).Error; err != nil {
		return nil, nil, err
	}
	var items []modelsProduct.PurchaseOrderItem
	if err := r.db.WithContext(ctx).Where("po_id = ?", id).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	return &po, items, nil
}

// ReceivePO marks a purchase order as received and adds stock to the warehouse.
func (r *SupplierRepository) ReceivePO(ctx context.Context, id, warehouseID string, receivedItems map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var po modelsProduct.PurchaseOrder
		if err := tx.Where("id = ?", id).First(&po).Error; err != nil {
			return err
		}
		if po.Status == modelsProduct.POStatusReceived || po.Status == modelsProduct.POStatusCancelled {
			return fmt.Errorf("PO is already %s", po.Status)
		}
		// Load PO items to get UnitCost for each product
		var poItems []modelsProduct.PurchaseOrderItem
		if err := tx.Where("po_id = ?", id).Find(&poItems).Error; err != nil {
			return err
		}
		itemCosts := make(map[string]float64, len(poItems))
		for _, item := range poItems {
			itemCosts[item.ProductID] = item.UnitCost
		}

		now := time.Now()
		allReceived := true
		for productID, qty := range receivedItems {
			res := tx.Model(&modelsProduct.PurchaseOrderItem{}).
				Where("po_id = ? AND product_id = ?", id, productID).
				Update("received_qty", gorm.Expr("received_qty + ?", qty))
			if res.Error != nil {
				return res.Error
			}
			// Add stock to warehouse
			if warehouseID != "" {
				if err := tx.Model(&modelsProduct.WarehouseStock{}).
					Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
					Update("quantity", gorm.Expr("quantity + ?", qty)).Error; err != nil {
					return err
				}
				// Also update product aggregate stock
				if se := tx.Model(&modelsProduct.Product{}).
					Where("id = ?", productID).
					Update("stock_quantity", gorm.Expr("stock_quantity + ?", qty)).Error; se != nil {
					log.Printf("supplier: stock update failed for product %s: %v", productID, se)
				}
				// Create ProductBatch with UnitCost from PO item
				unitCost := itemCosts[productID]
				batchNumber := fmt.Sprintf("%s-%s-%d", id, productID, time.Now().UnixNano())
				if createErr := tx.Create(&modelsProduct.ProductBatch{
					ID:          batchNumber,
					ProductID:   productID,
					WarehouseID: warehouseID,
					BatchNumber: batchNumber,
					Quantity:    qty,
					UnitCost:    unitCost,
					CreatedAt:   now,
					UpdatedAt:   now,
				}).Error; createErr != nil {
					log.Printf("supplier: failed to create batch for product %s: %v", productID, createErr)
				}
				recomputeWeightedAvgCost(tx, productID)

				// Stock audit
				_ = tx.Create(&modelsOrder.StockTransaction{
					ProductID:   productID,
					Change:      qty,
					StockBefore: 0,
					StockAfter:  qty,
					Reason:      modelsOrder.StockReasonGoodsIssued,
					ReferenceID: id,
					OperatorID:  "system",
					CreatedAt:   now,
				})
			}
		}

		// Check if all items are fully received
		var items []modelsProduct.PurchaseOrderItem
		tx.Where("po_id = ?", id).Find(&items)
		for _, item := range items {
			if item.ReceivedQty < item.Quantity {
				allReceived = false
				break
			}
		}
		newStatus := modelsProduct.POStatusPartial
		if allReceived {
			newStatus = modelsProduct.POStatusReceived
		}
		return tx.Model(&po).Updates(map[string]interface{}{
			"status":        newStatus,
			"received_date": now,
			"updated_at":    now,
		}).Error
	})
}


// recomputeWeightedAvgCost calculates the weighted average unit cost for a product
// from all non-expired batches and updates Product.WeightedAvgCost.
func recomputeWeightedAvgCost(tx *gorm.DB, productID string) {
	var batches []modelsProduct.ProductBatch
	if err := tx.Where("product_id = ? AND quantity > 0 AND is_expired = false", productID).Find(&batches).Error; err != nil {
		return
	}
	var totalCost float64
	var totalQty int
	for _, b := range batches {
		totalCost += float64(b.Quantity) * b.UnitCost
		totalQty += b.Quantity
	}
	if totalQty == 0 {
		return
	}
	avgCost := totalCost / float64(totalQty)
	tx.Model(&modelsProduct.Product{}).Where("id = ?", productID).Update("weighted_avg_cost", avgCost)
}

// ComputeWeightedAvgCost returns the weighted average unit cost for a product
// across all non-expired, in-stock batches. Returns 0 if no batches exist.
func (r *SupplierRepository) ComputeWeightedAvgCost(ctx context.Context, productID string) float64 {
	var batches []modelsProduct.ProductBatch
	if err := r.db.WithContext(ctx).Where("product_id = ? AND quantity > 0 AND is_expired = false", productID).Find(&batches).Error; err != nil {
		return 0
	}
	var totalCost float64
	var totalQty int
	for _, b := range batches {
		totalCost += float64(b.Quantity) * b.UnitCost
		totalQty += b.Quantity
	}
	if totalQty == 0 {
		return 0
	}
	return totalCost / float64(totalQty)
}
