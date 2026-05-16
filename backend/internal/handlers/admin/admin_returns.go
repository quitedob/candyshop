package admin

import (
	"net/http"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminListReturns handles GET /api/v1/admin/returns
func (h *Handler) AdminListReturns(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	status := c.Query("status")
	page, limit := pagination.ParsePagination(c, 20, 100)
	returns, total, err := h.services.Return.FindAll(c.Request.Context(), status, page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_list_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       returns,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}

// AdminGetReturn handles GET /api/v1/admin/returns/:id
func (h *Handler) AdminGetReturn(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	ret, items, err := h.services.Return.FindByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "return_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"return": ret, "items": items})
}

// AdminApproveReturn handles PUT /api/v1/admin/returns/:id/approve
func (h *Handler) AdminApproveReturn(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	userID := c.GetString("userID")
	if err := h.services.Return.UpdateStatus(c.Request.Context(), id, modelsOrder.ReturnStatusApproved, userID, ""); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_approve_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

// AdminReceiveReturn handles PUT /api/v1/admin/returns/:id/receive
func (h *Handler) AdminReceiveReturn(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	userID := c.GetString("userID")
	if err := h.services.Return.UpdateStatus(c.Request.Context(), id, modelsOrder.ReturnStatusReceived, userID, ""); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_receive_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}

// AdminRefundReturn handles PUT /api/v1/admin/returns/:id/refund
func (h *Handler) AdminRefundReturn(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	userID := c.GetString("userID")
	if err := h.services.Return.UpdateStatus(c.Request.Context(), id, modelsOrder.ReturnStatusRefunded, userID, ""); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_refund_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "refunded"})
}

// AdminRejectReturn handles PUT /api/v1/admin/returns/:id/reject
func (h *Handler) AdminRejectReturn(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	userID := c.GetString("userID")
	var req struct {
		Reason string `json:"reason"`
	}
	response.BindJSONOrInvalid(c, &req)
	if err := h.services.Return.UpdateStatus(c.Request.Context(), id, modelsOrder.ReturnStatusRejected, userID, req.Reason); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_reject_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}

