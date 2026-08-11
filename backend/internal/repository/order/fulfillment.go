package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FulfillmentRepository handles fulfillment data operations.
type FulfillmentRepository struct {
	db *gorm.DB
}

// NewFulfillmentRepository creates a new FulfillmentRepository.
func NewFulfillmentRepository(db *gorm.DB) *FulfillmentRepository {
	return &FulfillmentRepository{db: db}
}

// Create creates a fulfillment, deducts warehouse stock, and updates order item fulfilled quantities.
func (r *FulfillmentRepository) Create(ctx context.Context, fulfillment *modelsOrder.Fulfillment, items []modelsOrder.FulfillmentItem, warehouseID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Lock the order row
		var order modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", fulfillment.OrderID).First(&order).Error; err != nil {
			return err
		}

		now := time.Now()
		fulfillment.CreatedAt = now
		fulfillment.UpdatedAt = now
		if err := tx.Create(fulfillment).Error; err != nil {
			return err
		}

		// G21-d: an order whose stock was already issued through the trade dispatch
		// path (DispatchShipment writes StockReasonDispatched audit rows for the
		// whole order) must not have the same stock deducted again by a
		// fulfillment. Skip the physical deduction in that case — the goods already
		// left the warehouse — but still book the fulfilled quantities.
		var dispatched int64
		if err := tx.Model(&modelsOrder.StockTransaction{}).
			Where("reference_id = ? AND reason = ?", order.ID, modelsOrder.StockReasonDispatched).
			Count(&dispatched).Error; err != nil {
			return err
		}
		alreadyDispatched := dispatched > 0

		var allAudit []*modelsOrder.StockTransaction
		for i := range items {
			items[i].FulfillmentID = fulfillment.ID

			// Validate item index
			if items[i].OrderItemIdx < 0 || items[i].OrderItemIdx >= len(order.Items) {
				return fmt.Errorf("invalid order item index %d", items[i].OrderItemIdx)
			}
			oi := &order.Items[items[i].OrderItemIdx]
			if items[i].ProductID != oi.ProductID {
				return fmt.Errorf("product mismatch at index %d", items[i].OrderItemIdx)
			}
			remaining := oi.Quantity - oi.FulfilledQuantity
			if items[i].Quantity > remaining {
				return fmt.Errorf("fulfillment qty %d exceeds remaining %d for product %s",
					items[i].Quantity, remaining, oi.ProductID)
			}

			if !alreadyDispatched {
				// Deduct warehouse stock (actual goods leaving warehouse). The
				// reserved flag tells deductWarehouseStock whether
				// Product.StockQuantity was already decremented at reservation time
				// (G21-c).
				wid, err := resolveWarehouseID(tx, warehouseID)
				if err != nil {
					return err
				}
				recs, err := deductWarehouseStock(tx, wid, items[i].ProductID, items[i].Quantity,
					order.StockReserved, modelsOrder.StockReasonGoodsIssued, fulfillment.ID, "system", now)
				if err != nil {
					return err
				}
				allAudit = append(allAudit, recs...)
			}

			// Update order item fulfilled quantity
			oi.FulfilledQuantity += items[i].Quantity
		}

		// Persist updated order items
		itemsJSON, err := json.Marshal(order.Items)
		if err != nil {
			return err
		}
		if err := tx.Model(&order).Update("items", itemsJSON).Error; err != nil {
			return err
		}

		// Update order status based on fulfillment.
		//
		// M4: enforce the order state machine before shipping. The legal flow is
		// confirmed -> production -> shipped / partially_shipped, so a confirmed
		// order is advanced to production first, and any state that cannot legally
		// ship (pending, pending_approval, cancelled, delivered, ...) is rejected
		// here. This runs inside the same transaction as the fulfillment create and
		// the order row lock acquired above, keeping the transition atomic.
		switch order.Status {
		case modelsOrder.OrderStatusConfirmed:
			// Advance through production before shipping.
			if err := tx.Model(&order).Updates(map[string]interface{}{
				"status":     modelsOrder.OrderStatusProduction,
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
			order.Status = modelsOrder.OrderStatusProduction
		case modelsOrder.OrderStatusProduction, modelsOrder.OrderStatusPartiallyShipped:
			// Already in a ship-eligible state.
		default:
			return fmt.Errorf("order %s cannot be shipped from status %s", order.ID, order.Status)
		}

		allFulfilled := true
		anyFulfilled := false
		for _, oi := range order.Items {
			if oi.FulfilledQuantity > 0 {
				anyFulfilled = true
			}
			if oi.FulfilledQuantity < oi.Quantity {
				allFulfilled = false
			}
		}
		newStatus := order.Status
		if allFulfilled {
			newStatus = modelsOrder.OrderStatusShipped
		} else if anyFulfilled && order.Status == modelsOrder.OrderStatusProduction {
			newStatus = modelsOrder.OrderStatusPartiallyShipped
		} else if anyFulfilled && order.Status == modelsOrder.OrderStatusPartiallyShipped {
			newStatus = modelsOrder.OrderStatusPartiallyShipped
		}
		if newStatus != order.Status {
			// Belt-and-suspenders: the transition matrix must permit this move.
			if err := modelsOrder.ValidateOrderStatusTransition(order.Status, newStatus); err != nil {
				return err
			}
			if err := tx.Model(&order).Updates(map[string]interface{}{
				"status":     newStatus,
				"updated_at": now,
			}).Error; err != nil {
				return err
			}
		}

		return writeStockAuditEntries(tx, allAudit)
	})
}

