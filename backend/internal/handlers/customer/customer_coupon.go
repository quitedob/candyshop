package customer

import (
	"net/http"
	"strings"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// CustomerValidateCartCoupon handles POST /api/v1/user/cart/coupon/validate — 结账前验证优惠码
func (h *Handler) CustomerValidateCartCoupon(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Code     string  `json:"code" binding:"required"`
		Subtotal float64 `json:"subtotal"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	orderAmount := req.Subtotal
	if orderAmount <= 0 {
		items, err := h.services.Cart.GetCart(c.Request.Context(), userID)
		if err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "cart_fetch_failed")
			return
		}
		for _, it := range items {
			orderAmount += float64(it.Quantity) * it.UnitPrice
		}
	}
	amount, coupon, err := h.services.Coupon.PreviewCouponDiscount(c.Request.Context(), strings.TrimSpace(req.Code), userID, orderAmount)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "coupon_apply_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":           coupon.Code,
		"amount":         amount,
		"discountAmount": amount,
	})
}

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
