package customer

import (
	"context"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	countrypkg "candypro/api/internal/pkg/country"
	"strings"

	"github.com/gin-gonic/gin"
)

// 费用估算状态：区分「已算出参考值」与「尚无法估算」
const (
	feeEstimateComputed           = "computed"            // 系统按费率表算出（含 0 元）
	feeEstimatePendingDestination = "pending_destination" // 缺收货国
	feeEstimatePendingRates       = "pending_rates"       // 无匹配费率
	feeEstimatePendingWeight      = "pending_weight"      // 缺重量数据
)

// buildCartTaxEstimate 购物车税费参考估算（非最终报价）
func (h *Handler) buildCartTaxEstimate(ctx context.Context, subtotal float64, destination, region string) gin.H {
	destination = countrypkg.NormalizeCountryCode(strings.TrimSpace(destination))
	if destination == "" {
		return gin.H{"status": feeEstimatePendingDestination}
	}
	if h.services == nil || h.services.Tax == nil {
		return gin.H{"status": feeEstimatePendingRates}
	}
	tax, name, ratePct, err := h.services.Tax.CalculateTax(ctx, subtotal, destination, region)
	if err != nil || name == "" {
		return gin.H{"status": feeEstimatePendingRates}
	}
	return gin.H{
		"status":   feeEstimateComputed,
		"amount":   tax,
		"rate":     ratePct,
		"rateName": name,
	}
}

// buildCartShippingEstimate 购物车运费参考估算（非最终报价）
func (h *Handler) buildCartShippingEstimate(ctx context.Context, destination string, weightKg float64, incoterms string) gin.H {
	destination = countrypkg.NormalizeCountryCode(strings.TrimSpace(destination))
	if destination == "" {
		return gin.H{"status": feeEstimatePendingDestination}
	}
	if weightKg <= 0 {
		return gin.H{"status": feeEstimatePendingWeight}
	}
	if h.services == nil || h.services.Shipping == nil {
		return gin.H{"status": feeEstimatePendingRates}
	}
	cost, currency, err := h.services.Shipping.CalculateShippingCostWithIncoterms(ctx, destination, weightKg, incoterms)
	if err != nil {
		return gin.H{"status": feeEstimatePendingRates}
	}
	inc := strings.ToUpper(strings.TrimSpace(incoterms))
	if inc == "EXW" {
		return gin.H{
			"status":    feeEstimateComputed,
			"cost":      0,
			"currency":  currency,
			"weightKg":  weightKg,
			"incoterms": "EXW",
		}
	}
	if strings.TrimSpace(currency) == "" {
		return gin.H{"status": feeEstimatePendingRates}
	}
	return gin.H{
		"status":   feeEstimateComputed,
		"cost":     cost,
		"currency": currency,
		"weightKg": weightKg,
	}
}

// sumCartWeightKg 按购物车行汇总毛重
func sumCartWeightKg(items []modelsOrder.CartItem, productByID map[string]modelsProduct.Product) float64 {
	var weightKg float64
	for _, it := range items {
		if product, ok := productByID[it.ProductID]; ok && product.GrossWeightPerCarton > 0 {
			weightKg += float64(it.Quantity) * product.GrossWeightPerCarton
		}
	}
	return weightKg
}

// pricingNoticeForCheckout 下单响应中的金额说明
func pricingNoticeForCheckout(taxStatus, shipStatus string) string {
	if taxStatus == feeEstimateComputed && shipStatus == feeEstimateComputed {
		return "reference_until_pi"
	}
	if taxStatus == feeEstimateComputed || shipStatus == feeEstimateComputed {
		return "partial_until_pi"
	}
	return "pending_until_pi"
}

// checkoutTaxStatus 下单时税费项状态（用于 pricing 说明）
func checkoutTaxStatus(ctx context.Context, h *Handler, country, region string, subtotal, taxAmount float64) string {
	country = countrypkg.NormalizeCountryCode(strings.TrimSpace(country))
	if country == "" {
		return feeEstimatePendingDestination
	}
	if h.services == nil || h.services.Tax == nil {
		return feeEstimatePendingRates
	}
	_, name, _, err := h.services.Tax.CalculateTax(ctx, subtotal, country, region)
	if err != nil || name == "" {
		return feeEstimatePendingRates
	}
	return feeEstimateComputed
}

// checkoutShippingStatus 下单时运费项状态（与购物车估算逻辑一致）
func checkoutShippingStatus(ctx context.Context, h *Handler, country string, weightKg float64, incoterms string) string {
	est := h.buildCartShippingEstimate(ctx, country, weightKg, incoterms)
	if s, ok := est["status"].(string); ok && s != "" {
		return s
	}
	return feeEstimatePendingRates
}
