package system

import (
	"candypro/api/internal/pkg/kyb"
	"candypro/api/internal/pkg/response"
	"math"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// usdCapValue converts a monetary amount to its USD-equivalent using the
// configured exchange rates (base USD). Amounts already in USD pass through;
// empty currency is treated as USD. A currency without a configured rate fails
// closed (+Inf) so it can never satisfy a USD cap (M6: non-USD bypass). Kept in
// sync with the customer portal's identical helper in internal/handlers/customer.
func usdCapValue(rates map[string]float64, currency string, amount float64) float64 {
	cur := strings.ToUpper(strings.TrimSpace(currency))
	if cur == "" || cur == "USD" {
		return amount
	}
	rate, ok := rates[cur]
	if !ok || rate <= 0 {
		return math.Inf(1)
	}
	return amount / rate
}

// ensureActiveOrKYBBypassAmount 与 customer 门户一致：pending 用户适用免审额度与可选样品档位。
// currency 是 orderTotal 的 ISO 币种（"" 视为 USD）。
func (h *Handler) ensureActiveOrKYBBypassAmount(c *gin.Context, userID string, orderTotal float64, currency string, lineProductIDs ...string) bool {
	if h.services == nil || h.services.User == nil {
		response.ErrorResp(c, http.StatusServiceUnavailable, "service_not_configured")
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
	// 统一换算成 USD 后再与额度比较；无汇率的不支持币种 fail-closed（+Inf），绝不放行（M6）。
	orderTotalUSD := orderTotal
	if h.cfg != nil {
		orderTotalUSD = usdCapValue(h.cfg.ExchangeRates, currency, orderTotal)
	}
	if !kyb.PendingOrderAllowed(tier, orderTotalUSD, lineProductIDs) {
		response.ErrorResp(c, http.StatusForbidden, "kyb_order_limit_exceeded")
		return false
	}
	return true
}
