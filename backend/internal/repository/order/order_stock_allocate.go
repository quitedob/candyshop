package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"log"
	"gorm.io/gorm/clause"
)

// sumActiveOEMHolds 统计 OEM 项目对 SKU 的活跃预留量
func sumActiveOEMHolds(tx *gorm.DB, productID string) (int, error) {
	var sum int64
	if err := tx.Model(&modelsProduct.OEMProjectInventoryHold{}).
		Where("product_id = ? AND status = ?", productID, "active").
		Select("COALESCE(SUM(quantity),0)").
		Scan(&sum).Error; err != nil {
		return 0, err
	}
	return int(sum), nil
}

// resolveWarehouseID resolves a warehouse ID. If warehouseID is provided and non-empty,
// validates it exists and is active. Otherwise falls back to the default warehouse.
func resolveWarehouseID(tx *gorm.DB, warehouseID string) (string, error) {
	if warehouseID != "" {
		var w modelsProduct.Warehouse
		if err := tx.Where("id = ? AND is_active = ?", warehouseID, true).First(&w).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", fmt.Errorf("warehouse %s not found or inactive", warehouseID)
			}
			return "", err
		}
		return w.ID, nil
	}
	return resolveDefaultWarehouseID(tx, "")
}

// resolveDefaultWarehouseID resolves the best warehouse for stock operations.
// If warehouseID is provided and active, it is used directly.
// Otherwise falls back to the default warehouse (is_default=true, then code=MAIN).
func resolveDefaultWarehouseID(tx *gorm.DB, warehouseID string) (string, error) {
	if warehouseID != "" {
		var w modelsProduct.Warehouse
		if err := tx.Where("id = ? AND is_active = ?", warehouseID, true).First(&w).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return "", fmt.Errorf("warehouse %s not found or inactive", warehouseID)
			}
			return "", err
		}
		return w.ID, nil
	}
	var w modelsProduct.Warehouse
	if err := tx.Where("is_default = ? AND is_active = ?", true, true).First(&w).Error; err == nil {
		return w.ID, nil
	}
	if err := tx.Where("code = ? AND is_active = ?", "MAIN", true).First(&w).Error; err == nil {
		return w.ID, nil
	}
	return "", gorm.ErrRecordNotFound
}

func countWarehouseStockRows(tx *gorm.DB, productID string) (int64, error) {
	var n int64
	err := tx.Model(&modelsProduct.WarehouseStock{}).Where("product_id = ?", productID).Count(&n).Error
	return n, err
}

// reserveWarehouseStock reserves stock in a warehouse (increment Reserved, decrement Product.StockQuantity).
// Used when an order is confirmed but goods haven't shipped yet.
func reserveWarehouseStock(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	var ws modelsProduct.WarehouseStock
	err := tx.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(&ws).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		var p modelsProduct.Product
		if perr := tx.Where("id = ?", productID).First(&p).Error; perr != nil {
			return nil, perr
		}
		ws = modelsProduct.WarehouseStock{
			WarehouseID: warehouseID,
			ProductID:   productID,
			Quantity:    p.StockQuantity,
			Reserved:    0,
			UpdatedAt:   t,
		}
		if cerr := tx.Create(&ws).Error; cerr != nil {
			return nil, cerr
		}
	} else if err != nil {
		return nil, err
	}
	sellableWh := ws.Quantity - ws.Reserved
	if sellableWh < 0 {
		return nil, fmt.Errorf("%w: negative sellable warehouse stock for product %s", modelsOrder.ErrInsufficientStock, productID)
	}
	if sellableWh < qty {
		return nil, fmt.Errorf("%w: insufficient warehouse sellable stock for product %s", modelsOrder.ErrInsufficientStock, productID)
	}
	res := tx.Model(&modelsProduct.WarehouseStock{}).
		Where("id = ? AND (quantity - reserved) >= ?", ws.ID, qty).
		Update("reserved", gorm.Expr("reserved + ?", qty))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("%w: insufficient warehouse stock for product %s (concurrent)", modelsOrder.ErrInsufficientStock, productID)
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
		return nil, fmt.Errorf("%w: insufficient aggregate stock_quantity for product %s", modelsOrder.ErrInsufficientStock, productID)
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

