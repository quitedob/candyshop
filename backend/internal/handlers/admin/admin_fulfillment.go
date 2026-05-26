package admin

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminCreateFulfillment handles POST /api/v1/admin/orders/:id/fulfillments
func (h *Handler) AdminCreateFulfillment(c *gin.Context) {
	if h.services == nil || h.services.Fulfillment == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidResp(c, "missing_order_id")
		return
	}
	var req struct {
		WarehouseID string `json:"warehouseId"`
		Notes       string `json:"notes"`
		Items       []struct {
			OrderItemIdx int    `json:"orderItemIdx" binding:"required,min=0"`
			ProductID    string `json:"productId" binding:"required"`
			Quantity     int    `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required,min=1,dive"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	warehouseID := req.WarehouseID
	if warehouseID == "" {
		var wid *string
		h.assignOrderWarehouseID(c, &wid)
		if wid != nil {
			warehouseID = *wid
		}
	}
	if warehouseID == "" {
		response.InvalidResp(c, "warehouse_id_required")
		return
	}

	fulfillment := &modelsOrder.Fulfillment{
		ID:          fmt.Sprintf("FUL%d", time.Now().UnixNano()),
		OrderID:     orderID,
		WarehouseID: warehouseID,
		Status:      modelsOrder.FulfillmentStatusPending,
		Notes:       req.Notes,
	}
	items := make([]modelsOrder.FulfillmentItem, len(req.Items))
	for i, it := range req.Items {
		items[i] = modelsOrder.FulfillmentItem{
			OrderItemIdx: it.OrderItemIdx,
			ProductID:    it.ProductID,
			Quantity:     it.Quantity,
		}
	}

	if err := h.services.Fulfillment.Create(c.Request.Context(), fulfillment, items, warehouseID); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fulfillment_create_failed")
		return
	}
	h.logActivityAudit(c, "fulfillment_create", "fulfillment", fulfillment.ID, "", orderID)
	c.JSON(http.StatusCreated, fulfillment)
}

// AdminGetFulfillments handles GET /api/v1/admin/fulfillments — paginated list with orderNumber join.
func (h *Handler) AdminGetFulfillments(c *gin.Context) {
	if h.services == nil || h.services.Fulfillment == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	status := strings.TrimSpace(c.Query("status"))

	rows, total, err := h.services.Fulfillment.ListPaginated(c.Request.Context(), page, limit, status)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fulfillment_list_failed")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: rows,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// AdminListFulfillments handles GET /api/v1/admin/orders/:id/fulfillments
func (h *Handler) AdminListFulfillments(c *gin.Context) {
	if h.services == nil || h.services.Fulfillment == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	if orderID == "" {
		response.InvalidResp(c, "missing_order_id")
		return
	}
	fulfillments, err := h.services.Fulfillment.FindByOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fulfillment_list_failed")
		return
	}
	c.JSON(http.StatusOK, fulfillments)
}

// AdminShipFulfillment handles PUT /api/v1/admin/fulfillments/:id/ship
func (h *Handler) AdminShipFulfillment(c *gin.Context) {
	if h.services == nil || h.services.Fulfillment == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if id == "" {
		response.InvalidResp(c, "missing_id")
		return
	}
	var req struct {
		TrackingNumber string `json:"trackingNumber" binding:"required"`
		Carrier        string `json:"carrier"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	now := time.Now()
	if err := h.services.Fulfillment.Ship(c.Request.Context(), id, req.TrackingNumber, req.Carrier, now); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fulfillment_ship_failed")
		return
	}
	h.syncFulfillmentTrackingToOrderAndTrade(c, id, req.TrackingNumber, req.Carrier)
	h.logActivityAudit(c, "fulfillment_ship", "fulfillment", id, req.Carrier, req.TrackingNumber)
	c.JSON(http.StatusOK, gin.H{"status": "shipped"})
}

// AdminDeliverFulfillment handles PUT /api/v1/admin/fulfillments/:id/deliver
func (h *Handler) AdminDeliverFulfillment(c *gin.Context) {
	if h.services == nil || h.services.Fulfillment == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if id == "" {
		response.InvalidResp(c, "missing_id")
		return
	}
	now := time.Now()
	if err := h.services.Fulfillment.Deliver(c.Request.Context(), id, now); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fulfillment_deliver_failed")
		return
	}
	h.logActivityAudit(c, "fulfillment_deliver", "fulfillment", id, "", "delivered")
	c.JSON(http.StatusOK, gin.H{"status": "delivered"})
}

// AdminCancelFulfillment handles PUT /api/v1/admin/fulfillments/:id/cancel
func (h *Handler) AdminCancelFulfillment(c *gin.Context) {
	if h.services == nil || h.services.Fulfillment == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if id == "" {
		response.InvalidResp(c, "missing_id")
		return
	}
	if err := h.services.Fulfillment.Cancel(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fulfillment_cancel_failed")
		return
	}
	h.logActivityAudit(c, "fulfillment_cancel", "fulfillment", id, "", "cancelled")
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}
