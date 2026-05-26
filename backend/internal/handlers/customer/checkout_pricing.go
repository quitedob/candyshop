package customer

import (
	"context"
	"strings"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	countrypkg "candypro/api/internal/pkg/country"
	"candypro/api/internal/pkg/money"
)

type checkoutPricingInput struct {
	Items             []modelsOrder.OrderItem
	ProductByID       map[string]modelsProduct.Product
	ShippingAddress   modelsOrder.Address
	Incoterms         string
	EstimatedWeightKg float64
	Subtotal          float64
	Currency          string
}

type checkoutPricingResult struct {
	TaxAmount      float64
	ShippingAmount float64
	Currency       string
}

// computeCheckoutPricing 服务端计算税/运费，不依赖客户端传值
func (h *Handler) computeCheckoutPricing(ctx context.Context, in checkoutPricingInput) checkoutPricingResult {
	currency := strings.TrimSpace(in.Currency)
	if currency == "" {
		currency = "USD"
	}
	result := checkoutPricingResult{Currency: currency}

	destCountry := countrypkg.NormalizeCountryCode(strings.TrimSpace(in.ShippingAddress.Country))
	incoterms := strings.ToUpper(strings.TrimSpace(in.Incoterms))
	if incoterms == "" {
		incoterms = "FOB"
	}

	estimatedWeightKg := in.EstimatedWeightKg
	if estimatedWeightKg <= 0 {
		for _, item := range in.Items {
			if product, ok := in.ProductByID[item.ProductID]; ok && product.GrossWeightPerCarton > 0 {
				estimatedWeightKg += float64(item.Quantity) * product.GrossWeightPerCarton
			}
		}
	}

	if estimatedWeightKg > 0 && h.services != nil && h.services.Shipping != nil && destCountry != "" {
		if cost, shipCurr, err := h.services.Shipping.CalculateShippingCostWithIncoterms(ctx, destCountry, estimatedWeightKg, incoterms); err == nil && shipCurr != "" {
			result.ShippingAmount = money.RoundMoney(cost)
			result.Currency = shipCurr
		}
	}

	state := normalizeShippingState(in.ShippingAddress.State, destCountry)
	if h.services != nil && h.services.Tax != nil && destCountry != "" {
		if tax, name, _, err := h.services.Tax.CalculateTax(ctx, in.Subtotal, destCountry, state); err == nil && name != "" {
			result.TaxAmount = money.RoundMoney(tax)
		}
	}

	result.ShippingAmount = money.RoundMoney(result.ShippingAmount)
	return result
}

// normalizeShippingState 将 US 州全名规范为两字母代码，提高税率匹配率
func normalizeShippingState(state, country string) string {
	s := strings.TrimSpace(state)
	if s == "" {
		return s
	}
	if country != "US" {
		return s
	}
	if len(s) == 2 {
		return strings.ToUpper(s)
	}
	usStates := map[string]string{
		"new york": "NY", "california": "CA", "texas": "TX", "florida": "FL",
		"illinois": "IL", "pennsylvania": "PA", "ohio": "OH", "georgia": "GA",
		"north carolina": "NC", "michigan": "MI", "new jersey": "NJ", "virginia": "VA",
		"washington": "WA", "arizona": "AZ", "massachusetts": "MA", "tennessee": "TN",
		"indiana": "IN", "missouri": "MO", "maryland": "MD", "wisconsin": "WI",
		"colorado": "CO", "minnesota": "MN", "south carolina": "SC", "alabama": "AL",
		"louisiana": "LA", "kentucky": "KY", "oregon": "OR", "oklahoma": "OK",
		"connecticut": "CT", "utah": "UT", "iowa": "IA", "nevada": "NV",
		"arkansas": "AR", "mississippi": "MS", "kansas": "KS", "new mexico": "NM",
		"nebraska": "NE", "idaho": "ID", "west virginia": "WV", "hawaii": "HI",
		"new hampshire": "NH", "maine": "ME", "montana": "MT", "rhode island": "RI",
		"delaware": "DE", "south dakota": "SD", "north dakota": "ND", "alaska": "AK",
		"vermont": "VT", "wyoming": "WY", "district of columbia": "DC",
	}
	if code, ok := usStates[strings.ToLower(s)]; ok {
		return code
	}
	return s
}