// FulfillmentListRow 履约列表行（含订单号）
type FulfillmentListRow struct {
	modelsOrder.Fulfillment
	OrderNumber string `json:"orderNumber"`
}

// ListPaginated 分页查询履约记录并关联订单号
func (r *FulfillmentRepository) ListPaginated(ctx context.Context, page, limit int, status string) ([]FulfillmentListRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	base := r.db.WithContext(ctx).
		Table("fulfillments AS f").
		Joins("LEFT JOIN orders AS o ON o.id = f.order_id")
	if status != "" {
		base = base.Where("f.status = ?", status)
	}
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []FulfillmentListRow
	offset := (page - 1) * limit
	err := base.
		Select("f.*, o.order_number AS order_number").
		Order("f.created_at DESC").
		Offset(offset).Limit(limit).
		Scan(&rows).Error
	return rows, total, err
}

// FindByOrder returns all fulfillments for an order.
func (r *FulfillmentRepository) FindByOrder(ctx context.Context, orderID string) ([]modelsOrder.Fulfillment, error) {
	var fulfillments []modelsOrder.Fulfillment
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).
		Order("created_at DESC").Find(&fulfillments).Error; err != nil {
		return nil, err
	}
	return fulfillments, nil
}

// FindByID returns a fulfillment with its items.
func (r *FulfillmentRepository) FindByID(ctx context.Context, id string) (*modelsOrder.Fulfillment, []modelsOrder.FulfillmentItem, error) {
	var f modelsOrder.Fulfillment
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&f).Error; err != nil {
		return nil, nil, err
	}
	var items []modelsOrder.FulfillmentItem
	if err := r.db.WithContext(ctx).Where("fulfillment_id = ?", id).Find(&items).Error; err != nil {
		return nil, nil, err
	}
	return &f, items, nil
}

// Ship marks a fulfillment as shipped with tracking info.
//
// G21-e: the order row is read with FOR UPDATE and the fulfillment status
// transition is a conditional UPDATE so concurrent ships cannot lose
// ShippedQuantity updates. Only one caller can win the status transition for a
// given fulfillment, and concurrent ships of different fulfillments of the same
// order serialize on the order row lock before reading order.Items.
func (r *FulfillmentRepository) Ship(ctx context.Context, id, trackingNumber, carrier string, shippedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var f modelsOrder.Fulfillment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).First(&f).Error; err != nil {
			return err
		}
		if f.Status != modelsOrder.FulfillmentStatusPacked && f.Status != modelsOrder.FulfillmentStatusPending {
			return fmt.Errorf("fulfillment %s cannot be shipped from status %s", id, f.Status)
		}
		updates := map[string]interface{}{
			"status":          modelsOrder.FulfillmentStatusShipped,
			"tracking_number": trackingNumber,
			"carrier":         carrier,
			"shipped_at":      shippedAt,
			"updated_at":      shippedAt,
		}
		res := tx.Model(&modelsOrder.Fulfillment{}).
			Where("id = ? AND status IN ?", id, []string{modelsOrder.FulfillmentStatusPacked, modelsOrder.FulfillmentStatusPending}).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("fulfillment %s cannot be shipped (concurrent ship)", id)
		}
		// Update order ShippedQuantity for each item
		var items []modelsOrder.FulfillmentItem
		if err := tx.Where("fulfillment_id = ?", id).Find(&items).Error; err != nil {
			return err
		}
		var order modelsOrder.Order
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", f.OrderID).First(&order).Error; err != nil {
			return err
		}
		for _, item := range items {
			if item.OrderItemIdx >= 0 && item.OrderItemIdx < len(order.Items) {
				order.Items[item.OrderItemIdx].ShippedQuantity += item.Quantity
			}
		}
		itemsJSON, _ := json.Marshal(order.Items)
		return tx.Model(&order).Update("items", itemsJSON).Error
	})
}

