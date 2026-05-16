package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"context"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OrderRepository handles order data operations
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// FindAll returns paginated orders with optional filters.
func (r *OrderRepository) FindAll(ctx context.Context, page, limit int, status, userID, dateFrom, dateTo string) ([]modelsOrder.Order, int64, error) {
	var orders []modelsOrder.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsOrder.Order{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Preload("User").Preload("Inquiry").Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// FindByID returns an order by ID
func (r *OrderRepository) FindByID(ctx context.Context, id string) (*modelsOrder.Order, error) {
	var order modelsOrder.Order
	if err := r.db.WithContext(ctx).Preload("User").Preload("Inquiry").Where("id = ?", id).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// FindByUserID returns paginated orders for a specific user
func (r *OrderRepository) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsOrder.Order, int64, error) {
	var orders []modelsOrder.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsOrder.Order{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

// Create creates a new order
func (r *OrderRepository) Create(ctx context.Context, order *modelsOrder.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// CreateWithStockReservation creates an order and atomically reserves stock (three-phase: reserve on create, deduct on ship).
func (r *OrderRepository) CreateWithStockReservation(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var all []*modelsOrder.StockTransaction
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			recs, err := reserveStockForOrderLine(tx, warehouseIDFromOrder(order), productID, qty, modelsOrder.StockReasonStockReserved, order.ID, order.UserID, now)
			if err != nil {
				return err
			}
			all = append(all, recs...)
		}
		order.StockReserved = true
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return writeStockAuditEntries(tx, all)
	})
}

// Update updates an order
func (r *OrderRepository) Update(ctx context.Context, order *modelsOrder.Order) error {
	return r.db.WithContext(ctx).Omit("User", "Inquiry").Save(order).Error
}

// UpdateWithStockAdjustment adjusts stock atomically during order edits.
// Positive delta reserves additional stock (three-phase); negative delta releases reserved stock.
func (r *OrderRepository) UpdateWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var extra []*modelsOrder.StockTransaction
		for productID, delta := range stockDeltas {
			switch {
			case delta > 0:
				recs, err := reserveStockForOrderLine(tx, warehouseIDFromOrder(order), productID, delta, modelsOrder.StockReasonStockReserved, order.ID, order.UserID, now)
				if err != nil {
					return err
				}
				extra = append(extra, recs...)
			case delta < 0:
				releaseQty := -delta
				recs, err := releaseStockForOrderLine(tx, warehouseIDFromOrder(order), productID, releaseQty, modelsOrder.StockReasonStockReleased, order.ID, order.UserID, now)
				if err != nil {
					return err
				}
				extra = append(extra, recs...)
			}
		}
		if err := writeStockAuditEntries(tx, extra); err != nil {
			return err
		}
		return tx.Save(order).Error
	})
}

// Delete deletes an order
func (r *OrderRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsOrder.Order{}, "id = ?", id).Error
}

// ReserveStockForOrder atomically reserves stock for an existing order (increment Reserved, not deduct Quantity).
// When warehouse stock rows exist, increments WarehouseStock.Reserved and decrements Product.StockQuantity.
// Falls back to legacy direct deduction when no warehouse rows exist.
func (r *OrderRepository) ReserveStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", order.ID).First(order).Error; err != nil {
			return err
		}
		if order.StockReserved {
			return nil
		}

		now := time.Now()
		var all []*modelsOrder.StockTransaction
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			recs, err := reserveStockForOrderLine(tx, warehouseIDFromOrder(order), productID, qty, modelsOrder.StockReasonStockReserved, order.ID, order.UserID, now)
			if err != nil {
				return err
			}
			all = append(all, recs...)
		}

		order.StockReserved = true
		order.UpdatedAt = now
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return writeStockAuditEntries(tx, all)
	})
}

