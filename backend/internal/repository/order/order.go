package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
)

import (
	"context"
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

// FindAll returns paginated orders
func (r *OrderRepository) FindAll(ctx context.Context, page, limit int) ([]modelsOrder.Order, int64, error) {
	var orders []modelsOrder.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsOrder.Order{})

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

// CreateWithStockReservation creates an order and atomically deducts product stock.
func (r *OrderRepository) CreateWithStockReservation(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var all []*modelsOrder.StockTransaction
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			recs, err := deductStockForProductLine(tx, productID, qty, modelsOrder.StockReasonOrderCreated, order.ID, order.UserID, now)
			if err != nil {
				return err
			}
			all = append(all, recs...)
		}
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

// UpdateWithStockAdjustment updates an order and adjusts stock atomically.
// Positive delta reserves additional stock; negative delta releases stock.
func (r *OrderRepository) UpdateWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var extra []*modelsOrder.StockTransaction
		for productID, delta := range stockDeltas {
			switch {
			case delta > 0:
				recs, err := deductStockForProductLine(tx, productID, delta, modelsOrder.StockReasonOrderUpdated, order.ID, order.UserID, now)
				if err != nil {
					return err
				}
				extra = append(extra, recs...)
			case delta < 0:
				releaseQty := -delta
				var p modelsProduct.Product
				if err := tx.Where("id = ?", productID).First(&p).Error; err != nil {
					return err
				}
				before := p.StockQuantity
				if err := restoreLegacyProductStock(tx, productID, releaseQty); err != nil {
					return err
				}
				after := before + releaseQty
				extra = append(extra, &modelsOrder.StockTransaction{
					ProductID:   productID,
					Change:      releaseQty,
					StockBefore: before,
					StockAfter:  after,
					Reason:      modelsOrder.StockReasonOrderUpdated,
					ReferenceID: order.ID,
					OperatorID:  order.UserID,
					CreatedAt:   now,
				})
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

// ReleaseStockForOrder returns reserved stock and marks the order as not reserved.
func (r *OrderRepository) ReleaseStockForOrder(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		beforeStock := readStockLevels(tx, stockDeltas)
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			if err := restoreLegacyProductStock(tx, productID, qty); err != nil {
				return err
			}
		}
		order.StockReserved = false
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		// Negate deltas for audit: releasing stock means positive change
		negDeltas := make(map[string]int, len(stockDeltas))
		for pid, qty := range stockDeltas {
			negDeltas[pid] = -qty
		}
		return writeStockAudit(tx, negDeltas, beforeStock, modelsOrder.StockReasonOrderCancelled, order.ID, order.UserID)
	})
}

// DeleteWithStockRestore restores reserved stock and deletes the order in one transaction.
func (r *OrderRepository) DeleteWithStockRestore(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		beforeStock := readStockLevels(tx, stockDeltas)
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			if err := restoreLegacyProductStock(tx, productID, qty); err != nil {
				return err
			}
		}
		if err := tx.Delete(&modelsOrder.Order{}, "id = ?", order.ID).Error; err != nil {
			return err
		}
		negDeltas := make(map[string]int, len(stockDeltas))
		for pid, qty := range stockDeltas {
			negDeltas[pid] = -qty
		}
		return writeStockAudit(tx, negDeltas, beforeStock, modelsOrder.StockReasonOrderDeleted, order.ID, order.UserID)
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
		// Verify order is still pending_confirmation
		var order modelsOrder.Order
		if err := tx.Where("id = ? AND status = ?", id, "pending_confirmation").First(&order).Error; err != nil {
			return err
		}

		now := time.Now()
		var all []*modelsOrder.StockTransaction
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			recs, err := deductStockForProductLine(tx, productID, qty, modelsOrder.StockReasonOrderConfirmed, id, order.UserID, now)
			if err != nil {
				return err
			}
			all = append(all, recs...)
		}

		// Update order: status → pending, stock_reserved → true
		res := tx.Model(&modelsOrder.Order{}).
			Where("id = ? AND status = ?", id, "pending_confirmation").
			Updates(map[string]interface{}{
				"status":         "pending",
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
func (r *OrderRepository) ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	released := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var orders []modelsOrder.Order
		query := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND created_at <= ?", "pending_confirmation", olderThan).
			Order("created_at ASC")
		if limit > 0 {
			query = query.Limit(limit)
		}
		if err := query.Find(&orders).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, order := range orders {
			// Only release stock if it was actually reserved (legacy orders)
			if order.StockReserved {
				stockDeltas := buildStockDeltasFromItems(order.Items)
				beforeStock := readStockLevels(tx, stockDeltas)
				for productID, qty := range stockDeltas {
					if qty <= 0 {
						continue
					}
					if err := tx.Model(&modelsProduct.Product{}).
						Where("id = ?", productID).
						Update("stock_quantity", gorm.Expr("stock_quantity + ?", qty)).Error; err != nil {
						return err
					}
				}
				negDeltas := make(map[string]int, len(stockDeltas))
				for pid, qty := range stockDeltas {
					negDeltas[pid] = -qty
				}
				_ = writeStockAudit(tx, negDeltas, beforeStock, modelsOrder.StockReasonDraftExpired, order.ID, "system")
			}

			res := tx.Model(&modelsOrder.Order{}).
				Where("id = ? AND status = ?", order.ID, "pending_confirmation").
				Updates(map[string]interface{}{
					"status":         "cancelled",
					"stock_reserved": false,
					"updated_at":     now,
				})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected > 0 {
				released++
			}
		}
		return nil
	})
	return released, err
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
		Month   string
		Revenue float64
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Order{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') AS month, COALESCE(SUM(total_amount), 0) AS revenue").
		Where("created_at >= NOW() - INTERVAL '? months' AND status != 'cancelled'", months).
		Group("month").
		Order("month").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		result[i] = map[string]interface{}{
			"month":   row.Month,
			"revenue": row.Revenue,
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
		Where("created_at >= NOW() - INTERVAL '? months' AND status != 'cancelled'", months).
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
		Where("created_at >= NOW() - INTERVAL '? days' AND status != 'cancelled'", days).
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
