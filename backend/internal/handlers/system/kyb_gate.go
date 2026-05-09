package system

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/kyb"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ensureActiveOrKYBBypassAmount 与 customer 门户一致：pending 用户适用免审额度与可选样品档位
func (h *Handler) ensureActiveOrKYBBypassAmount(c *gin.Context, userID string, orderTotalUSD float64, lineProductIDs ...string) bool {
	if h.services == nil || h.services.User == nil {
		c.JSON(http.StatusServiceUnavailable, modelsProduct.ErrorResponse{Error: "service_unavailable", Message: "Service not configured"})
		return false
	}
	u, err := h.services.User.GetByID(c.Request.Context(), strings.TrimSpace(userID))
	if err != nil || u == nil {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{Error: "unauthorized", Message: "User not found"})
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
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "kyb_bypass_limit_exceeded",
			Message: "Order exceeds the automatic approval limit for pending accounts.",
		})
		return false
	}
	return true
}