// releaseWarehouseStock releases previously reserved stock (decrement Reserved, increment Product.StockQuantity).
func releaseWarehouseStock(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	var ws modelsProduct.WarehouseStock
	err := tx.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(&ws).Error
	if err != nil {
		return nil, err
	}
	if ws.Reserved < qty {
		return nil, fmt.Errorf("cannot release more than reserved for product %s (reserved=%d, release=%d)", productID, ws.Reserved, qty)
	}
	res := tx.Model(&modelsProduct.WarehouseStock{}).
		Where("id = ? AND reserved >= ?", ws.ID, qty).
		Update("reserved", gorm.Expr("reserved - ?", qty))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient reserved stock for product %s (concurrent)", productID)
	}

	var p modelsProduct.Product
	if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	beforeP := p.StockQuantity
	res2 := tx.Model(&modelsProduct.Product{}).
		Where("id = ?", productID).
		Update("stock_quantity", gorm.Expr("stock_quantity + ?", qty))
	if res2.Error != nil {
		return nil, res2.Error
	}
	afterP := beforeP + qty
	wid := warehouseID
	rec := &modelsOrder.StockTransaction{
		ProductID:   productID,
		Change:      qty,
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

// deductWarehouseStock deducts physical stock from a warehouse (decrement both Quantity and Reserved).
// Used when goods are actually shipped/picked.
func deductWarehouseStock(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	var ws modelsProduct.WarehouseStock
	err := tx.Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).First(&ws).Error
	if err != nil {
		return nil, err
	}
	if ws.Quantity < qty {
		return nil, fmt.Errorf("insufficient warehouse quantity for product %s", productID)
	}
	res := tx.Model(&modelsProduct.WarehouseStock{}).
		Where("id = ? AND quantity >= ?", ws.ID, qty).
		Updates(map[string]interface{}{
			"quantity": gorm.Expr("quantity - ?", qty),
			"reserved": gorm.Expr("GREATEST(reserved - ?, 0)", qty),
		})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient warehouse stock for product %s (concurrent)", productID)
	}

	var p modelsProduct.Product
	if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	beforeP := p.StockQuantity
	wid := warehouseID
	rec := &modelsOrder.StockTransaction{
		ProductID:   productID,
		Change:      -qty,
		StockBefore: beforeP,
		StockAfter:  beforeP,
		Reason:      reason,
		ReferenceID: refID,
		OperatorID:  operatorID,
		WarehouseID: &wid,
		CreatedAt:   t,
	}
	return []*modelsOrder.StockTransaction{rec}, nil
}

// applyWarehouseStockChange 按 reason 路由到三阶段库存操作（预留/释放/扣减）
func applyWarehouseStockChange(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	switch reason {
	case modelsOrder.StockReasonStockReleased, modelsOrder.StockReasonOrderCancelled, modelsOrder.StockReasonOrderDeleted, modelsOrder.StockReasonDraftExpired:
		return releaseWarehouseStock(tx, warehouseID, productID, qty, reason, refID, operatorID, t)
	case modelsOrder.StockReasonDispatched, modelsOrder.StockReasonGoodsIssued, modelsOrder.StockReasonStockDeducted:
		return deductWarehouseStock(tx, warehouseID, productID, qty, reason, refID, operatorID, t)
	default:
		return reserveWarehouseStock(tx, warehouseID, productID, qty, reason, refID, operatorID, t)
	}
}

// deductStockForProductLine 扣减库存：若存在 ProductBatch 则按 FEFO；否则多仓（有仓记录时走默认仓）或回退单仓字段
func deductStockForProductLine(tx *gorm.DB, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	if qty <= 0 {
		return nil, nil
	}
	var batchCount int64
	if err := tx.Model(&modelsProduct.ProductBatch{}).Where("product_id = ?", productID).Count(&batchCount).Error; err != nil {
		return nil, err
	}
	if batchCount == 0 {
		return deductLegacyProductStock(tx, productID, qty, reason, refID, operatorID, t)
	}
	return deductFEFOFromBatches(tx, productID, qty, reason, refID, operatorID, t)
}

// deductStockForProductLineWithWarehouse is like deductStockForProductLine but supports warehouse routing.
func deductStockForProductLineWithWarehouse(tx *gorm.DB, productID, warehouseID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	if qty <= 0 {
		return nil, nil
	}
	wid, err := resolveWarehouseID(tx, warehouseID)
	if err != nil {
		return nil, err
	}
	var batchCount int64
	if err := tx.Model(&modelsProduct.ProductBatch{}).Where("product_id = ? AND warehouse_id = ?", productID, wid).Count(&batchCount).Error; err != nil {
		return nil, err
	}
	if batchCount == 0 {
		return deductLegacyProductStockWithWarehouse(tx, wid, productID, qty, reason, refID, operatorID, t)
	}
	return deductFEFOFromBatchesWithWarehouse(tx, wid, productID, qty, reason, refID, operatorID, t)
}

