package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/realtime"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type adminSendMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

// AdminGetOrderMessages returns paginated messages for an order.
func (h *Handler) AdminGetOrderMessages(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	if orderID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if _, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	result, svcErr := h.services.OrderMessage.GetMessages(c.Request.Context(), orderID, page, limit)
	if svcErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fetch_messages_failed")
		return
	}
	h.markAdminOrderMessagesRead(c, orderID)

	c.JSON(http.StatusOK, result)
}

// AdminGetOrderMessageUnreadCount returns unread customer messages for one order.
func (h *Handler) AdminGetOrderMessageUnreadCount(c *gin.Context) {
	orderID, ok := h.authorizeAdminOrderMessages(c)
	if !ok {
		return
	}
	count, err := h.services.OrderMessage.CountUnread(c.Request.Context(), orderID, "admin")
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fetch_messages_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"unreadCount": count})
}

// AdminGetOrderMessageUnreadTotal returns unread customer messages across all orders.
func (h *Handler) AdminGetOrderMessageUnreadTotal(c *gin.Context) {
	if h.services == nil || h.services.OrderMessage == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	count, err := h.services.OrderMessage.CountUnreadForAdmin(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fetch_messages_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"unreadCount": count})
}

// AdminMarkOrderMessagesRead marks unread customer messages as read.
func (h *Handler) AdminMarkOrderMessagesRead(c *gin.Context) {
	orderID, ok := h.authorizeAdminOrderMessages(c)
	if !ok {
		return
	}
	count, readAt, err := h.services.OrderMessage.MarkRead(c.Request.Context(), orderID, "admin")
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "mark_messages_read_failed")
		return
	}
	if h.realtime != nil {
		h.realtime.PublishOrderMessagesRead(orderID, "admin", readAt, count)
	}
	c.JSON(http.StatusOK, gin.H{"readCount": count, "readAt": readAt})
}

// AdminSendOrderMessage creates a new message from an admin on an order.
func (h *Handler) AdminSendOrderMessage(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	rawID, exists := c.Get("userID")
	if !exists {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	adminID, ok := rawID.(string)
	if !ok || adminID == "" {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	if orderID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	var req adminSendMessageRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	now := time.Now()
	message := &modelsOrder.OrderMessage{
		ID:         crypto.GenerateID(),
		OrderID:    orderID,
		UserID:     adminID,
		SenderType: "admin",
		Message:    msg,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.services.OrderMessage.SendMessage(c.Request.Context(), message); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "send_message_failed")
		return
	}

	// Notify the order's customer about the new message
	if h.services.Notification != nil {
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    order.UserID,
			Type:      "order",
			Reference: order.ID,
			Title:     "New Order Message",
			Message:   "Admin sent a message on your order #" + order.OrderNumber + ".",
		})
	}

	// Fan out to any admin/customer WebSocket subscribers on this order's thread.
	if h.realtime != nil {
		h.realtime.PublishOrderMessage(orderID, message)
	}

	c.JSON(http.StatusCreated, message)
}

// AdminStreamOrderMessages upgrades to a WebSocket and streams new messages for
// any order. Admin role gating happens in the route middleware.
func (h *Handler) AdminStreamOrderMessages(c *gin.Context) {
	if h.services == nil || h.realtime == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	if orderID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if _, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	h.realtime.Serve(c.Writer, c.Request, realtime.OrderRoom(orderID))
}

func (h *Handler) authorizeAdminOrderMessages(c *gin.Context) (string, bool) {
	if h.services == nil || h.services.OrderMessage == nil {
		response.ServiceUnavailableResp(c)
		return "", false
	}
	orderID := strings.TrimSpace(c.Param("id"))
	if orderID == "" {
		response.InvalidResp(c, "invalid_request")
		return "", false
	}
	if _, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return "", false
	}
	return orderID, true
}

func (h *Handler) markAdminOrderMessagesRead(c *gin.Context, orderID string) {
	count, readAt, err := h.services.OrderMessage.MarkRead(c.Request.Context(), orderID, "admin")
	if err == nil && h.realtime != nil {
		h.realtime.PublishOrderMessagesRead(orderID, "admin", readAt, count)
	}
}
