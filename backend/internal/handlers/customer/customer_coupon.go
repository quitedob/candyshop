package customer

import (
	"net/http"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// CustomerApplyCoupon handles POST /api/v1/user/cart/coupon
func (h *Handler) CustomerApplyCoupon(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := c.Get("userID")
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		OrderID string `json:"orderId" binding:"required"`
		Code    string `json:"code" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	discount, err := h.services.Coupon.ApplyCouponToCart(c.Request.Context(), req.OrderID, userID.(string), req.Code)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "coupon_apply_failed")
		return
	}
	c.JSON(http.StatusOK, discount)
}

// CustomerRemoveCoupon handles DELETE /api/v1/user/cart/coupon
func (h *Handler) CustomerRemoveCoupon(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		OrderID string `json:"orderId" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if err := h.services.Coupon.RemoveCouponFromCart(c.Request.Context(), req.OrderID); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "coupon_remove_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}
