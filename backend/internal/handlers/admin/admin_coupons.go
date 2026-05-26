package admin

import (
	"fmt"
	"net/http"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminCreateCoupon handles POST /api/v1/admin/coupons
func (h *Handler) AdminCreateCoupon(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Code           string  `json:"code" binding:"required"`
		Type           string  `json:"type" binding:"required"`
		Value          float64 `json:"value" binding:"required,gt=0"`
		MinOrderAmount float64 `json:"minOrderAmount"`
		MaxUses        int     `json:"maxUses"`
		MaxUsesPerUser int     `json:"maxUsesPerUser"`
		StartsAt       string  `json:"startsAt" binding:"required"`
		ExpiresAt      string  `json:"expiresAt" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	// R2 A-3: previously the parse errors were swallowed, leaving coupons with
	// year-0001 timestamps (always-expired). Reject malformed RFC3339 input.
	startsAt, err := time.Parse(time.RFC3339, req.StartsAt)
	if err != nil {
		response.InvalidResp(c, "invalid_starts_at")
		return
	}
	expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
	if err != nil {
		response.InvalidResp(c, "invalid_expires_at")
		return
	}
	if !expiresAt.After(startsAt) {
		response.InvalidResp(c, "invalid_expires_before_starts")
		return
	}

	coupon := &modelsOrder.Coupon{
		ID:             fmt.Sprintf("CP%d", time.Now().UnixNano()),
		Code:           req.Code,
		Type:           req.Type,
		Value:          req.Value,
		MinOrderAmount: req.MinOrderAmount,
		MaxUses:        req.MaxUses,
		MaxUsesPerUser: req.MaxUsesPerUser,
		StartsAt:       startsAt,
		ExpiresAt:      expiresAt,
		Status:         modelsOrder.CouponStatusActive,
	}
	if err := h.services.Coupon.CreateCoupon(c.Request.Context(), coupon); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "coupon_create_failed")
		return
	}
	// H-23: audit coupon lifecycle so abuse / unauthorised edits leave a trail.
	h.logActivityAudit(c, "coupon_create", "coupon", coupon.ID, "", coupon.Code)
	c.JSON(http.StatusCreated, coupon)
}

// AdminListCoupons handles GET /api/v1/admin/coupons
func (h *Handler) AdminListCoupons(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	coupons, total, err := h.services.Coupon.FindAllCoupons(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "coupon_list_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       coupons,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}

// AdminDeleteCoupon handles DELETE /api/v1/admin/coupons/:id
func (h *Handler) AdminDeleteCoupon(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if err := h.services.Coupon.DeleteCoupon(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "coupon_delete_failed")
		return
	}
	h.logActivityAudit(c, "coupon_delete", "coupon", id, "", "")
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// AdminCreateGiftCard handles POST /api/v1/admin/gift-cards
func (h *Handler) AdminCreateGiftCard(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Code           string  `json:"code" binding:"required"`
		InitialBalance float64 `json:"initialBalance" binding:"required,gt=0"`
		Currency       string  `json:"currency"`
		ExpiresAt      string  `json:"expiresAt"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	gc := &modelsOrder.GiftCard{
		ID:             fmt.Sprintf("GC%d", time.Now().UnixNano()),
		Code:           req.Code,
		InitialBalance: req.InitialBalance,
		CurrentBalance: req.InitialBalance,
		Currency:       req.Currency,
		Status:         modelsOrder.GiftCardStatusActive,
	}
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			response.InvalidResp(c, "invalid_expires_at")
			return
		}
		gc.ExpiresAt = &t
	}
	if err := h.services.Coupon.CreateGiftCard(c.Request.Context(), gc); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "giftcard_create_failed")
		return
	}
	h.logActivityAudit(c, "giftcard_create", "giftcard", gc.ID, "", gc.Code)
	c.JSON(http.StatusCreated, gc)
}

// AdminListGiftCards handles GET /api/v1/admin/gift-cards
func (h *Handler) AdminListGiftCards(c *gin.Context) {
	if h.services == nil || h.services.Coupon == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	cards, total, err := h.services.Coupon.FindAllGiftCards(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "giftcard_list_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       cards,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}
