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

// computeCheckoutPricing 服务端计算税/运费，不依赖客户端传值。税/运费费率查询
// 失败时返回 error —— 调用方必须 fail closed（500），绝不能把一次真实的 DB
// 查询失败当作"未配置税率/运费"而按 0 入账（G24c）。费率服务在无匹配费率时
// 返回 gorm.ErrRecordNotFound 并折叠为 (0,"",nil)，这是"未覆盖目的地"的合法
// 状态，不是 error；只有真正的查询失败才会传播到这里。
func (h *Handler) computeCheckoutPricing(ctx context.Context, in checkoutPricingInput) (checkoutPricingResult, error) {
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
		cost, shipCurr, err := h.services.Shipping.CalculateShippingCostWithIncoterms(ctx, destCountry, estimatedWeightKg, incoterms)
		if err != nil {
			return checkoutPricingResult{}, fmt.Errorf("compute checkout shipping: %w", err)
		}
		if shipCurr != "" {
			result.ShippingAmount = money.RoundMoney(cost)
			result.Currency = shipCurr
		}
	}

	state := normalizeShippingState(in.ShippingAddress.State, destCountry)
	if h.services != nil && h.services.Tax != nil && destCountry != "" {
		tax, name, _, err := h.services.Tax.CalculateTax(ctx, in.Subtotal, destCountry, state)
		if err != nil {
			return checkoutPricingResult{}, fmt.Errorf("compute checkout tax: %w", err)
		}
		if name != "" {
			result.TaxAmount = money.RoundMoney(tax)
		}
	}

	result.ShippingAmount = money.RoundMoney(result.ShippingAmount)
	return result, nil
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
		unitPrice, perr := h.resolveConfirmUnitPrice(c, contractPriceListID, &product, it.Quantity, country)
		if perr != nil {
			return nil, nil, 0, perr
		}
		it.UnitPrice = unitPrice
		subtotal += float64(it.Quantity) * unitPrice
	}
	return priced, productByID, subtotal, nil
}

// resolveConfirmUnitPrice resolves the server-authoritative unit price for a
// product using the full catalog cascade shared by the confirm re-price paths:
// contract price list → base price → checkout market-cost stack → channel
// multiplier. It returns errRepriceNoPrice when no usable price exists.
// Extracted so repriceOrderItems and the client-added-line branch of
// priceInquiryConfirmItems price a product identically — a base-price-only
// lookup on inquiry confirms would undercharge B2B accounts with a contract
// price list.
func (h *Handler) resolveConfirmUnitPrice(c *gin.Context, contractPriceListID *string, product *modelsProduct.Product, quantity int, country string) (float64, error) {
	var unitPrice float64
	if contractPriceListID != nil && h.services.Price != nil {
		if cp, priceErr := h.services.Price.GetPriceForProduct(c.Request.Context(), product.ID, *contractPriceListID, quantity); priceErr == nil {
			unitPrice = cp
		}
	}
	if unitPrice <= 0 {
		unitPrice = product.BasePrice
	}
	if unitPrice <= 0 {
		return 0, fmt.Errorf("%w: %s", errRepriceNoPrice, product.ID)
	}
	unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), product, unitPrice, country)
	unitPrice = h.applyChannelUnitPrice(c, unitPrice)
	return unitPrice, nil
}

