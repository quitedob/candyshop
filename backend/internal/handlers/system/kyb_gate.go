package system

import (
	"candypro/api/internal/pkg/kyb"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ensureActiveOrKYBBypassAmount 与 customer 门户一致：pending 用户适用免审额度与可选样品档位
func (h *Handler) ensureActiveOrKYBBypassAmount(c *gin.Context, userID string, orderTotalUSD float64, lineProductIDs ...string) bool {
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
	if !kyb.PendingOrderAllowed(tier, orderTotalUSD, lineProductIDs) {
		response.ErrorResp(c, http.StatusForbidden, "kyb_order_limit_exceeded")
		return false
	}
	return true
}
