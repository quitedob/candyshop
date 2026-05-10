package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsTrade "candypro/api/internal/models/trade"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminGetOrders returns all orders for admins.
// @Summary Admin get orders
// @Tags admin-orders
// @Produce json
// @Router /admin/orders [get]
func (h *Handler) AdminGetOrders(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	orders, err := h.services.Order.GetOrders(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, orders)
}

// AdminGetOrder returns a single order details for admins.
// @Summary Admin get order
// @Tags admin-orders
// @Produce json
// @Param id path string true "Order ID"
// @Router /admin/orders/{id} [get]
func (h *Handler) AdminGetOrder(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	c.JSON(http.StatusOK, order)
}

// AdminUpdateOrderStatus updates the status of an order.
// @Summary Admin update order status
// @Tags admin-orders
// @Produce json
// @Param id path string true "Order ID"
// @Param request body map[string]interface{} true "Order status"
// @Router /admin/orders/{id}/status [put]
func (h *Handler) AdminUpdateOrderStatus(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	var req struct {
		Status string `json:"status"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	order, err := h.services.Order.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	previousStatus := strings.ToLower(strings.TrimSpace(order.Status))
	if req.Status != "" {
		order.Status = req.Status
	}
	targetStatus := strings.ToLower(strings.TrimSpace(order.Status))
	if previousStatus != targetStatus {
		if err := modelsOrder.ValidateOrderStatusTransition(previousStatus, targetStatus); err != nil {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "invalid_status_transition")
			return
		}
	}

	// Per-customer payment policy: check company PaymentTerms and CreditLimit
	company := h.resolveUserCompany(c, order.UserID)
	if company != nil {
		// Credit limit check before confirmation
		if requiresPaidBeforeExecution(targetStatus) && !isPaidInFull(order.PaymentStatus) {
			if requiresPrepaymentByTerms(company.PaymentTerms) {
				response.ErrorResp(c, http.StatusUnprocessableEntity, "payment_policy_violation")
				return
			}
			// For NET terms, check credit limit
			if company.CreditLimit > 0 && order.TotalAmount > company.CreditLimit {
				response.ErrorResp(c, http.StatusUnprocessableEntity, "credit_limit_exceeded")
				return
			}
		}
	} else {
		// Fallback: country-based payment policy for users without a company
		if h.requiresFullPrepaymentCountry(order.ShippingAddress.Country) &&
			requiresPaidBeforeExecution(targetStatus) &&
			!isPaidInFull(order.PaymentStatus) {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "payment_policy_violation")
			return
		}
	}

	if previousStatus != "cancelled" && targetStatus == "cancelled" && order.StockReserved {
		if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_stock_release_failed")
			return
		}
	}

	now := time.Now()
	order.UpdatedAt = now
	if targetStatus == "confirmed" && order.ConfirmedAt == nil {
		order.ConfirmedAt = &now
	}
	if targetStatus == "shipped" && order.ShippedAt == nil {
		order.ShippedAt = &now
	}
	if targetStatus == "delivered" && order.DeliveredAt == nil {
		order.DeliveredAt = &now
	}

	if err := h.services.Order.UpdateOrderForAdmin(c.Request.Context(), order, previousStatus, targetStatus); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_update_failed")
		return
	}

	// Send status change email notification
	if previousStatus != targetStatus {
		userEmail := ""
		userName := ""
		if order.User != nil {
			userEmail = order.User.Email
			userName = order.User.FirstName + " " + order.User.LastName
		} else if h.services.User != nil {
			if u, err := h.services.User.GetByID(c.Request.Context(), order.UserID); err == nil && u != nil {
				userEmail = u.Email
				userName = u.FirstName + " " + u.LastName
			}
		}
		h.services.Order.SendOrderStatusEmail(order, userEmail,
			userName, targetStatus)

		// Send in-app notification
		if h.services.Notification != nil {
			_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
				UserID:    order.UserID,
				Type:      "order",
				Reference: order.ID,
				Title:     "Order " + strings.ToUpper(targetStatus[:1]) + targetStatus[1:],
				Message:   "Your order #" + order.OrderNumber + " has been " + targetStatus + ".",
			})
		}
	}

	c.JSON(http.StatusOK, order)
}

// resolveUserCompany looks up the company associated with a user.
func (h *Handler) resolveUserCompany(c *gin.Context, userID string) *modelsUser.Company {
	if h.services == nil || h.services.User == nil {
		return nil
	}
	user, err := h.services.User.GetByID(c.Request.Context(), userID)
	if err != nil {
		return nil
	}
	if user.CompanyID == nil || *user.CompanyID == "" {
		return nil
	}
	company, err := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID)
	if err != nil {
		return nil
	}
	return company
}

// AdminCreateTradeFromOrder creates a TradeTransaction linked to an existing order.
// POST /admin/orders/:id/create-trade
func (h *Handler) AdminCreateTradeFromOrder(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	orderID := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	var req struct {
		Terms    string `json:"terms"`
		Currency string `json:"currency"`
	}
	_ = c.ShouldBindJSON(&req)

	terms := strings.TrimSpace(req.Terms)
	if terms == "" {
		terms = "FOB"
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = order.Currency
	}
	if currency == "" {
		currency = "USD"
	}

	// Check if a trade already exists for this order
	if h.services != nil && h.services.Trade != nil {
		has, err := h.services.Trade.HasTransactionForOrder(c.Request.Context(), orderID)
		if err == nil && has {
			response.ErrorResp(c, http.StatusConflict, "duplicate_trade")
			return
		}
	}

	trade := &modelsTrade.TradeTransaction{
		UserID:      order.UserID,
		OrderID:     &order.ID,
		InquiryID:   order.InquiryID,
		Status:      modelsTrade.TradeStatusPending,
		Currency:    currency,
		TotalAmount: order.TotalAmount,
		Terms:       terms,
	}

	if err := h.services.Trade.CreateTransaction(c.Request.Context(), trade); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "duplicate_trade")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "trade_create_failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Trade transaction created successfully",
		"trade":   trade,
	})
}
