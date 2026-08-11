package customer

import (
	"context"
	"errors"
	"fmt"
	"strings"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	countrypkg "candypro/api/internal/pkg/country"
	"candypro/api/internal/pkg/money"

	"github.com/gin-gonic/gin"
)

// reprice errors — mapped by the confirm handler to HTTP responses.
var (
	errRepriceProductNotFound = errors.New("reprice: product not found")
	errRepriceNoPrice         = errors.New("reprice: product has no price")
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

// repriceOrderItems resolves server-authoritative unit prices for order items
// from the catalog / the user's contract price list. Client-supplied unit
// prices are never trusted — quantities and specifications are preserved but
// every price is recomputed (H1/H2: bulk/requisition drafts confirm at zero
// price, and confirm-time item overrides were accepted without re-pricing).
// It returns the re-priced items, a product lookup map covering all items, and
// the recomputed subtotal.
func (h *Handler) repriceOrderItems(c *gin.Context, userID string, items []modelsOrder.OrderItem, country string) ([]modelsOrder.OrderItem, map[string]modelsProduct.Product, float64, error) {
	priced := make([]modelsOrder.OrderItem, len(items))
	copy(priced, items)

	// Resolve the contract price list for the user (if applicable) — mirrors
	// CustomerCreateOrder so draft and confirm share the same pricing source.
	var contractPriceListID *string
	if h.services.Price != nil && h.services.User != nil && h.services.Company != nil {
		if usr, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr == nil && usr.CompanyID != nil {
			if company, compErr := h.services.Company.GetCompany(c.Request.Context(), *usr.CompanyID); compErr == nil && company.PriceListID != nil {
				contractPriceListID = company.PriceListID
			}
		}
	}

	// Batch-load products for every distinct item id.
	productByID := make(map[string]modelsProduct.Product, len(priced))
	uniqueIDs := make([]string, 0, len(priced))
	idSeen := make(map[string]struct{}, len(priced))
	for _, it := range priced {
		pid := strings.TrimSpace(it.ProductID)
		if pid == "" || it.Quantity < 1 {
			continue
		}
		if _, ok := idSeen[pid]; ok {
			continue
		}
		idSeen[pid] = struct{}{}
		uniqueIDs = append(uniqueIDs, pid)
	}
	if len(uniqueIDs) > 0 {
		batch, batchErr := h.services.Product.GetProductsByIDs(c.Request.Context(), uniqueIDs)
		if batchErr != nil {
			return nil, nil, 0, batchErr
		}
		for i := range batch {
			productByID[batch[i].ID] = batch[i]
		}
	}

	subtotal := 0.0
	for i := range priced {
		it := &priced[i]
		pid := strings.TrimSpace(it.ProductID)
		if pid == "" || it.Quantity < 1 {
			continue
		}
		product, ok := productByID[pid]
		if !ok {
			return nil, nil, 0, fmt.Errorf("%w: %s", errRepriceProductNotFound, pid)
		}
		if status := strings.ToLower(strings.TrimSpace(product.Status)); status != "" && status != "active" {
			return nil, nil, 0, fmt.Errorf("%w: %s", errRepriceProductNotFound, pid)
		}
		// Price resolution mirrors CustomerCreateOrder: contract price list →
		// base price → checkout market-cost stack → channel multiplier.
		var unitPrice float64
		if contractPriceListID != nil && h.services.Price != nil {
			if cp, priceErr := h.services.Price.GetPriceForProduct(c.Request.Context(), pid, *contractPriceListID, it.Quantity); priceErr == nil {
				unitPrice = cp
			}
		}
		if unitPrice <= 0 {
			unitPrice = product.BasePrice
		}
		if unitPrice <= 0 {
			return nil, nil, 0, fmt.Errorf("%w: %s", errRepriceNoPrice, pid)
		}
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), &product, unitPrice, country)
		unitPrice = h.applyChannelUnitPrice(c, unitPrice)
		it.UnitPrice = unitPrice
		subtotal += float64(it.Quantity) * unitPrice
	}
	return priced, productByID, subtotal, nil
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
