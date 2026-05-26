package database

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/catalog"
	"log"

	"gorm.io/gorm"
)

// BackfillOrderCOGS recalculates inflated COGS (cogs > subtotal) for confirmed+ orders.
func BackfillOrderCOGS(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var orders []modelsOrder.Order
	statuses := []string{
		modelsOrder.OrderStatusConfirmed,
		"production",
		modelsOrder.OrderStatusShipped,
		modelsOrder.OrderStatusDelivered,
	}
	if err := db.Where("status IN ? AND subtotal > 0 AND cogs > subtotal", statuses).Find(&orders).Error; err != nil {
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
		if newCOGS <= 0 || newCOGS >= order.COGS {
			continue
		}
		if err := db.Model(&modelsOrder.Order{}).Where("id = ?", order.ID).Update("cogs", newCOGS).Error; err != nil {
			return err
		}
		fixed++
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
				continue
			}
			p = productCOGSCache{slug: slug, basePrice: basePrice, weighted: weighted}
			cache[productID] = p
		}
		ref := catalog.COGSReferencePrice(p.slug, p.basePrice, p.weighted)
		if p.weighted <= 0 || item.UnitPrice <= 0 || ref <= 0 {
			continue
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
