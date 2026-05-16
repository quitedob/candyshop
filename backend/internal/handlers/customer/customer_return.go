package customer

import (
	"fmt"
	"net/http"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// CustomerCreateReturn handles POST /api/v1/user/orders/:id/returns
func (h *Handler) CustomerCreateReturn(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	userID, ok := c.Get("userID")
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
		Notes  string `json:"notes"`
		Items  []struct {
			OrderItemIdx int     `json:"orderItemIdx" binding:"required,min=0"`
			ProductID    string  `json:"productId" binding:"required"`
			Quantity     int     `json:"quantity" binding:"required,min=1"`
			ReasonCode   string  `json:"reasonCode" binding:"required"`
			Condition    string  `json:"condition"`
			RefundAmount float64 `json:"refundAmount"`
		} `json:"items" binding:"required,min=1,dive"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	ret := &modelsOrder.ReturnRequest{
		ID:      fmt.Sprintf("RET%d", time.Now().UnixNano()),
		OrderID: orderID,
		UserID:  userID.(string),
		Status:  modelsOrder.ReturnStatusPending,
		Reason:  req.Reason,
		Notes:   req.Notes,
	}
	items := make([]modelsOrder.ReturnItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = modelsOrder.ReturnItem{
			OrderItemIdx: it.OrderItemIdx,
			ProductID:    it.ProductID,
			Quantity:     it.Quantity,
			ReasonCode:   it.ReasonCode,
			Condition:    it.Condition,
			RefundAmount: it.RefundAmount,
		}
	}

	if err := h.services.Return.Create(c.Request.Context(), ret, items); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_create_failed")
		return
	}
	c.JSON(http.StatusCreated, ret)
}

// CustomerListReturns handles GET /api/v1/user/returns
func (h *Handler) CustomerListReturns(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := c.Get("userID")
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	returns, _, err := h.services.Return.FindByUserID(c.Request.Context(), userID.(string), 1, 100)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_list_failed")
		return
	}
	c.JSON(http.StatusOK, returns)
}

// CustomerGetReturn handles GET /api/v1/user/returns/:id
func (h *Handler) CustomerGetReturn(c *gin.Context) {
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
