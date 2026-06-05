package customer

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

type customerSendMessageRequest struct {
	Message string `json:"message"`
}

// CustomerGetOrderMessages returns paginated messages for an order.
func (h *Handler) CustomerGetOrderMessages(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
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
	if order.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	result, svcErr := h.services.OrderMessage.GetMessages(c.Request.Context(), orderID, page, limit)
	if svcErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fetch_messages_failed")
		return
	}
	h.markCustomerOrderMessagesRead(c, orderID)

	c.JSON(http.StatusOK, result)
}

// CustomerGetOrderMessageUnreadCount returns unread admin messages for one order.
func (h *Handler) CustomerGetOrderMessageUnreadCount(c *gin.Context) {
	orderID, ok := h.authorizeCustomerOrderMessages(c)
	if !ok {
		return
	}
	count, err := h.services.OrderMessage.CountUnread(c.Request.Context(), orderID, "customer")
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fetch_messages_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"unreadCount": count})
}

// CustomerGetOrderMessageUnreadTotal returns unread admin messages across the customer's orders.
func (h *Handler) CustomerGetOrderMessageUnreadTotal(c *gin.Context) {
	if h.services == nil || h.services.OrderMessage == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	count, err := h.services.OrderMessage.CountUnreadForCustomer(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "fetch_messages_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"unreadCount": count})
}

// CustomerMarkOrderMessagesRead marks unread admin messages as read.
func (h *Handler) CustomerMarkOrderMessagesRead(c *gin.Context) {
	orderID, ok := h.authorizeCustomerOrderMessages(c)
	if !ok {
		return
	}
	count, readAt, err := h.services.OrderMessage.MarkRead(c.Request.Context(), orderID, "customer")
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "mark_messages_read_failed")
		return
	}
	if h.realtime != nil {
		h.realtime.PublishOrderMessagesRead(orderID, "customer", readAt, count)
	}
	c.JSON(http.StatusOK, gin.H{"readCount": count, "readAt": readAt})
}

// CustomerSendOrderMessage creates a new message from the customer on an order (JSON or multipart + files).
func (h *Handler) CustomerSendOrderMessage(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
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
	if order.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	msgText := ""
	var attachments []string

	ct := strings.ToLower(c.GetHeader("Content-Type"))
	if strings.Contains(ct, "multipart/form-data") {
		if err := c.Request.ParseMultipartForm(64 << 20); err != nil {
			response.ErrorResp(c, http.StatusBadRequest, "form_parse_failed")
			return
		}
		msgText = strings.TrimSpace(c.PostForm("message"))
		var upErr error
		attachments, upErr = h.uploadCustomerAttachments(c, c.Request.MultipartForm.File["files"], "order-messages", maxCustomerMessageFiles)
		if upErr != nil {
			return
		}
	} else {
		var req customerSendMessageRequest
		if !response.BindJSONOrInvalid(c, &req) {
			return
		}
		msgText = strings.TrimSpace(req.Message)
	}

	if msgText == "" && len(attachments) == 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}

	now := time.Now()
	message := &modelsOrder.OrderMessage{
		ID:          crypto.GenerateID(),
		OrderID:     orderID,
		UserID:      userID,
		SenderType:  "customer",
		Message:     msgText,
		Attachments: attachments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.services.OrderMessage.SendMessage(c.Request.Context(), message); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "send_message_failed")
		return
	}

	if h.services.User != nil && h.services.Notification != nil {
		adminUsers, userErr := h.services.User.FindAdminUsers(c.Request.Context())
		if userErr == nil {
			for _, admin := range adminUsers {
				_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
					UserID:    admin.ID,
					Type:      "order",
					Reference: order.ID,
					Title:     "New Order Message",
					Message:   "Customer sent a message on order #" + order.OrderNumber + ".",
				})
			}
		}
	}

	// Fan out to any admin/customer WebSocket subscribers on this order's thread.
	if h.realtime != nil {
		h.realtime.PublishOrderMessage(orderID, message)
	}

	c.JSON(http.StatusCreated, message)
}

// CustomerStreamOrderMessages upgrades to a WebSocket and streams new messages
// for an order the authenticated customer owns. Auth + role gating happen in the
// route middleware; here we enforce per-order ownership before the upgrade.
func (h *Handler) CustomerStreamOrderMessages(c *gin.Context) {
	if h.services == nil || h.realtime == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
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
	if order.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	h.realtime.Serve(c.Writer, c.Request, realtime.OrderRoom(orderID))
}

func (h *Handler) authorizeCustomerOrderMessages(c *gin.Context) (string, bool) {
	if h.services == nil || h.services.OrderMessage == nil {
		response.ServiceUnavailableResp(c)
		return "", false
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return "", false
	}
	orderID := strings.TrimSpace(c.Param("id"))
	if orderID == "" {
		response.InvalidResp(c, "invalid_request")
		return "", false
	}
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return "", false
	}
	if order.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return "", false
	}
	return orderID, true
}

func (h *Handler) markCustomerOrderMessagesRead(c *gin.Context, orderID string) {
	count, readAt, err := h.services.OrderMessage.MarkRead(c.Request.Context(), orderID, "customer")
	if err == nil && h.realtime != nil {
		h.realtime.PublishOrderMessagesRead(orderID, "customer", readAt, count)
	}
}
