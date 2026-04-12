package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
)

import (
	"context"
	"fmt"
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
		for productID, qty := range stockDeltas {
			if qty <= 0 {
				continue
			}
			res := tx.Model(&modelsProduct.Product{}).
				Where("id = ? AND stock_quantity >= ?", productID, qty).
				Update("stock_quantity", gorm.Expr("stock_quantity - ?", qty))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return fmt.Errorf("insufficient stock for product %s during reservation", productID)
			}
		}
		return tx.Create(order).Error
	})
}

// Update updates an order
func (r *OrderRepository) Update(ctx context.Context, order *modelsOrder.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// UpdateWithStockAdjustment updates an order and adjusts stock atomically.
// Positive delta reserves additional stock; negative delta releases stock.
func (r *OrderRepository) UpdateWithStockAdjustment(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for productID, delta := range stockDeltas {
			switch {
			case delta > 0:
				res := tx.Model(&modelsProduct.Product{}).
					Where("id = ? AND stock_quantity >= ?", productID, delta).
					Update("stock_quantity", gorm.Expr("stock_quantity - ?", delta))
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return fmt.Errorf("insufficient stock for product %s during update adjustment", productID)
				}
			case delta < 0:
				releaseQty := -delta
				if err := tx.Model(&modelsProduct.Product{}).
					Where("id = ?", productID).
					Update("stock_quantity", gorm.Expr("stock_quantity + ?", releaseQty)).Error; err != nil {
					return err
				}
			}
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
		order.StockReserved = false
		return tx.Save(order).Error
	})
}

// DeleteWithStockRestore restores reserved stock and deletes the order in one transaction.
func (r *OrderRepository) DeleteWithStockRestore(ctx context.Context, order *modelsOrder.Order, stockDeltas map[string]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

// ReleaseExpiredPendingConfirmationOrders releases stock for expired pending_confirmation drafts and marks them cancelled.
func (r *OrderRepository) ReleaseExpiredPendingConfirmationOrders(ctx context.Context, olderThan time.Time, limit int) (int, error) {
	released := 0
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var orders []modelsOrder.Order
		query := tx.
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ? AND stock_reserved = ? AND created_at <= ?", "pending_confirmation", true, olderThan).
			Order("created_at ASC")
		if limit > 0 {
			query = query.Limit(limit)
		}
		if err := query.Find(&orders).Error; err != nil {
			return err
		}

		now := time.Now()
		for _, order := range orders {
			stockDeltas := buildStockDeltasFromItems(order.Items)
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

			res := tx.Model(&modelsOrder.Order{}).
				Where("id = ? AND status = ? AND stock_reserved = ?", order.ID, "pending_confirmation", true).
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
