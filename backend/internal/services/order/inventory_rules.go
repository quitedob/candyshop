package order

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/money"
	"context"
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

// ValidateMOQ re-enforces each product's minimum order quantity against the
// order's aggregated per-product quantity. Draft-creation paths (bulk CSV,
// requisition conversion, reorder-from-history) never validate MOQ, and a
// confirm-time M-19 quantity override can drop a line below MOQ, so the check
// is re-run at confirm before stock is committed (G20). Returns the violation
// strings; an empty slice means every line clears its product MOQ. Products
// absent from productsByID are skipped (they would already have failed the
// re-price step) and products without a configured MOQ always pass.
func (s *OrderService) ValidateMOQ(items []modelsOrder.OrderItem, productsByID map[string]modelsProduct.Product) []string {
	var violations []string
	if len(items) == 0 {
		return violations
	}
	aggregatedQty := make(map[string]int, len(items))
	for _, item := range items {
		pid := strings.TrimSpace(item.ProductID)
		if pid == "" || item.Quantity < 1 {
			continue
		}
		aggregatedQty[pid] += item.Quantity
	}
	for productID, qty := range aggregatedQty {
		product, ok := productsByID[productID]
		if !ok {
			continue
		}
		if product.MOQ > 0 && qty < product.MOQ {
			violations = append(violations, fmt.Sprintf("Product %s requested quantity %d is below MOQ %d.", productID, qty, product.MOQ))
		}
	}
	return violations
}

// openOrderTerminalStatuses are order statuses that no longer represent
// outstanding credit exposure — cancelled / returned / expired orders are
// excluded from the cumulative credit-limit sum.
var openOrderTerminalStatuses = map[string]bool{
	modelsOrder.OrderStatusCancelled: true,
	modelsOrder.OrderStatusReturned:  true,
	modelsOrder.OrderStatusExpired:   true,
}

// SumOpenOrderTotalsByUser returns the sum of total_amount for all open
// (non-terminal) orders of a user, optionally excluding a single order ID —
// the order being confirmed, whose confirmed total the caller adds separately.
// Orders in draft states (pending / pending_confirmation / pending_approval)
// and in-flight states (confirmed → delivered) all count as open exposure.
//
// NOTE (G20 r3): the cumulative credit check at confirm now uses
// SumOpenOrderTotalsByCompany — the credit limit is per COMPANY, and a per-user
// sum lets two buyer users of one company each confirm up to the full limit.
// This per-user helper is retained for callers that genuinely need a single
// buyer's exposure.
func (s *OrderService) SumOpenOrderTotalsByUser(ctx context.Context, userID, excludeOrderID string) (float64, error) {
	const pageSize = 200
	var sum float64
	for page := 1; ; page++ {
		orders, total, err := s.repo.FindByUserID(ctx, userID, page, pageSize)
		if err != nil {
			return 0, err
		}
		for i := range orders {
			o := &orders[i]
			if o.ID == excludeOrderID {
				continue
			}
			if openOrderTerminalStatuses[o.Status] {
				continue
			}
			sum += o.TotalAmount
		}
		if len(orders) == 0 || int64(page*pageSize) >= total {
			break
		}
	}
	return money.RoundMoney(sum), nil
}

// companyOpenOrderSummer is implemented by order repositories that can scope the
// open-order sum to a buyer company. The orderRepository interface (order.go)
// predates the company dimension of credit exposure and cannot be extended here,
// so the concrete OrderRepository advertises the capability through this
// interface and the service asserts it at runtime — the same idiom
// ConfirmAndReserveOrderWithFinancials uses for its extended repo method.
type companyOpenOrderSummer interface {
	SumOpenOrderTotalsByCompany(ctx context.Context, companyID, excludeOrderID string) (float64, error)
}

// SumOpenOrderTotalsByCompany returns the sum of total_amount for all open
// (non-terminal) orders of every user belonging to a buyer company, excluding a
// single order ID. Enforces the cumulative credit limit at confirm against the
// company dimension (G20 r3): the credit limit is per COMPANY, so the sum must
// cover every buyer user of the company — the previous per-user sum let two
// users of one company each confirm up to the full limit.
//
// A repository that cannot scope by company is an ERROR, deliberately failing
// CLOSED: silently falling back to a per-user sum would re-open the exact
// cross-user stacking bypass this method exists to close. The concrete
// OrderRepository always implements it in production.
func (s *OrderService) SumOpenOrderTotalsByCompany(ctx context.Context, companyID, excludeOrderID string) (float64, error) {
	summer, ok := s.repo.(companyOpenOrderSummer)
	if !ok {
		return 0, fmt.Errorf("orderRepository does not support company-scoped open-order sum")
	}
	return summer.SumOpenOrderTotalsByCompany(ctx, companyID, excludeOrderID)
}
