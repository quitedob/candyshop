package admin

import (
	modelsTrade "candypro/api/internal/models/trade"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/utils"
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	page, limit := utils.ParsePagination(c, 20, 100)
	orders, err := h.services.Order.GetOrders(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch orders")
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	var req struct {
		Status        string `json:"status"`
		PaymentStatus string `json:"paymentStatus"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	order, err := h.services.Order.GetOrder(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
		return
	}

	previousStatus := strings.ToLower(strings.TrimSpace(order.Status))
	if req.Status != "" {
		order.Status = req.Status
	}
	if req.PaymentStatus != "" {
		order.PaymentStatus = req.PaymentStatus
	}
	targetStatus := strings.ToLower(strings.TrimSpace(order.Status))
	if previousStatus != targetStatus {
		if err := validateStatusTransition(previousStatus, targetStatus); err != nil {
			utils.ErrorResponse(c, http.StatusUnprocessableEntity, "invalid_status_transition", err.Error())
			return
		}
	}

	// Per-customer payment policy: check company PaymentTerms and CreditLimit
	company := h.resolveUserCompany(c, order.UserID)
	if company != nil {
		// Credit limit check before confirmation
		if requiresPaidBeforeExecution(targetStatus) && !isPaidInFull(order.PaymentStatus) {
			if requiresPrepaymentByTerms(company.PaymentTerms) {
				utils.ErrorResponse(c, http.StatusUnprocessableEntity, "payment_policy_violation",
					"Full prepayment is required for this customer's payment terms ("+company.PaymentTerms+")")
				return
			}
			// For NET terms, check credit limit
			if company.CreditLimit > 0 && order.TotalAmount > company.CreditLimit {
				utils.ErrorResponse(c, http.StatusUnprocessableEntity, "credit_limit_exceeded",
					"Order amount exceeds the customer's credit limit")
				return
			}
		}
	} else {
		// Fallback: country-based payment policy for users without a company
		if requiresFullPrepaymentCountry(order.ShippingAddress.Country) &&
			requiresPaidBeforeExecution(targetStatus) &&
			!isPaidInFull(order.PaymentStatus) {
			utils.ErrorResponse(c, http.StatusUnprocessableEntity, "payment_policy_violation",
				"Full prepayment is required before moving orders to confirmed/production/shipping stages")
			return
		}
	}

	if previousStatus != "cancelled" && targetStatus == "cancelled" && order.StockReserved {
		if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to release reserved stock for cancelled order")
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

	if err := h.services.Order.UpdateOrder(c.Request.Context(), order); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update order")
		return
	}

	// Send status change email notification
	if previousStatus != targetStatus && order.User != nil {
		h.services.Order.SendOrderStatusEmail(order, order.User.Email,
			order.User.FirstName+" "+order.User.LastName, targetStatus)
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	orderID := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create trade transaction")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Trade transaction created successfully",
		"trade":   trade,
	})
}