// ReleaseStockForOrder releases previously reserved stock (decrement Reserved, increment Product.StockQuantity).
func (r *OrderRepository) ReleaseStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", order.ID).First(order).Error; err != nil {
			return err
		}
		if !order.StockReserved {
			return nil
		}

		now := time.Now()
		var all []*modelsOrder.StockTransaction
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			recs, err := releaseStockForOrderLine(tx, warehouseIDFromOrder(order), productID, qty, modelsOrder.StockReasonStockReleased, order.ID, order.UserID, now)
			if err != nil {
				return err
			}
			all = append(all, recs...)
		}
		order.StockReserved = false
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		return writeStockAuditEntries(tx, all)
	})
}

// DeleteWithStockRestore releases reserved stock and deletes the order in one transaction.
func (r *OrderRepository) DeleteWithStockRestore(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			if _, err := releaseStockForOrderLine(tx, warehouseIDFromOrder(order), productID, qty, modelsOrder.StockReasonStockReleased, order.ID, order.UserID, now); err != nil {
				return err
			}
		}
		return tx.Delete(&modelsOrder.Order{}, "id = ?", order.ID).Error
	})
}

// ConfirmPendingOrder atomically confirms an order only when still pending_confirmation with reserved stock.
func (r *OrderRepository) ConfirmPendingOrder(ctx context.Context, id string, confirmedAt time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Where("id = ? AND status = ? AND stock_reserved = ?", id, "pending_confirmation", true).
		Updates(map[string]interface{}{
			"status":       "pending",
			"confirmed_at": confirmedAt,
			"updated_at":   confirmedAt,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ConfirmAndReserveStock atomically confirms a pending_confirmation order and reserves stock.
// Used for AI draft orders that defer stock reservation until customer confirmation.
func (r *OrderRepository) ConfirmAndReserveStock(ctx context.Context, id string, stockDeltas map[string]int, confirmedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Verify order is still pending_confirmation (with row lock to prevent concurrent confirms)
		var order modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND status = ?", id, modelsOrder.OrderStatusPendingConfirm).
			First(&order).Error; err != nil {
			return err
		}

		now := time.Now()
		var all []*modelsOrder.StockTransaction
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			recs, err := reserveStockForOrderLine(tx, warehouseIDFromOrder(&order), productID, qty, modelsOrder.StockReasonStockReserved, id, order.UserID, now)
			if err != nil {
				return err
			}
			all = append(all, recs...)
		}

		// Update order: status → pending, stock_reserved → true
		res := tx.Model(&modelsOrder.Order{}).
			Where("id = ? AND status = ?", id, modelsOrder.OrderStatusPendingConfirm).
			Updates(map[string]interface{}{
				"status":         modelsOrder.OrderStatusPending,
				"stock_reserved": true,
				"confirmed_at":   confirmedAt,
				"updated_at":     confirmedAt,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return writeStockAuditEntries(tx, all)
	})
}

// ReleaseExpiredPendingConfirmationOrders releases stock (if reserved) for expired pending_confirmation drafts and marks them cancelled.
// Each order is processed in its own transaction to prevent one failure from rolling back the entire batch.
func (r *OrderRepository) ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	var orders []modelsOrder.Order
	query := r.db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Where("status = ? AND created_at <= ?", "pending_confirmation", olderThan).
		Order("created_at ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&orders).Error; err != nil {
		return 0, err
	}

	released := 0
	now := time.Now()
	for _, order := range orders {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if order.StockReserved {
				stockDeltas := buildStockDeltasFromItems(order.Items)
				now := time.Now()
				var allAudit []*modelsOrder.StockTransaction
				for productID, qty := range stockDeltas {
					if qty <= 0 {
						continue
					}
					recs, err := releaseStockForOrderLine(tx, warehouseIDFromOrder(&order), productID, qty, modelsOrder.StockReasonDraftExpired, order.ID, "system", now)
					if err != nil {
						return err
					}
					allAudit = append(allAudit, recs...)
				}
				if ae := writeStockAuditEntries(tx, allAudit); ae != nil {
					log.Printf("order: writeStockAuditEntries failed: %v", ae)
				}
			}

			res := tx.Model(&modelsOrder.Order{}).
				Where("id = ? AND status = ?", order.ID, modelsOrder.OrderStatusPendingConfirm).
				Updates(map[string]interface{}{
					"status":         modelsOrder.OrderStatusExpired,
					"stock_reserved": false,
					"updated_at":     now,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				released++
			}
			return nil
		})
		if err != nil {
			// Log the error but continue processing remaining orders
			log.Printf("Warning: failed to release expired draft order %s: %v", order.ID, err)
		}
	}
	return released, nil
}

