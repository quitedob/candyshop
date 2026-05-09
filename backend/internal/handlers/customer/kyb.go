package customer

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/kyb"
	"candypro/api/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// cartSubtotalUSD calculates cart line total (unit price * quantity). Non-USD lines counted at face value.
func cartSubtotalUSD(items []modelsOrder.CartItem) float64 {
	var s float64
	for _, it := range items {
		if strings.EqualFold(strings.TrimSpace(it.Currency), "USD") || strings.TrimSpace(it.Currency) == "" {
			s += float64(it.Quantity) * it.UnitPrice
		} else {
			s += float64(it.Quantity) * it.UnitPrice
		}
	}
	return s
}

// projectedCartUSDAfterAdd estimates USD subtotal after adding/merging a line.
func projectedCartUSDAfterAdd(items []modelsOrder.CartItem, productID string, addQty int, unitPrice float64) float64 {
	pid := strings.TrimSpace(productID)
	var oldQty int
	var sum float64
	for _, it := range items {
		if strings.TrimSpace(it.ProductID) == pid {
			oldQty = it.Quantity
			continue
		}
		sum += float64(it.Quantity) * it.UnitPrice
	}
	sum += float64(oldQty+addQty) * unitPrice
	return sum
}

// projectedCartUSDAfterQtyChange estimates USD subtotal after changing a line's quantity.
func projectedCartUSDAfterQtyChange(items []modelsOrder.CartItem, itemID uint, newQty int) float64 {
	var sum float64
	for _, it := range items {
		qty := it.Quantity
		if it.ID == itemID {
			qty = newQty
		}
		sum += float64(qty) * it.UnitPrice
	}
	return sum
}

// ensureActiveOrKYBBypassForAmount allows active users; pending users are subject to KYB bypass limits.
func (h *Handler) ensureActiveOrKYBBypassForAmount(c *gin.Context, userID string, orderTotalUSD float64, lineProductIDs ...string) bool {
	if h.services == nil || h.services.User == nil {
		utils.ServiceUnavailableResp(c)
		return false
	}
	u, err := h.services.User.GetByID(c.Request.Context(), strings.TrimSpace(userID))
	if err != nil || u == nil {
		utils.ErrorResp(c, http.StatusUnauthorized, "user_not_found")
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
		utils.ErrorResp(c, http.StatusForbidden, "kyb_order_limit_exceeded")
		return false
	}
	return true
}
