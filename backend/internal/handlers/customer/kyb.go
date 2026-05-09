package customer

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/kyb"
	"candypro/api/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// cartSubtotalUSD 购物车按行金额汇总（单价 × 数量）
func cartSubtotalUSD(items []modelsOrder.CartItem) float64 {
	var s float64
	for _, it := range items {
		if strings.EqualFold(strings.TrimSpace(it.Currency), "USD") || strings.TrimSpace(it.Currency) == "" {
			s += float64(it.Quantity) * it.UnitPrice
		}
	}
	return s
}

// projectedCartUSDAfterAdd 预估加入/合并行后的美元小计（仅用于 pending 额度校验）
func projectedCartUSDAfterAdd(items []modelsOrder.CartItem, productID string, addQty int, unitPrice float64) float64 {
	pid := strings.TrimSpace(productID)
	var oldQty int
	var sum float64
	for _, it := range items {
		if strings.TrimSpace(it.ProductID) == pid {
			oldQty = it.Quantity
			continue
		}
		if strings.EqualFold(strings.TrimSpace(it.Currency), "USD") || strings.TrimSpace(it.Currency) == "" {
			sum += float64(it.Quantity) * it.UnitPrice
		}
	}
	sum += float64(oldQty+addQty) * unitPrice
	return sum
}

// projectedCartUSDAfterQtyChange 修改某行数量后的美元小计
func projectedCartUSDAfterQtyChange(items []modelsOrder.CartItem, itemID uint, newQty int) float64 {
	var sum float64
	for _, it := range items {
		qty := it.Quantity
		if it.ID == itemID {
			qty = newQty
		}
		if strings.EqualFold(strings.TrimSpace(it.Currency), "USD") || strings.TrimSpace(it.Currency) == "" {
			sum += float64(qty) * it.UnitPrice
		}
	}
	return sum
}

// ensureActiveOrKYBBypassForAmount 已激活用户放行；pending 用户适用 KYB_BYPASS_MAX_ORDER_USD 与可选样品档位 KYB_BYPASS_SAMPLE_MAX_ORDER_USD + KYB_SAMPLE_PRODUCT_IDS
func (h *Handler) ensureActiveOrKYBBypassForAmount(c *gin.Context, userID string, orderTotalUSD float64, lineProductIDs ...string) bool {
	if h.services == nil || h.services.User == nil {
		utils.ServiceUnavailableResponse(c)
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
			Error:   "account_not_active",
			Message: "Your account is pending approval, or the order exceeds the automatic approval limit. Please complete verification or reduce the order value.",
		})
		return false
	}
	return true
}