// priceInquiryConfirmItems preserves the server-authoritative unit prices already
// resolved onto an inquiry draft — an accepted negotiation offer's unit price, the
// catalog price for a no-offer conversion, or the OEM project's resolved unit price
// (H10). The catalog / contract price list must NOT override these agreed prices at
// confirm, or the negotiated total is lost and a BasePrice=0 OEM-only offer would 422
// (no_price) on a validly-priced order. Client-supplied confirm prices are still never
// trusted: each line's price is matched back to the draft line by ProductID, and only
// lines the client added that are absent from the draft are priced from the catalog
// (guarded by the same errRepriceNoPrice as the standard re-price path).
func (h *Handler) priceInquiryConfirmItems(c *gin.Context, userID string, draft, items []modelsOrder.OrderItem, country string) ([]modelsOrder.OrderItem, map[string]modelsProduct.Product, float64, error) {
	// Resolve the contract price list for the user (if applicable) so a
	// client-added line is priced with the same cascade as a normal confirm
	// re-price — a base-price-only lookup would undercharge B2B accounts that
	// carry a contract price list.
	var contractPriceListID *string
	if h.services.Price != nil && h.services.User != nil && h.services.Company != nil {
		if usr, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr == nil && usr.CompanyID != nil {
			if company, compErr := h.services.Company.GetCompany(c.Request.Context(), *usr.CompanyID); compErr == nil && company.PriceListID != nil {
				contractPriceListID = company.PriceListID
			}
		}
	}

	draftPrice := make(map[string]float64, len(draft))
	for _, it := range draft {
		pid := strings.TrimSpace(it.ProductID)
		if pid == "" || it.Quantity < 1 || it.UnitPrice <= 0 {
			continue
		}
		draftPrice[pid] = it.UnitPrice
	}

	priced := make([]modelsOrder.OrderItem, len(items))
	copy(priced, items)

	// Batch-load products for validation and to price any client-added lines
	// (mirrors repriceOrderItems).
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
		if dp, ok := draftPrice[pid]; ok {
			// Server-authoritative negotiated / intake price — never overwrite it.
			it.UnitPrice = dp
		} else {
			// Client added a line absent from the draft: price it with the full
			// catalog cascade (contract price list → base price → cost-stack →
			// channel multiplier) so a client-supplied price can never reach the
			// order (H2) and B2B contract pricing is not undercharged.
			unitPrice, perr := h.resolveConfirmUnitPrice(c, contractPriceListID, &product, it.Quantity, country)
			if perr != nil {
				return nil, nil, 0, perr
			}
			it.UnitPrice = unitPrice
		}
		subtotal += float64(it.Quantity) * it.UnitPrice
	}
	return priced, productByID, subtotal, nil
}

// scaleInquiryLinesToTotal re-prices the confirmed lines of an inquiry draft so
// their quantity-weighted sum equals the negotiated TotalAmount exactly (H10,
// invoice-agreement follow-up). The intake draft persists the offer's unit price
// on each line — draft Subtotal = line sum — while Order.TotalAmount carries the
// negotiated deal-level total, which for a round-number / discounted offer
// differs from the line sum. Confirm must NOT persist that divergence: the
// invoice-derivation path (services/order/invoice_policy.go
// CreateInvoiceFromOrder) sets Amount = Subtotal and TotalAmount = Subtotal +
// Tax + Shipping and assumes TotalAmount == Subtotal + Tax + Shipping, so a
// confirmed order with Subtotal 2500 but TotalAmount 2000 would bill the derived
// invoice at 2500 — over-charging a discounted deal by exactly the discount.
// Scaling the lines proportionally (unit × negotiated/line-sum, rounding residual
// absorbed into the last active line) restores the invariant while keeping the
// negotiated total authoritative and the invoice line items consistent with the
// total. Returns the scaled items and the exact subtotal, which equals the
// negotiated total.
func scaleInquiryLinesToTotal(items []modelsOrder.OrderItem, total float64) ([]modelsOrder.OrderItem, float64) {
	if len(items) == 0 || total <= 0 {
		return items, 0
	}
	scaled := make([]modelsOrder.OrderItem, len(items))
	copy(scaled, items)

	lineSum := 0.0
	active := 0
	for _, it := range scaled {
		if it.Quantity < 1 {
			continue
		}
		active++
		lineSum += it.UnitPrice * float64(it.Quantity)
	}
	if active == 0 || lineSum <= 0 {
		return items, 0
	}
	if money.RoundMoney(lineSum) == money.RoundMoney(total) {
		// Already internally consistent — keep the negotiated unit prices verbatim.
		return scaled, money.RoundMoney(total)
	}

	scale := total / lineSum

	// The last active line absorbs the rounding residual so the persisted subtotal
	// is exact; every other line is priced proportionally.
	absorber := -1
	for i := len(scaled) - 1; i >= 0; i-- {
		if scaled[i].Quantity >= 1 {
			absorber = i
			break
		}
	}
	soFar := 0.0
	for i := range scaled {
		if i == absorber || scaled[i].Quantity < 1 {
			continue
		}
		u := money.RoundMoney(scaled[i].UnitPrice * scale)
		if u <= 0 {
			u = 0.01 // never zero out a priced line on a positive deal
		}
		scaled[i].UnitPrice = u
		soFar += u * float64(scaled[i].Quantity)
	}
	if absorber >= 0 {
		remainder := money.RoundMoney(total - soFar)
		q := scaled[absorber].Quantity
		u := money.RoundMoney(remainder / float64(q))
		if u < 0 {
			u = 0
		}
		scaled[absorber].UnitPrice = u
	}
	// The persisted subtotal is the negotiated total exactly (the invoice path and
	// payment gateway derive their amounts from it), regardless of a sub-cent line
	// rounding residual on the absorber.
	return scaled, money.RoundMoney(total)
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
