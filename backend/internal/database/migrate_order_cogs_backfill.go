package database

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/catalog"
	"log"

	"gorm.io/gorm"
)

// BackfillOrderCOGS repairs missing or inflated COGS for committed orders using
// available product costs. Orders without usable cost data remain unchanged.
func BackfillOrderCOGS(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var orders []modelsOrder.Order
	statuses := []string{
		modelsOrder.OrderStatusConfirmed,
		modelsOrder.OrderStatusProduction,
		modelsOrder.OrderStatusPartiallyShipped,
		modelsOrder.OrderStatusShipped,
		modelsOrder.OrderStatusPartiallyDelivered,
		modelsOrder.OrderStatusDelivered,
		modelsOrder.OrderStatusPartiallyReturned,
		modelsOrder.OrderStatusReturned,
	}
	if err := db.Where("(status IN ? OR (status = ? AND confirmed_at IS NOT NULL)) AND subtotal > 0 AND (COALESCE(cogs, 0) <= 0 OR cogs > subtotal)",
		statuses, modelsOrder.OrderStatusPending).Find(&orders).Error; err != nil {
		return err
	}
	if len(orders) == 0 {
		return nil
	}

	productCache := make(map[string]productCOGSCache)
	var fixed int64
	for i := range orders {
		order := &orders[i]
		newCOGS := backfillComputeOrderCOGS(db, productCache, order.Items)
		if newCOGS <= 0 || (order.COGS > 0 && newCOGS >= order.COGS) {
			continue
		}
		// Avoid overwriting a concurrent financial correction and invalidate stale
		// editor snapshots whenever a historical cost is repaired.
		result := db.Model(&modelsOrder.Order{}).
			Where("id = ? AND version = ? AND (COALESCE(cogs, 0) <= 0 OR cogs > subtotal)", order.ID, order.Version).
			Updates(map[string]interface{}{"cogs": newCOGS, "version": gorm.Expr("version + 1")})
		if result.Error != nil {
			return result.Error
		}
		fixed += result.RowsAffected
	}
	if fixed > 0 {
		log.Printf("Backfilled COGS for %d order(s)", fixed)
	}
	return nil
}

func backfillComputeOrderCOGS(db *gorm.DB, cache map[string]productCOGSCache, items modelsOrder.OrderItemArray) float64 {
	var total float64
	for _, item := range items {
		if item.Quantity < 1 {
			continue
		}
		productID := item.ProductID
		p, ok := cache[productID]
		if !ok {
			var slug string
			var basePrice, weighted float64
			if err := db.Raw(`SELECT slug, base_price, weighted_avg_cost FROM products WHERE id = ?`, productID).
				Row().Scan(&slug, &basePrice, &weighted); err != nil {
				// A partial positive total would stop this order qualifying for the
				// missing-cost repair after its remaining product data is corrected.
				return 0
			}
			p = productCOGSCache{slug: slug, basePrice: basePrice, weighted: weighted}
			cache[productID] = p
		}
		ref := catalog.COGSReferencePrice(p.slug, p.basePrice, p.weighted)
		if p.weighted <= 0 || item.UnitPrice <= 0 || ref <= 0 {
			return 0
		}
		total += float64(item.Quantity) * item.UnitPrice * (p.weighted / ref)
	}
	return total
}

type productCOGSCache struct {
	slug      string
	basePrice float64
	weighted  float64
}