// Deliver marks a fulfillment as delivered.
func (r *FulfillmentRepository) Deliver(ctx context.Context, id string, deliveredAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var f modelsOrder.Fulfillment
		if err := tx.Where("id = ?", id).First(&f).Error; err != nil {
			return err
		}
		if f.Status != modelsOrder.FulfillmentStatusShipped {
			return fmt.Errorf("fulfillment %s cannot be delivered from status %s", id, f.Status)
		}
		updates := map[string]interface{}{
			"status":       modelsOrder.FulfillmentStatusDelivered,
			"delivered_at": deliveredAt,
			"updated_at":   deliveredAt,
		}
		if err := tx.Model(&f).Updates(updates).Error; err != nil {
			return err
		}
		// Check if all fulfillments for this order are delivered
		var allFulfillments []modelsOrder.Fulfillment
		if err := tx.Where("order_id = ?", f.OrderID).Find(&allFulfillments).Error; err != nil {
			return err
		}
		allDelivered := true
		for _, ff := range allFulfillments {
			if ff.ID != id && ff.Status != modelsOrder.FulfillmentStatusDelivered {
				allDelivered = false
				break
			}
		}
		if allDelivered {
			return tx.Model(&modelsOrder.Order{}).Where("id = ?", f.OrderID).
				Updates(map[string]interface{}{
					"status":       modelsOrder.OrderStatusDelivered,
					"delivered_at": deliveredAt,
					"updated_at":   deliveredAt,
				}).Error
		}
		// Partial delivery
		return tx.Model(&modelsOrder.Order{}).Where("id = ?", f.OrderID).
			Updates(map[string]interface{}{
				"status":     modelsOrder.OrderStatusPartiallyDelivered,
				"updated_at": deliveredAt,
			}).Error
	})
}

// Cancel cancels a fulfillment and restores warehouse stock.
func (r *FulfillmentRepository) Cancel(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var f modelsOrder.Fulfillment
		if err := tx.Where("id = ?", id).First(&f).Error; err != nil {
			return err
		}
		if f.Status == modelsOrder.FulfillmentStatusDelivered {
			return fmt.Errorf("cannot cancel delivered fulfillment %s", id)
		}
		// Restore warehouse stock
		var items []modelsOrder.FulfillmentItem
		if err := tx.Where("fulfillment_id = ?", id).Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			if err := addStockToWarehouse(tx, f.WarehouseID, item.ProductID, item.Quantity); err != nil {
				return err
			}
		}
		// Restore order item fulfilled quantities
		var order modelsOrder.Order
		if err := tx.Where("id = ?", f.OrderID).First(&order).Error; err != nil {
			return err
		}
		for _, item := range items {
			if item.OrderItemIdx >= 0 && item.OrderItemIdx < len(order.Items) {
				order.Items[item.OrderItemIdx].FulfilledQuantity -= item.Quantity
				if order.Items[item.OrderItemIdx].FulfilledQuantity < 0 {
					order.Items[item.OrderItemIdx].FulfilledQuantity = 0
				}
			}
		}
		itemsJSON, _ := json.Marshal(order.Items)
		if err := tx.Model(&order).Update("items", itemsJSON).Error; err != nil {
			return err
		}
		return tx.Model(&f).Updates(map[string]interface{}{
			"status":     modelsOrder.FulfillmentStatusCancelled,
			"updated_at": time.Now(),
		}).Error
	})
}
