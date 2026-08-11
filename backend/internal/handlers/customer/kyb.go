package customer

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/kyb"
	"candypro/api/internal/pkg/response"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// isUSD reports whether a cart/order currency is USD; empty currency is treated
// as USD, the platform default.
func isUSD(currency string) bool {
	c := strings.ToUpper(strings.TrimSpace(currency))
	return c == "" || c == "USD"
}

// cartSubtotalUSD calculates the USD-equivalent cart line total (unit price *
// quantity) for the KYB cap comparison. Non-USD lines are converted with the
// configured exchange rates; a line whose currency has no known rate fails
// closed (+Inf) so a weaker-currency cart can never slip under the USD cap (M6).
func cartSubtotalUSD(rates map[string]float64, items []modelsOrder.CartItem) float64 {
	var s float64
	for _, it := range items {
		s += kyb.UsdCapValue(rates, it.Currency, float64(it.Quantity)*it.UnitPrice)
	}
	return s
}

// projectedCartUSDAfterAdd estimates the USD subtotal after adding/merging a
// line. The new line's currency and every existing line's currency must be USD
// (empty = USD); otherwise the projection fails closed (+Inf) because the total
// cannot be priced in USD (M6).
func projectedCartUSDAfterAdd(items []modelsOrder.CartItem, productID string, addQty int, unitPrice float64, currency string) float64 {
	if !isUSD(currency) {
		return math.Inf(1)
	}
	pid := strings.TrimSpace(productID)
	var oldQty int
	var sum float64
	for _, it := range items {
		if !isUSD(it.Currency) {
			return math.Inf(1)
		}
		if strings.TrimSpace(it.ProductID) == pid {
			oldQty = it.Quantity
			continue
		}
		sum += float64(it.Quantity) * it.UnitPrice
	}
	sum += float64(oldQty+addQty) * unitPrice
	return sum
}

// projectedCartUSDAfterQtyChange estimates USD subtotal after changing a line's
// quantity. Any line priced in a non-USD currency fails closed (+Inf) (M6).
func projectedCartUSDAfterQtyChange(items []modelsOrder.CartItem, itemID uint, newQty int) float64 {
	var sum float64
	for _, it := range items {
		if !isUSD(it.Currency) {
			return math.Inf(1)
		}
		qty := it.Quantity
		if it.ID == itemID {
			qty = newQty
		}
		sum += float64(qty) * it.UnitPrice
	}
	return sum
}

// ensureActiveOrKYBBypassForAmount allows active users; pending users are subject to KYB bypass limits.
// currency is the ISO currency the orderTotal is denominated in ("" = USD).
func (h *Handler) ensureActiveOrKYBBypassForAmount(c *gin.Context, userID string, orderTotal float64, currency string, lineProductIDs ...string) bool {
	if h.services == nil || h.services.User == nil {
		response.ServiceUnavailableResp(c)
		return false
	}
	u, err := h.services.User.GetByID(c.Request.Context(), strings.TrimSpace(userID))
	if err != nil || u == nil {
		response.ErrorResp(c, http.StatusUnauthorized, "user_not_found")
		return false
	}
	if strings.EqualFold(strings.TrimSpace(u.Status), "active") {
		return true
	}
	tier := kyb.TierConfig{}
	if h.cfg != nil {
		tier.BypassMaxOrderUSD = h.cfg.KYB.BypassMaxOrderUSD
		tier.BypassSampleMaxOrderUSD = h.cfg.KYB.BypassSampleMaxOrderUSD
		tier.SampleProductIDs = h.cfg.KYB.SampleProductIDs
	}
	// Convert to a common currency (USD) before comparing against the caps.
	// Non-USD totals use the configured exchange rates; unsupported currencies
	// fail closed (+Inf) and are never allowed to bypass (M6).
	orderTotalUSD := orderTotal
	if h.cfg != nil {
		orderTotalUSD = kyb.UsdCapValue(h.cfg.ExchangeRates, currency, orderTotal)
	}
	if !kyb.PendingOrderAllowed(tier, orderTotalUSD, lineProductIDs) {
		response.ErrorResp(c, http.StatusForbidden, "kyb_order_limit_exceeded")
		return false
	}
	return true
}
