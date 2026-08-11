package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsTrade "candypro/api/internal/models/trade"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/request"
	"candypro/api/internal/pkg/response"
	orderSvc "candypro/api/internal/services/order"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	_ modelsCommon.PaginatedResponse
	_ modelsOrder.Order
)

// AdminGetOrders returns all orders for admins.
// @Summary Admin get orders
// @Tags admin-orders
// @Produce json
// @Success 200 {object} modelsCommon.PaginatedResponse
// @Router /admin/orders [get]
func (h *Handler) AdminGetOrders(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	status := strings.TrimSpace(c.Query("status"))
	userID := strings.TrimSpace(c.Query("userId"))
	dateFrom := strings.TrimSpace(c.Query("dateFrom"))
	dateTo := strings.TrimSpace(c.Query("dateTo"))
	orders, err := h.services.Order.GetOrders(c.Request.Context(), page, limit, status, userID, dateFrom, dateTo)
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

	resp := orderDetailResponse{Order: order, StatusHistory: []statusHistoryEntry{}, ActivityHistory: []activityHistoryEntry{}}
	if h.services.ActivityLog != nil {
		if logs, logErr := h.services.ActivityLog.FindByEntity(c.Request.Context(), "order", id); logErr == nil {
			resp.StatusHistory = buildOrderStatusHistory(logs)
			resp.ActivityHistory = buildOrderActivityHistory(logs)
		}
	}
	resp.InventoryWarnings = h.buildOrderInventoryWarnings(c, order)

	c.JSON(http.StatusOK, resp)
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
		Status              string `json:"status"`
		TrackingNumber      string `json:"trackingNumber"`
		TrackingNumberSnake string `json:"tracking_number"`
		// Refund, when true, authorises the handler to refund any confirmed
		// payments before cancelling. Without it, cancelling an order that has
		// captured money is blocked so funds are never silently stranded
		// (P0.3 / G-ORD-3).
		Refund bool `json:"refund"`
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
		order.Status = strings.TrimSpace(req.Status)
	}
	if tracking := request.FirstNonEmpty(req.TrackingNumber, req.TrackingNumberSnake); tracking != "" {
		order.TrackingNumber = tracking
	}
	targetStatus := strings.ToLower(strings.TrimSpace(order.Status))
	if previousStatus != targetStatus {
		if err := modelsOrder.ValidateOrderStatusTransition(previousStatus, targetStatus); err != nil {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "invalid_status_transition")
			return
		}
	}

	// 执行阶段（confirmed 及以后）须已付款或部分付款
	if paymentInsufficientForExecution(targetStatus, order.PaymentStatus) {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "payment_policy_violation")
		return
	}

	// G20 (follow-up): the admin confirm path must re-enforce the same policy the
	// customer confirm path does — per-product MOQ and the buyer company's
	// CUMULATIVE credit exposure — before stock is committed. Drafts
	// (bulk/requisition/reorder) never validated MOQ or the contract price-list
	// min, and an admin confirming under-limit orders directly bypassed the
	// customer's cumulative credit guard, letting a company stack several
	// under-limit orders past its limit. The customer confirm path enforces all
	// three; the admin confirm path must not be a bypass.
	if targetStatus == modelsOrder.OrderStatusConfirmed && previousStatus != targetStatus {
		if !h.validateAdminConfirmMOQ(c, order.Items) {
			return
		}
		if !h.validateAdminConfirmPriceListMin(c, order.UserID, order.Items) {
			return
		}
		if !h.checkAdminCompanyCreditLimitCumulative(c, order.UserID, order.ID, order.TotalAmount) {
			return
		}
	}

	if previousStatus != "cancelled" && targetStatus == "cancelled" && order.StockReserved {
		if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_stock_release_failed")
			return
		}
	}

	// P0.3 / G-ORD-3: cancelling an order with confirmed (captured) payments must
	// not silently strand money behind an audit log. Either refund every
	// confirmed payment (when the admin explicitly requests it) or block the
	// cancellation so finance makes a deliberate decision.
	if previousStatus != "cancelled" && targetStatus == "cancelled" && h.services.Payment != nil {
		count, total, perr := h.services.Payment.CountConfirmedPaymentsForOrder(c.Request.Context(), order.ID)
		if perr != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "payment_fetch_failed")
			return
		}
		if count > 0 {
			if !req.Refund {
				// Block: surface the captured amount so the admin can re-submit
				// with refund=true once they've confirmed the refund is intended.
				response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "order_cancel_requires_refund", gin.H{
					"confirmedPaymentCount": count,
					"confirmedTotal":        total,
					"currency":              order.Currency,
				})
				return
			}
			if refunded, rerr := h.refundConfirmedPayments(c, order.ID); rerr != nil {
				h.logOrderAudit(c, "order_cancel_refund_failed",
					order.ID, c.GetString("userID"),
					fmt.Sprintf("refunded=%d/%d", refunded, count),
					rerr.Error(),
				)
				response.ErrorResp(c, http.StatusBadGateway, "refund_failed")
				return
			}
			h.logOrderAudit(c, "order_cancel_with_refund",
				order.ID, c.GetString("userID"),
				fmt.Sprintf("count=%d", count),
				fmt.Sprintf("total=%.2f %s", total, order.Currency),
			)
		}
	}

	now := time.Now()
	order.UpdatedAt = now
	if targetStatus == modelsOrder.OrderStatusConfirmed && previousStatus != targetStatus && h.services.Product != nil {
		order.COGS = orderSvc.ComputeOrderCOGS(c.Request.Context(), order.Items, h.services.Product)
	}

	// H-12: re-validate payment status against the freshest read just before
	// invoking the transactional update. This shrinks the check-then-act window
	// — a webhook flipping payment between the original GetOrder() and the write
	// would otherwise sneak past `paymentInsufficientForExecution`. The repo
	// update itself takes a row lock, so this fetch + repo guard together
	// approximate a SELECT FOR UPDATE on payment_status.
	//
	// R2 A-10: also detect concurrent admin/webhook writes by comparing Version
	// on the freshest read against the version we loaded at the top. If it
	// changed, our in-memory mutations were derived from stale state and the
	// safest action is to ask the operator to refresh and re-decide rather
	// than silently overwrite the competing change.
	if requiresPaidBeforeExecution(targetStatus) {
		fresh, freshErr := h.services.Order.GetOrder(c.Request.Context(), order.ID)
		if freshErr == nil && fresh != nil {
			if paymentInsufficientForExecution(targetStatus, fresh.PaymentStatus) {
				response.ErrorResp(c, http.StatusUnprocessableEntity, "payment_policy_violation")
				return
			}
			if fresh.Version != order.Version {
				response.ErrorResp(c, http.StatusConflict, "order_modified_concurrently")
				return
			}
		}
	} else {
		// Cheap version recheck for non-execution transitions too.
		if fresh, freshErr := h.services.Order.GetOrder(c.Request.Context(), order.ID); freshErr == nil && fresh != nil && fresh.Version != order.Version {
			response.ErrorResp(c, http.StatusConflict, "order_modified_concurrently")
			return
		}
	}
	if targetStatus == "confirmed" && order.ConfirmedAt == nil {
		order.ConfirmedAt = &now
	}
	if targetStatus == modelsOrder.OrderStatusShipped && order.ShippedAt == nil {
		order.ShippedAt = &now
	}
	if targetStatus == modelsOrder.OrderStatusDelivered && order.DeliveredAt == nil {
		order.DeliveredAt = &now
	}

	if err := h.services.Order.UpdateOrderForAdmin(c.Request.Context(), order, previousStatus, targetStatus); err != nil {
		if errors.Is(err, modelsOrder.ErrInsufficientStock) {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "inventory_violation")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "order_update_failed")
		return
	}

	// Send status change email notification
	if previousStatus != targetStatus {
		h.logOrderAudit(c, "order_status_change", order.ID, c.GetString("userID"), previousStatus, targetStatus)
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
			statusLabel := targetStatus
			statusKey := "enum.order_status_" + strings.ReplaceAll(targetStatus, "-", "_")
			if msg := i18n.T(c, statusKey); msg != statusKey {
				statusLabel = msg
			}
			vars := map[string]string{
				"orderNumber": order.OrderNumber,
				"statusLabel": statusLabel,
			}
			_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
				UserID:    order.UserID,
				Type:      "order",
				Reference: order.ID,
				Title:     i18n.TWithVars(c, "notifications.order_status_updated_title", vars),
				Message:   i18n.TWithVars(c, "notifications.order_status_updated_message", vars),
			})
		}
	}

	c.JSON(http.StatusOK, order)
}

// refundConfirmedPayments refunds every confirmed payment on the order via the
// gateway service, which routes gateway-backed payments through Stripe/PayPal
// and falls back to a local status flip for manual (bank transfer) payments.
// Returns the number of successfully refunded payments and the first error.
func (h *Handler) refundConfirmedPayments(c *gin.Context, orderID string) (int, error) {
	if h.services == nil || h.services.Payment == nil {
		return 0, fmt.Errorf("payment service unavailable")
	}
	payments, err := h.services.Payment.GetPaymentsByOrder(c.Request.Context(), orderID)
	if err != nil {
		return 0, err
	}
	refunded := 0
	for i := range payments {
		p := payments[i]
		if p.Status != modelsOrder.PaymentRecordStatusConfirmed {
			continue
		}
		if h.services.GatewayPayment != nil {
			if rerr := h.services.GatewayPayment.RefundGateway(c.Request.Context(), &p); rerr != nil {
				return refunded, rerr
			}
		} else if rerr := h.services.Payment.RefundPayment(c.Request.Context(), p.ID); rerr != nil {
			return refunded, rerr
		}
		refunded++
	}
	return refunded, nil
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
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

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