func buildStockDeltasFromItems(items []modelsOrder.OrderItem) map[string]int {
	stockDeltas := make(map[string]int, len(items))
	for _, item := range items {
		if item.Quantity < 1 {
			continue
		}
		stockDeltas[item.ProductID] += item.Quantity
	}
	return stockDeltas
}

// CountAll returns total order count.
func (r *OrderRepository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&modelsOrder.Order{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountByStatuses returns order count for given statuses.
func (r *OrderRepository) CountByStatuses(ctx context.Context, statuses []string) (int64, error) {
	var total int64
	if len(statuses) == 0 {
		return 0, nil
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Where("status IN ?", statuses).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// SumTotalAmount returns sum(total_amount).
func (r *OrderRepository) SumTotalAmount(ctx context.Context) (float64, error) {
	var result struct {
		Amount float64
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Select("COALESCE(SUM(total_amount), 0) AS amount").
		Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Amount, nil
}

// SumTotalAmountSince returns sum(total_amount) after given time.
func (r *OrderRepository) SumTotalAmountSince(ctx context.Context, since time.Time) (float64, error) {
	var result struct {
		Amount float64
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Where("created_at >= ?", since).
		Select("COALESCE(SUM(total_amount), 0) AS amount").
		Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Amount, nil
}

// FindRecent returns latest orders.
func (r *OrderRepository) FindRecent(ctx context.Context, limit int) ([]modelsOrder.Order, error) {
	var orders []modelsOrder.Order
	if err := r.db.WithContext(ctx).
		Preload("User").
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

// RevenueByMonth returns monthly revenue for the last N months.
func (r *OrderRepository) RevenueByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	type row struct {
		Month      string
		Revenue    float64
		OrderCount int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') AS month, COALESCE(SUM(total_amount), 0) AS revenue, COUNT(*) AS order_count").
		Where("created_at >= NOW() - (? * INTERVAL '1 month') AND status != 'cancelled'", months).
		Group("month").
		Order("month").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		result[i] = map[string]interface{}{
			"month":       row.Month,
			"revenue":     row.Revenue,
			"order_count": row.OrderCount,
		}
	}
	return result, nil
}

// OrderCountByMonth returns monthly order counts for the last N months.
func (r *OrderRepository) OrderCountByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	type row struct {
		Month string
		Count int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) AS count").
		Where("created_at >= NOW() - (? * INTERVAL '1 month') AND status != 'cancelled'", months).
		Group("month").
		Order("month").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		result[i] = map[string]interface{}{
			"month": row.Month,
			"count": row.Count,
		}
	}
	return result, nil
}

// TopProductsByRevenue analyzes order items to find top products by revenue.
func (r *OrderRepository) TopProductsByRevenue(ctx context.Context, limit int) ([]map[string]interface{}, error) {
	var orders []modelsOrder.Order
	if err := r.db.WithContext(ctx).
		Where("status != ?", "cancelled").
		Order("created_at DESC").
		Limit(500).
		Find(&orders).Error; err != nil {
		return nil, err
	}

	productMap := make(map[string]float64)
	for _, order := range orders {
		for _, item := range order.Items {
			productMap[item.ProductID] += float64(item.Quantity) * item.UnitPrice
		}
	}

	type productRevenue struct {
		ProductID string
		Revenue   float64
	}
	var products []productRevenue
	for id, rev := range productMap {
		products = append(products, productRevenue{ProductID: id, Revenue: rev})
	}

	sort.Slice(products, func(i, j int) bool {
		return products[i].Revenue > products[j].Revenue
	})

	if limit > len(products) {
		limit = len(products)
	}
	result := make([]map[string]interface{}, limit)
	for i := 0; i < limit; i++ {
		result[i] = map[string]interface{}{
			"productId": products[i].ProductID,
			"revenue":   products[i].Revenue,
		}
	}
	return result, nil
}

// DistinctOrderingUsers counts unique users who placed orders since a date.
func (r *OrderRepository) DistinctOrderingUsers(ctx context.Context, since time.Time) (int64, error) {
	var result struct {
		Count int64
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Where("created_at >= ?", since).
		Select("COUNT(DISTINCT user_id) AS count").
		Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Count, nil
}

// RevenueByDay returns daily revenue for the last N days.
func (r *OrderRepository) RevenueByDay(ctx context.Context, days int) ([]map[string]interface{}, error) {
	type row struct {
		Date    string
		Revenue float64
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Select("DATE(created_at) AS date, COALESCE(SUM(total_amount), 0) AS revenue").
		Where("created_at >= NOW() - (? * INTERVAL '1 day') AND status != 'cancelled'", days).
		Group("date").
		Order("date").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		result[i] = map[string]interface{}{
			"date":    row.Date,
			"revenue": row.Revenue,
		}
	}
	return result, nil
}

// readStockLevels returns a map of productID → current stock_quantity for the given deltas.
func readStockLevels(tx *gorm.DB, deltas map[string]int) map[string]int {
	ids := make([]string, 0, len(deltas))
	for id := range deltas {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return map[string]int{}
	}
	type row struct {
		ID            string
		StockQuantity int
	}
	var rows []row
	tx.Model(&modelsProduct.Product{}).Select("id, stock_quantity").Where("id IN ?", ids).Find(&rows)
	levels := make(map[string]int, len(rows))
	for _, r := range rows {
		levels[r.ID] = r.StockQuantity
	}
	return levels
}

// writeStockAuditEntries 批量写入库存流水
func writeStockAuditEntries(tx *gorm.DB, records []*modelsOrder.StockTransaction) error {
	if len(records) == 0 {
		return nil
	}
	return tx.Create(records).Error
}

// writeStockAudit writes StockTransaction records for each product delta within the current transaction.
func writeStockAudit(tx *gorm.DB, deltas map[string]int, beforeStock map[string]int, reason, referenceID, operatorID string) error {
	now := time.Now()
	records := make([]*modelsOrder.StockTransaction, 0, len(deltas))
	for productID, change := range deltas {
		if change == 0 {
			continue
		}
		before := beforeStock[productID]
		after := before + (-change) // change is negative for deduction, positive for restoration
		records = append(records, &modelsOrder.StockTransaction{
			ProductID:   productID,
			Change:      -change, // store as actual stock delta (negative = deducted, positive = restored)
			StockBefore: before,
			StockAfter:  after,
			Reason:      reason,
			ReferenceID: referenceID,
			OperatorID:  operatorID,
			CreatedAt:   now,
		})
	}
	return writeStockAuditEntries(tx, records)
}

// warehouseIDFromOrder returns the warehouse ID from an order, or empty string if nil.
func warehouseIDFromOrder(order *modelsOrder.Order) string {
	if order.WarehouseID != nil {
		return *order.WarehouseID
	}
	return ""
}

// reserveStockForOrderLine reserves stock for a single product line.
// When warehouse stock rows exist, uses three-phase reserve (increment Reserved, decrement Product.StockQuantity).
// warehouseID routes to a specific warehouse; empty string falls back to default warehouse.
// Falls back to legacy direct deduction when no warehouse rows exist.
func reserveStockForOrderLine(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	if qty <= 0 {
		return nil, nil
	}
	nWh, err := countWarehouseStockRows(tx, productID)
	if err != nil {
		return nil, err
	}
	if nWh > 0 {
		wid, werr := resolveWarehouseID(tx, warehouseID)
		if werr != nil {
			return nil, werr
		}
		return reserveWarehouseStock(tx, wid, productID, qty, reason, refID, operatorID, t)
	}
	return deductLegacyProductStock(tx, productID, qty, reason, refID, operatorID, t)
}

// releaseStockForOrderLine releases stock for a single product line.
// When warehouse stock rows exist, decrements Reserved and increments Product.StockQuantity.
// warehouseID routes to a specific warehouse; empty string falls back to default warehouse.
// Falls back to legacy direct restoration when no warehouse rows exist.
func releaseStockForOrderLine(tx *gorm.DB, warehouseID, productID string, qty int, reason, refID, operatorID string, t time.Time) ([]*modelsOrder.StockTransaction, error) {
	if qty <= 0 {
		return nil, nil
	}
	nWh, err := countWarehouseStockRows(tx, productID)
	if err != nil {
		return nil, err
	}
	if nWh > 0 {
		wid, werr := resolveWarehouseID(tx, warehouseID)
		if werr != nil {
			return nil, werr
		}
		return releaseWarehouseStock(tx, wid, productID, qty, reason, refID, operatorID, t)
	}
	// Legacy path: restore product stock_quantity directly
	if err := restoreLegacyProductStock(tx, productID, qty); err != nil {
		return nil, err
	}
	var p modelsProduct.Product
	if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
		return nil, err
	}
	rec := &modelsOrder.StockTransaction{
		ProductID:   productID,
		Change:      qty,
		StockBefore: p.StockQuantity - qty,
		StockAfter:  p.StockQuantity,
		Reason:      reason,
		ReferenceID: refID,
		OperatorID:  operatorID,
		CreatedAt:   t,
	}
	return []*modelsOrder.StockTransaction{rec}, nil
}

// ── Analytics ──

// SalesVelocityResult holds per-product sales velocity.
type SalesVelocityResult struct {
	ProductID   string  `json:"productId"`
	ProductName string  `json:"productName"`
	Category    string  `json:"category"`
	TotalQty    int     `json:"totalQty"`
	Revenue     float64 `json:"revenue"`
}

// SalesVelocity returns per-product sales velocity for the last N months.
func (r *OrderRepository) SalesVelocity(ctx context.Context, months int) ([]SalesVelocityResult, error) {
	var rows []SalesVelocityResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.id AS product_id, p.name AS product_name, p.category,
			COALESCE(SUM(oi.quantity), 0)::int AS total_qty,
			COALESCE(SUM(oi.total), 0) AS revenue
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id
			AND o.status NOT IN ('cancelled', 'expired')
			AND o.created_at >= date_trunc('month', NOW()) - (? * INTERVAL '1 month')
		GROUP BY p.id, p.name, p.category
		ORDER BY revenue DESC
	`, months).Scan(&rows).Error
	return rows, err
}

// RFMRecord holds Recency/Frequency/Monetary values for a customer.
type RFMRecord struct {
	UserID      string  `json:"userId"`
	UserName    string  `json:"userName"`
	UserEmail   string  `json:"userEmail"`
	RecencyDays int     `json:"recencyDays"`
	Frequency   int     `json:"frequency"`
	Monetary    float64 `json:"monetary"`
	Segment     string  `json:"segment"`
}

// RFMAnalysis returns RFM values for all ordering users.
func (r *OrderRepository) RFMAnalysis(ctx context.Context) ([]RFMRecord, error) {
	var rows []RFMRecord
	err := r.db.WithContext(ctx).Raw(`
		SELECT u.id AS user_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.email) AS user_name,
			u.email AS user_email,
			EXTRACT(DAY FROM NOW() - MAX(o.created_at))::int AS recency_days,
			COUNT(DISTINCT o.id)::int AS frequency,
			COALESCE(SUM(o.total_amount), 0) AS monetary
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		WHERE o.status NOT IN ('cancelled', 'expired')
		GROUP BY u.id, u.first_name, u.last_name, u.email
		ORDER BY monetary DESC
	`).Scan(&rows).Error
	return rows, err
}

// CustomerChurnResult holds churn risk for a customer.
type CustomerChurnResult struct {
	UserID    string `json:"userId"`
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
	LastOrder string `json:"lastOrder"`
	DaysSince int    `json:"daysSince"`
	RiskLevel string `json:"riskLevel"`
}

// CustomerChurn finds customers at risk of churning (no order in N days).
func (r *OrderRepository) CustomerChurn(ctx context.Context, dormantDays int) ([]CustomerChurnResult, error) {
	var rows []CustomerChurnResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT u.id AS user_id,
			COALESCE(u.first_name || ' ' || u.last_name, u.email) AS user_name,
			u.email AS user_email,
			MAX(o.created_at)::text AS last_order,
			EXTRACT(DAY FROM NOW() - MAX(o.created_at))::int AS days_since
		FROM users u
		INNER JOIN orders o ON u.id = o.user_id
		WHERE o.status NOT IN ('cancelled', 'expired')
		GROUP BY u.id, u.first_name, u.last_name, u.email
		HAVING MAX(o.created_at) < NOW() - (? * INTERVAL '1 day')
		ORDER BY MAX(o.created_at)
	`, dormantDays).Scan(&rows).Error
	return rows, err
}

// InventoryHealthResult holds inventory health metrics for a product.
type InventoryHealthResult struct {
	ProductID      string  `json:"productId"`
	ProductName    string  `json:"productName"`
	StockQuantity  int     `json:"stockQuantity"`
	SafetyStock    int     `json:"safetyStock"`
	AvgDailySales  float64 `json:"avgDailySales"`
	SellableDays   float64 `json:"sellableDays"`
	ShortageAlert  bool    `json:"shortageAlert"`
	OverstockAlert bool    `json:"overstockAlert"`
}

// InventoryHealth computes sellable days and shortage alerts for all products.
func (r *OrderRepository) InventoryHealth(ctx context.Context, salesWindowDays int) ([]InventoryHealthResult, error) {
	var rows []InventoryHealthResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.id AS product_id,
			p.name AS product_name,
			p.stock_quantity AS stock_quantity,
			COALESCE(p.safety_stock, 0) AS safety_stock,
			ROUND(COALESCE(SUM(oi.quantity), 0)::numeric / NULLIF(?, 0), 2) AS avg_daily_sales,
			CASE WHEN COALESCE(SUM(oi.quantity), 0) > 0
				THEN ROUND((p.stock_quantity::numeric / (SUM(oi.quantity)::numeric / NULLIF(?, 0))), 1)
				ELSE 999
			END AS sellable_days
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id
			AND o.status NOT IN ('cancelled', 'expired')
			AND o.created_at >= NOW() - (? * INTERVAL '1 day')
		GROUP BY p.id, p.name, p.stock_quantity, p.safety_stock
		ORDER BY sellable_days
	`, salesWindowDays, salesWindowDays, salesWindowDays).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		rows[i].ShortageAlert = rows[i].SellableDays < float64(rows[i].SafetyStock)
		rows[i].OverstockAlert = rows[i].SellableDays > 180
	}
	return rows, nil
}

// ProfitLossResult holds a P&L row.
type ProfitLossResult struct {
	Period         string  `json:"period"`
	Revenue        float64 `json:"revenue"`
	COGS           float64 `json:"cogs"`
	ShippingCost   float64 `json:"shippingCost"`
	GrossProfit     float64 `json:"grossProfit"`
	GrossMarginPct  float64 `json:"grossMarginPct"`
	OrderCount     int     `json:"orderCount"`
}

// ProfitLossByPeriod returns P&L grouped by period (day/week/month).
func (r *OrderRepository) ProfitLossByPeriod(ctx context.Context, groupBy string, periods int) ([]ProfitLossResult, error) {
	var truncate string
	var interval string
	switch groupBy {
	case "week":
		truncate = "week"
		interval = "week"
	case "day":
		truncate = "day"
		interval = "day"
	default:
		truncate = "month"
		interval = "month"
	}

	var rows []ProfitLossResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT date_trunc(?, o.created_at)::text AS period,
			COALESCE(SUM(o.total_amount), 0) AS revenue,
			COALESCE(SUM(o.cogs), 0) AS cogs,
			COALESCE(SUM(o.shipping_amount), 0) AS shipping_cost,
			COALESCE(SUM(o.total_amount) - SUM(o.cogs) - SUM(o.shipping_amount), 0) AS gross_profit,
			CASE WHEN SUM(o.total_amount) > 0
				THEN ROUND(((SUM(o.total_amount) - SUM(o.cogs) - SUM(o.shipping_amount)) / SUM(o.total_amount) * 100)::numeric, 1)
				ELSE 0
			END AS gross_margin_pct,
			COUNT(o.id)::int AS order_count
		FROM orders o
		WHERE o.status NOT IN ('cancelled', 'expired')
			AND o.created_at >= date_trunc(?, NOW()) - (? * INTERVAL '1 ' || ?)
		GROUP BY date_trunc(?, o.created_at)
		ORDER BY period
	`, truncate, truncate, periods, interval, truncate).Scan(&rows).Error
	return rows, err
}

// ReplenishmentItem holds a single replenishment suggestion.
type ReplenishmentItem struct {
	ProductID       string  `json:"productId"`
	ProductName     string  `json:"productName"`
	CurrentStock    int     `json:"currentStock"`
	SafetyStock     int     `json:"safetyStock"`
	AvgDailySales   float64 `json:"avgDailySales"`
	InTransit       int     `json:"inTransit"`
	PendingOrders   int     `json:"pendingOrders"`
	SuggestedQty    int     `json:"suggestedQty"`
	CycleDays       int     `json:"cycleDays"`
}

// ReplenishmentSuggestions computes smart replenishment for all products.
func (r *OrderRepository) ReplenishmentSuggestions(ctx context.Context, cycleDays int, salesWindowDays int) ([]ReplenishmentItem, error) {
	var rows []ReplenishmentItem
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.id AS product_id,
			p.name AS product_name,
			p.stock_quantity AS current_stock,
			COALESCE(p.safety_stock, 0) AS safety_stock,
			ROUND(COALESCE(SUM(oi.quantity), 0)::numeric / NULLIF(?, 0), 2) AS avg_daily_sales,
			0 AS in_transit,
			0 AS pending_orders,
			GREATEST(0, (
				(ROUND(COALESCE(SUM(oi.quantity), 0)::numeric / NULLIF(?, 0), 2) * ?)
				+ COALESCE(p.safety_stock, 0)
				- p.stock_quantity
			)::int) AS suggested_qty,
			? AS cycle_days
		FROM products p
		LEFT JOIN order_items oi ON p.id = oi.product_id
		LEFT JOIN orders o ON oi.order_id = o.id
			AND o.status NOT IN ('cancelled', 'expired')
			AND o.created_at >= NOW() - (? * INTERVAL '1 day')
		GROUP BY p.id, p.name, p.stock_quantity, p.safety_stock
		HAVING (ROUND(COALESCE(SUM(oi.quantity), 0)::numeric / NULLIF(?, 0), 2) * ?)
			+ COALESCE(p.safety_stock, 0)
			- p.stock_quantity > 0
		ORDER BY suggested_qty DESC
	`, salesWindowDays, salesWindowDays, cycleDays, cycleDays, salesWindowDays, salesWindowDays, cycleDays).Scan(&rows).Error
	return rows, err
}
