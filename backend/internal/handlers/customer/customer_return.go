package customer

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"
	orderRepo "candypro/api/internal/repository/order"

	"github.com/gin-gonic/gin"
)

// orderStatusAllowsReturn reports whether an order has reached a lifecycle
// stage in which a customer may file a return (goods already shipped or
// delivered). It mirrors the statuses from which ValidOrderStatusTransitions
// permits a move toward partially_returned/returned — orders still in
// pending/confirmed/production have not shipped anything returnable yet.
func orderStatusAllowsReturn(status string) bool {
	switch status {
	case modelsOrder.OrderStatusPartiallyShipped,
		modelsOrder.OrderStatusShipped,
		modelsOrder.OrderStatusPartiallyDelivered,
		modelsOrder.OrderStatusDelivered,
		modelsOrder.OrderStatusPartiallyReturned:
		return true
	}
	return false
}

// returnCreateMu serializes the already-returned tally and the ReturnRequest
// insert in CustomerCreateReturn so two concurrent POSTs for the same order
// cannot both pass the remaining-quantity ceiling and both persist (H2 TOCTOU).
// The lock is per-process: it closes the race for the single-instance deployment
// this monolith targets. A multi-instance deployment additionally needs a
// DB-level guard — a partial unique index on (order_id, order_item_idx) where
// status <> 'rejected', or a SELECT ... FOR UPDATE / version-guarded insert in
// the ReturnRepository — which is a repository/model change outside this
// handler's scope.
var returnCreateMu sync.Mutex

// sumReturnedQuantities walks every page of the caller's return requests and
// totals the quantity already claimed per order-line index by non-rejected
// returns on the given order. FindByUserID is page-limited, so reading only the
// first page would drop the oldest claims and let a customer with many returns
// undercount what they have already claimed and over-return (H2).
func sumReturnedQuantities(ctx context.Context, repo *orderRepo.ReturnRepository, userID, orderID string) (map[int]int, error) {
	returnedQty := make(map[int]int)
	const pageSize = 500
	for page := 1; ; page++ {
		existing, total, err := repo.FindByUserID(ctx, userID, page, pageSize)
		if err != nil {
			return nil, err
		}
		for _, r := range existing {
			if r.OrderID != orderID || r.Status == modelsOrder.ReturnStatusRejected {
				continue
			}
			for _, it := range r.Items {
				returnedQty[it.OrderItemIdx] += it.Quantity
			}
		}
		if page*pageSize >= int(total) {
			return returnedQty, nil
		}
	}
}

// CustomerCreateReturn handles POST /api/v1/user/orders/:id/returns
func (h *Handler) CustomerCreateReturn(c *gin.Context) {
	if h.services == nil || h.services.Return == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
		Notes  string `json:"notes"`
		Items  []struct {
			OrderItemIdx int     `json:"orderItemIdx" binding:"min=0"`
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

	// H2: the create path used to blind-insert a ReturnRequest keyed on
	// client-supplied orderID + raw-context userID. That let any authenticated
	// customer file returns against other users' orders with arbitrary
	// quantities and refund amounts. Load the order and validate ownership,
	// lifecycle status, per-line quantities and refund ceilings up front.
	if h.services.Order == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	if order.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	if !orderStatusAllowsReturn(order.Status) {
		response.InvalidResp(c, "return_status_not_allowed")
		return
	}

	// Hold the per-process lock across the already-returned tally, the
	// remaining-quantity/refund checks, and the insert so two concurrent creates
	// for the same order+line cannot both pass the ceiling (H2 TOCTOU). The
	// second caller re-reads the tally after the first commits and is rejected.
	// The lock is released via defer; emit/json are fast in-process DB calls, so
	// the hold time is bounded and return creation is low-volume.
	returnCreateMu.Lock()
	defer returnCreateMu.Unlock()

	// Tally of quantity already claimed by non-rejected returns on this order so
	// a customer cannot double-return a line. Read across all pages (fail-closed
	// on error: if we cannot verify the tally we must not create a claim).
	returnedQty, err := sumReturnedQuantities(c.Request.Context(), h.services.Return, userID, orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "return_create_failed")
		return
	}

	// Index the order's lines so each return item must map to a real line with
	// a matching product before its quantity/refund can be trusted.
	lineByIdx := make(map[int]modelsOrder.OrderItem, len(order.Items))
	for idx, li := range order.Items {
		lineByIdx[idx] = li
	}

	var refundTotal float64
	for _, it := range req.Items {
		line, ok := lineByIdx[it.OrderItemIdx]
		if !ok || line.ProductID != it.ProductID {
			response.InvalidResp(c, "return_item_invalid")
			return
		}
		// Only goods actually shipped/fulfilled can be returned. When the line
		// carries shipped/fulfilled counts, cap the claimable quantity to them
		// (min(ordered, max(shipped, fulfilled))); fall back to the ordered
		// quantity when neither is populated (some flows omit these fields).
		received := line.ShippedQuantity
		if line.FulfilledQuantity > received {
			received = line.FulfilledQuantity
		}
		returnable := line.Quantity
		if received > 0 && received < returnable {
			returnable = received
		}
		remaining := returnable - returnedQty[it.OrderItemIdx]
		if it.Quantity > remaining {
			response.InvalidResp(c, "return_item_quantity_exceeds_ordered")
			return
		}
		returnedQty[it.OrderItemIdx] += it.Quantity
		lineTotal := line.UnitPrice * float64(it.Quantity)
		if it.RefundAmount < 0 || it.RefundAmount > lineTotal {
			response.InvalidResp(c, "return_refund_exceeds_line")
			return
		}
		refundTotal += it.RefundAmount
	}
	if refundTotal > order.TotalAmount {
		response.InvalidResp(c, "return_refund_exceeds_order_total")
		return
	}

	ret := &modelsOrder.ReturnRequest{
		ID:      fmt.Sprintf("RET%d", time.Now().UnixNano()),
		OrderID: orderID,
		UserID:  userID,
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
	// Return the persisted lines in the response — Create populated ReturnID and
	// timestamps on items, so this reflects the stored rows (H2).
	ret.Items = items
	h.emitLifecycleEvent(c, modelsOrder.WebhookEventReturnCreated, ret.ID, ret)
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
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := c.Param("id")
	ret, items, err := h.services.Return.FindByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "return_not_found")
		return
	}
	if ret.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	c.JSON(http.StatusOK, gin.H{"return": ret, "items": items})
}
