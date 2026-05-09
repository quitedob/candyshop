package trade

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// --- Stock deduction functions for logistics dispatch ---
// These mirror repository/order/order_stock_allocate.go but accept a configurable reason
// to avoid cross-package circular dependencies. A future refactor should extract these
// into a shared internal/repository/stock/ package.

func logisticsSumActiveOEMHolds(tx *gorm.DB, productID string) (int, error) {
	var sum int64
	if err := tx.Model(&modelsProduct.OEMProjectInventoryHold{}).
		Where("product_id = ? AND status = ?", productID, "active").
		Select("COALESCE(SUM(quantity),0)").
		Scan(&sum).Error; err != nil {
		return 0, err
	}
	return int(sum), nil
}

func logisticsResolveDefaultWarehouseID(tx *gorm.DB) (string, error) {
	var w modelsProduct.Warehouse
	if err := tx.Where("is_default = ? AND is_active = ?", true, true).First(&w).Error; err == nil {
		return w.ID, nil
	}
	if err := tx.Where("code = ? AND is_active = ?", "MAIN", true).First(&w).Error; err == nil {
		return w.ID, nil
	}
	return "", gorm.ErrRecordNotFound
}

func logisticsCountWarehouseStockRows(tx *gorm.DB, productID string) (int64, error) {
	var n int64
	err := tx.Model(&modelsProduct.WarehouseStock{}).Where("product_id = ?", productID).Count(&n).Error
	return n, err
}

func logisticsDeductFromDefaultWarehouse(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	var ws modelsProduct.WarehouseStock
	err := tx.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(&ws).Error
	if err != nil {
		return nil, err
	}
	sellableWh := ws.Quantity - ws.Reserved
	if sellableWh < 0 {
		sellableWh = 0
	}
	if sellableWh < qty {
		return nil, fmt.Errorf("insufficient warehouse sellable stock for product %s during dispatch", productID)
	}
	res := tx.Model(&modelsProduct.WarehouseStock{}).
		Where("id = ? AND (quantity - reserved) >= ?", ws.ID, qty).
		Update("quantity", gorm.Expr("quantity - ?", qty))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient warehouse stock for product %s (concurrent dispatch)", productID)
	}
	var p modelsProduct.Product
	if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	beforeP := p.StockQuantity
	res2 := tx.Model(&modelsProduct.Product{}).
		Where("id = ? AND stock_quantity >= ?", productID, qty).
		Update("stock_quantity", gorm.Expr("stock_quantity - ?", qty))
	if res2.Error != nil {
		return nil, res2.Error
	}
	if res2.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient aggregate stock_quantity for product %s during dispatch", productID)
	}
	afterP := beforeP - qty
	wid := warehouseID
	rec := &modelsOrder.StockTransaction{
		ProductID:   productID,
		Change:      -qty,
		StockBefore: beforeP,
		StockAfter:  afterP,
		Reason:      reason,
		ReferenceID: refID,
		OperatorID:  operatorID,
		WarehouseID: &wid,
		CreatedAt:   t,
	}
	return []*modelsOrder.StockTransaction{rec}, nil
}

func logisticsDeductLegacyProductStock(tx *gorm.DB, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	var p modelsProduct.Product
	if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	holds, err := logisticsSumActiveOEMHolds(tx, productID)
	if err != nil {
		return nil, err
	}
	if p.StockQuantity-holds < qty {
		return nil, fmt.Errorf("insufficient sellable stock for product %s after OEM holds (dispatch)", productID)
	}
	nWh, err := logisticsCountWarehouseStockRows(tx, productID)
	if err != nil {
		return nil, err
	}
	if nWh > 0 {
		wid, werr := logisticsResolveDefaultWarehouseID(tx)
		if werr == nil {
			return logisticsDeductFromDefaultWarehouse(tx, wid, productID, qty, reason, refID, operatorID, t)
		}
	}
	before := p.StockQuantity
	res := tx.Model(&modelsProduct.Product{}).
		Where("id = ? AND stock_quantity >= ?", productID, qty).
		Update("stock_quantity", gorm.Expr("stock_quantity - ?", qty))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("%w for product %s during dispatch", modelsOrder.ErrInsufficientStock, productID)
	}
	after := before - qty
	rec := &modelsOrder.StockTransaction{
		ProductID:   productID,
		Change:      -qty,
		StockBefore: before,
		StockAfter:  after,
		Reason:      reason,
		ReferenceID: refID,
		OperatorID:  operatorID,
		CreatedAt:   t,
	}
	return []*modelsOrder.StockTransaction{rec}, nil
}

func logisticsDeductFEFOFromBatches(tx *gorm.DB, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	now := t.UTC()
	var sum int64
	if err := tx.Model(&modelsProduct.ProductBatch{}).
		Where("product_id = ? AND quantity > 0 AND is_expired = ? AND expiry_date >= ?", productID, false, now).
		Select("COALESCE(SUM(quantity),0)").
		Scan(&sum).Error; err != nil {
		return nil, err
	}
	holds, err := logisticsSumActiveOEMHolds(tx, productID)
	if err != nil {
		return nil, err
	}
	if int(sum)-holds < qty {
		return nil, fmt.Errorf("insufficient batch stock for product %s (FEFO dispatch)", productID)
	}

	remaining := qty
	var records []*modelsOrder.StockTransaction
	for remaining > 0 {
		var b modelsProduct.ProductBatch
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("product_id = ? AND quantity > 0 AND is_expired = ? AND expiry_date >= ?", productID, false, now).
			Order("expiry_date ASC, id ASC").
			Take(&b).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("insufficient batch stock for product %s (concurrent dispatch)", productID)
			}
			return nil, err
		}
		take := b.Quantity
		if take > remaining {
			take = remaining
		}
		beforeBatch := b.Quantity
		afterBatch := beforeBatch - take
		res := tx.Model(&modelsProduct.ProductBatch{}).
			Where("id = ? AND quantity >= ?", b.ID, take).
			Update("quantity", gorm.Expr("quantity - ?", take))
		if res.Error != nil {
			return nil, res.Error
		}
		if res.RowsAffected == 0 {
			continue
		}
		bid := b.ID
		lot := b.BatchNumber
		records = append(records, &modelsOrder.StockTransaction{
			ProductID:   productID,
			Change:      -take,
			StockBefore: beforeBatch,
			StockAfter:  afterBatch,
			Reason:      reason,
			ReferenceID: refID,
			OperatorID:  operatorID,
			BatchID:     &bid,
			LotNumber:   lot,
			CreatedAt:   t,
		})
		remaining -= take
	}

	res := tx.Model(&modelsProduct.Product{}).
		Where("id = ? AND stock_quantity >= ?", productID, qty).
		Update("stock_quantity", gorm.Expr("stock_quantity - ?", qty))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient aggregate stock_quantity for product %s during dispatch", productID)
	}
	return records, nil
}

func logisticsDeductStockForProductLine(tx *gorm.DB, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	if qty <= 0 {
		return nil, nil
	}
	var batchCount int64
	if err := tx.Model(&modelsProduct.ProductBatch{}).Where("product_id = ?", productID).Count(&batchCount).Error; err != nil {
		return nil, err
	}
	if batchCount == 0 {
		return logisticsDeductLegacyProductStock(tx, productID, qty, reason, refID, operatorID, t)
	}
	return logisticsDeductFEFOFromBatches(tx, productID, qty, reason, refID, operatorID, t)
}
