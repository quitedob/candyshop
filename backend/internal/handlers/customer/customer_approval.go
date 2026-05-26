package customer

import (
	"candypro/api/internal/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CustomerApproveOrder approves a pending-approval order (org approver only).
func (h *Handler) CustomerApproveOrder(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	var req struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.services.Approval.ApproveOrder(c.Request.Context(), orderID, userID, req.Comment); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "order_approve_failed")
		return
	}
	updated, _ := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if updated != nil {
		h.notifyOrderApprovalResult(c, updated, true)
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved", "orderId": order.ID})
}

// CustomerRejectOrder rejects a pending-approval order (org approver only).
func (h *Handler) CustomerRejectOrder(c *gin.Context) {
	if h.services == nil || h.services.Approval == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	var req struct {
		Comment string `json:"comment"`
	}
	_ = c.ShouldBindJSON(&req)
	if err := h.services.Approval.RejectOrder(c.Request.Context(), orderID, userID, req.Comment); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "order_reject_failed")
		return
	}
	updated, _ := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if updated != nil {
		h.notifyOrderApprovalResult(c, updated, false)
	} else {
		h.notifyOrderApprovalResult(c, order, false)
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected", "orderId": orderID})
}
