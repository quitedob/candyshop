package customer

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type customerSendMessageRequest struct {
	Message string `json:"message" binding:"required"`
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

	c.JSON(http.StatusOK, result)
}

// CustomerSendOrderMessage creates a new message from the customer on an order.
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

	var req customerSendMessageRequest
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
		UserID:     userID,
		SenderType: "customer",
		Message:    msg,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := h.services.OrderMessage.SendMessage(c.Request.Context(), message); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "send_message_failed")
		return
	}

	// Notify admin users about the new message
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

	c.JSON(http.StatusCreated, message)
}