func deductLegacyProductStock(tx *gorm.DB, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	var p modelsProduct.Product
	if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	holds, err := sumActiveOEMHolds(tx, productID)
	if err != nil {
		return nil, err
	}
	if p.StockQuantity-holds < qty {
		return nil, fmt.Errorf("%w for product %s after OEM holds", modelsOrder.ErrInsufficientStock, productID)
	}
	nWh, err := countWarehouseStockRows(tx, productID)
	if err != nil {
		return nil, err
	}
	if nWh > 0 {
		wid, werr := resolveDefaultWarehouseID(tx, "")
		if werr == nil {
			return applyWarehouseStockChange(tx, wid, productID, qty, reason, refID, operatorID, t)
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
		return nil, fmt.Errorf("%w for product %s during reservation", modelsOrder.ErrInsufficientStock, productID)
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

func deductLegacyProductStockWithWarehouse(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	var p modelsProduct.Product
	if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	holds, err := sumActiveOEMHolds(tx, productID)
	if err != nil {
		return nil, err
	}
	if p.StockQuantity-holds < qty {
		return nil, fmt.Errorf("%w for product %s after OEM holds", modelsOrder.ErrInsufficientStock, productID)
	}
	nWh, err := countWarehouseStockRows(tx, productID)
	if err != nil {
		return nil, err
	}
	if nWh > 0 {
		return applyWarehouseStockChange(tx, warehouseID, productID, qty, reason, refID, operatorID, t)
	}
	before := p.StockQuantity
	res := tx.Model(&modelsProduct.Product{}).
		Where("id = ? AND stock_quantity >= ?", productID, qty).
		Update("stock_quantity", gorm.Expr("stock_quantity - ?", qty))
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, fmt.Errorf("%w for product %s during reservation", modelsOrder.ErrInsufficientStock, productID)
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

// restoreFEFOBatches reverses batch-level FEFO reservations for an order. The
// reserve path (deductFEFOFromBatches) permanently decrements
// product_batches.quantity; without a matching restore, cancelling an order
// brings back only the aggregate products.stock_quantity and leaves batch
// quantities depleted — a later FEFO order then fails "insufficient batch
// stock" despite aggregate stock being available (M1). It restores exactly the
// batches recorded in the reservation's stock_transaction rows, so released
// quantities return to the same lots they were taken from.
func restoreFEFOBatches(tx *gorm.DB, productID, refID string) error {
	type batchQty struct {
		BatchID string
		Qty     int
	}
	var rows []batchQty
	if err := tx.Model(&modelsOrder.StockTransaction{}).
		Select("batch_id, COALESCE(SUM(ABS(change)),0) AS qty").
		Where("reference_id = ? AND product_id = ? AND reason = ? AND batch_id IS NOT NULL AND change < 0",
			refID, productID, modelsOrder.StockReasonStockReserved).
		Group("batch_id").
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, r := range rows {
		if r.BatchID == "" || r.Qty <= 0 {
			continue
		}
		if err := tx.Model(&modelsProduct.ProductBatch{}).
			Where("id = ?", r.BatchID).
			Update("quantity", gorm.Expr("quantity + ?", r.Qty)).Error; err != nil {
			return err
		}
	}
	return nil
}

// restoreLegacyProductStock 释放库存：若存在仓级记录则同步增加默认仓行与 product 汇总
func restoreLegacyProductStock(tx *gorm.DB, productID string, qty int) error {
	if qty <= 0 {
		return nil
	}
	nWh, err := countWarehouseStockRows(tx, productID)
	if err != nil {
		return err
	}
	if nWh > 0 {
		if wid, werr := resolveDefaultWarehouseID(tx, ""); werr == nil {
			if err := tx.Model(&modelsProduct.WarehouseStock{}).
				Where("warehouse_id = ? AND product_id = ?", wid, productID).
				Update("quantity", gorm.Expr("quantity + ?", qty)).Error; err != nil {
				return err
			}
		}
	}
	return tx.Model(&modelsProduct.Product{}).
		Where("id = ?", productID).
		Update("stock_quantity", gorm.Expr("stock_quantity + ?", qty)).Error
}

// restoreLegacyProductStockWithWarehouse restores stock to a specific warehouse.
func restoreLegacyProductStockWithWarehouse(tx *gorm.DB, warehouseID, productID string, qty int) error {
	if qty <= 0 {
		return nil
	}
	nWh, err := countWarehouseStockRows(tx, productID)
	if err != nil {
		return err
	}
	if nWh > 0 {
		if err := tx.Model(&modelsProduct.WarehouseStock{}).
			Where("warehouse_id = ? AND product_id = ?", warehouseID, productID).
			Update("quantity", gorm.Expr("quantity + ?", qty)).Error; err != nil {
			return err
		}
	}
	return tx.Model(&modelsProduct.Product{}).
		Where("id = ?", productID).
		Update("stock_quantity", gorm.Expr("stock_quantity + ?", qty)).Error
}

func deductFEFOFromBatches(tx *gorm.DB, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	now := t.UTC()
	var sum int64
	if err := tx.Model(&modelsProduct.ProductBatch{}).
		Where("product_id = ? AND quantity > 0 AND is_expired = ? AND expiry_date >= ?", productID, false, now).
		Select("COALESCE(SUM(quantity),0)").
		Scan(&sum).Error; err != nil {
		return nil, err
	}
	holds, err := sumActiveOEMHolds(tx, productID)
	if err != nil {
		return nil, err
	}
	if int(sum)-holds < qty {
		return nil, fmt.Errorf("%w: insufficient batch stock for product %s (FEFO)", modelsOrder.ErrInsufficientStock, productID)
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
				return nil, fmt.Errorf("%w: insufficient batch stock for product %s (concurrent reservation)", modelsOrder.ErrInsufficientStock, productID)
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
		return nil, fmt.Errorf("%w: insufficient aggregate stock_quantity for product %s", modelsOrder.ErrInsufficientStock, productID)
	}

	// Also update warehouse_stock for consistency (mirror legacy path behavior)
	nWh, whErr := countWarehouseStockRows(tx, productID)
	if whErr == nil && nWh > 0 {
		if wid, wErr := resolveDefaultWarehouseID(tx, ""); wErr == nil {
			if ue := tx.Model(&modelsProduct.WarehouseStock{}).
				Where("warehouse_id = ? AND product_id = ? AND quantity >= ?", wid, productID, qty).
				Update("quantity", gorm.Expr("quantity - ?", qty)).Error; ue != nil {
				log.Printf("order_stock_allocate: warehouse stock update failed for product %s: %v", productID, ue)
			}
		}
	}

	return records, nil
}

func deductFEFOFromBatchesWithWarehouse(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	now := t.UTC()
	var sum int64
	if err := tx.Model(&modelsProduct.ProductBatch{}).
		Where("product_id = ? AND warehouse_id = ? AND quantity > 0 AND is_expired = ? AND expiry_date >= ?", productID, warehouseID, false, now).
		Select("COALESCE(SUM(quantity),0)").
		Scan(&sum).Error; err != nil {
		return nil, err
	}
	holds, err := sumActiveOEMHolds(tx, productID)
	if err != nil {
		return nil, err
	}
	if int(sum)-holds < qty {
		return nil, fmt.Errorf("%w: insufficient batch stock for product %s in warehouse %s (FEFO)", modelsOrder.ErrInsufficientStock, productID, warehouseID)
	}

	remaining := qty
	var records []*modelsOrder.StockTransaction
	for remaining > 0 {
		var b modelsProduct.ProductBatch
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("product_id = ? AND warehouse_id = ? AND quantity > 0 AND is_expired = ? AND expiry_date >= ?", productID, warehouseID, false, now).
			Order("expiry_date ASC, id ASC").
			Take(&b).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("%w: insufficient batch stock for product %s in warehouse %s (concurrent)", modelsOrder.ErrInsufficientStock, productID, warehouseID)
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
		return nil, fmt.Errorf("%w: insufficient aggregate stock_quantity for product %s", modelsOrder.ErrInsufficientStock, productID)
	}

	if ue := tx.Model(&modelsProduct.WarehouseStock{}).
		Where("warehouse_id = ? AND product_id = ? AND quantity >= ?", warehouseID, productID, qty).
		Update("quantity", gorm.Expr("quantity - ?", qty)).Error; ue != nil {
		log.Printf("order_stock_allocate: warehouse stock update failed for product %s: %v", productID, ue)
	}

	return records, nil
}

// TransferStockBetweenWarehouses moves stock from one warehouse to another within a transaction.
func TransferStockBetweenWarehouses(tx *gorm.DB, fromWarehouseID, toWarehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	if qty <= 0 {
		return nil, nil
	}
	// Deduct from source warehouse (decrement both quantity and reserved if any)
	fromRecs, err := deductWarehouseStock(tx, fromWarehouseID, productID, qty, reason, refID, operatorID, t)
	if err != nil {
		return nil, fmt.Errorf("transfer from warehouse: %w", err)
	}
	// Add to target warehouse
	if err := tx.Model(&modelsProduct.WarehouseStock{}).
		Where("warehouse_id = ? AND product_id = ?", toWarehouseID, productID).
		Update("quantity", gorm.Expr("quantity + ?", qty)).Error; err != nil {
		return nil, fmt.Errorf("transfer to warehouse: %w", err)
	}
	return fromRecs, nil
}
