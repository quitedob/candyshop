package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"fmt"
	"strings"
)

// InventoryValidationResult contains validation output for inventory checks.
type InventoryValidationResult struct {
	Warnings   []string `json:"warnings"`
	Violations []string `json:"violations"`
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
		Warnings:   make([]string, 0, 8),
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
			result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s will be sold out after this order.", productID))
		case remaining <= lowStockThreshold(product.MOQ):
			result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has low remaining stock %d after this order.", productID, remaining))
		}
	}

	result.Warnings = dedupeLower(result.Warnings)
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
