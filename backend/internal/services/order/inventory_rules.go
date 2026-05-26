package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"fmt"
	"strings"
)

// InventoryWarning 结构化库存警告，供 handler 层 i18n 翻译
type InventoryWarning struct {
	Code      string `json:"code"`
	ProductID string `json:"productId"`
	Remaining int    `json:"remaining,omitempty"`
	Available int    `json:"available,omitempty"`
	Ordered   int    `json:"ordered,omitempty"`
	MOQ       int    `json:"moq,omitempty"`
}

const (
	InventoryWarningSoldOut           = "inventory_warning_sold_out"
	InventoryWarningLowStock          = "inventory_warning_low_stock"
	InventoryWarningSoldOutAdmin      = "inventory_warning_sold_out_admin"
	InventoryWarningBelowMOQAfterOrder = "inventory_warning_below_moq_after_order"
)

// InventoryValidationResult contains validation output for inventory checks.
type InventoryValidationResult struct {
	Warnings   []InventoryWarning `json:"warnings"`
	Violations []string           `json:"violations"`
}

// ValidateInventory validates quantities against current product inventory and MOQ.
func (s *OrderService) ValidateInventory(items []modelsOrder.OrderItem, productsByID map[string]modelsProduct.Product) InventoryValidationResult {
	return validateInventoryWithSellable(items, productsByID, nil)
}

// ValidateInventoryWithSellable 使用每 SKU 有效可售量（如 OMS 渠道封顶后）；sellable 为 nil 时回退 Product.StockQuantity
func (s *OrderService) ValidateInventoryWithSellable(items []modelsOrder.OrderItem, productsByID map[string]modelsProduct.Product, sellable map[string]int) InventoryValidationResult {
	return validateInventoryWithSellable(items, productsByID, sellable)
}

func validateInventoryWithSellable(items []modelsOrder.OrderItem, productsByID map[string]modelsProduct.Product, sellable map[string]int) InventoryValidationResult {
	result := InventoryValidationResult{
		Warnings:   make([]InventoryWarning, 0, 8),
		Violations: make([]string, 0, 8),
	}
	if len(items) == 0 {
		result.Violations = append(result.Violations, "No order items were provided for inventory validation.")
		return result
	}

	aggregatedQty := make(map[string]int, len(items))
	for _, item := range items {
		productID := strings.TrimSpace(item.ProductID)
		if productID == "" {
			result.Violations = append(result.Violations, "Order item has empty productId.")
			continue
		}
		if item.Quantity < 1 {
			result.Violations = append(result.Violations, fmt.Sprintf("Product %s has invalid quantity %d.", productID, item.Quantity))
			continue
		}
		aggregatedQty[productID] += item.Quantity
	}

	for productID, qty := range aggregatedQty {
		product, ok := productsByID[productID]
		if !ok {
			result.Violations = append(result.Violations, fmt.Sprintf("Product %s is missing from inventory context.", productID))
			continue
		}

		if product.MOQ > 0 && qty < product.MOQ {
			result.Violations = append(result.Violations, fmt.Sprintf("Product %s requested quantity %d is below MOQ %d.", productID, qty, product.MOQ))
			continue
		}

		stock := product.StockQuantity
		if sellable != nil {
			if sv, ok := sellable[productID]; ok {
				stock = sv
			}
		}
		if stock <= 0 {
			result.Violations = append(result.Violations, fmt.Sprintf("Product %s is out of stock.", productID))
			continue
		}
		if qty > stock {
			result.Violations = append(result.Violations, fmt.Sprintf("Product %s requested quantity %d exceeds stock %d.", productID, qty, stock))
			continue
		}

		remaining := stock - qty
		switch {
		case remaining == 0:
			result.Warnings = append(result.Warnings, InventoryWarning{
				Code: InventoryWarningSoldOut, ProductID: productID,
			})
		case product.MOQ > 0 && remaining > 0 && remaining < product.MOQ:
			result.Warnings = append(result.Warnings, InventoryWarning{
				Code: InventoryWarningBelowMOQAfterOrder, ProductID: productID,
				Remaining: remaining, MOQ: product.MOQ,
			})
		case remaining <= lowStockThreshold(product.MOQ):
			result.Warnings = append(result.Warnings, InventoryWarning{
				Code: InventoryWarningLowStock, ProductID: productID, Remaining: remaining,
			})
		}
	}

	result.Warnings = dedupeInventoryWarnings(result.Warnings)
	result.Violations = dedupeLower(result.Violations)
	return result
}

func lowStockThreshold(moq int) int {
	threshold := 100
	if moq > 0 && moq < threshold {
		threshold = moq
	}
	return threshold
}

func dedupeInventoryWarnings(warnings []InventoryWarning) []InventoryWarning {
	seen := make(map[string]struct{}, len(warnings))
	out := make([]InventoryWarning, 0, len(warnings))
	for _, w := range warnings {
		key := w.Code + "|" + w.ProductID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, w)
	}
	return out
}
