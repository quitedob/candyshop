package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/response"
	orderSvc "candypro/api/internal/services/order"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type adminCreateOrderRequest struct {
	OrderNumber     string                  `json:"orderNumber"`
	UserID          string                  `json:"userId" binding:"required"`
	InquiryID       *string                 `json:"inquiryId"`
	Status          string                  `json:"status"`
	PaymentStatus   string                  `json:"paymentStatus"`
	Items           []modelsOrder.OrderItem `json:"items"`
	Subtotal        float64                 `json:"subtotal"`
	TaxAmount       float64                 `json:"taxAmount"`
	ShippingAmount  float64                 `json:"shippingAmount"`
	TotalAmount     float64                 `json:"totalAmount"`
	Currency        string                  `json:"currency"`
	TrackingNumber  string                  `json:"trackingNumber"`
	ShippingAddress *modelsOrder.Address    `json:"shippingAddress"`
}

type adminUpdateOrderRequest struct {
	OrderNumber     *string                  `json:"orderNumber"`
	UserID          *string                  `json:"userId"`
	InquiryID       *string                  `json:"inquiryId"`
	Status          *string                  `json:"status"`
	PaymentStatus   *string                  `json:"paymentStatus"`
	Items           *[]modelsOrder.OrderItem `json:"items"`
	Subtotal        *float64                 `json:"subtotal"`
	TaxAmount       *float64                 `json:"taxAmount"`
	ShippingAmount  *float64                 `json:"shippingAmount"`
	TotalAmount     *float64                 `json:"totalAmount"`
	Currency        *string                  `json:"currency"`
	TrackingNumber  *string                  `json:"trackingNumber"`
	ShippingAddress *modelsOrder.Address     `json:"shippingAddress"`
	// FinancialAdjustmentReason 在订单已处于 confirmed+ 且修改金额/税/运费/币种/行时必填
	FinancialAdjustmentReason string `json:"financialAdjustmentReason"`
}

// AdminCreateOrder creates a new order.
func (h *Handler) AdminCreateOrder(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req adminCreateOrderRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if _, err := h.services.User.GetByID(c.Request.Context(), userID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	var inquiryID *string
	if req.InquiryID != nil && strings.TrimSpace(*req.InquiryID) != "" {
		trimmed := strings.TrimSpace(*req.InquiryID)
		if _, err := h.services.Inquiry.GetInquiry(c.Request.Context(), trimmed); err != nil {
			response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
			return
		}
		inquiryID = &trimmed
	}

	orderNumber := strings.TrimSpace(req.OrderNumber)
	if orderNumber == "" {
		orderNumber = fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), strings.ToUpper(crypto.GenerateSlug()))
	}

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "pending"
	}
	paymentStatus := strings.TrimSpace(req.PaymentStatus)
	if paymentStatus == "" {
		paymentStatus = "unpaid"
	}
	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	items := modelsOrder.OrderItemArray(req.Items)
	if items == nil {
		items = modelsOrder.OrderItemArray{}
	}

	subtotal := req.Subtotal
	if subtotal == 0 && len(items) > 0 {
		for _, item := range items {
			subtotal += float64(item.Quantity) * item.UnitPrice
		}
	}
	total := req.TotalAmount
	if total == 0 {
		total = subtotal + req.TaxAmount + req.ShippingAmount
	}

	order := &modelsOrder.Order{
		ID:             crypto.GenerateID(),
		OrderNumber:    orderNumber,
		UserID:         userID,
		InquiryID:      inquiryID,
		Status:         status,
		PaymentStatus:  paymentStatus,
		Items:          items,
		StockReserved:  len(items) > 0 && requiresPaidBeforeExecution(status),
		Subtotal:       subtotal,
		TaxAmount:      req.TaxAmount,
		ShippingAmount: req.ShippingAmount,
		TotalAmount:    total,
		Currency:       currency,
		TrackingNumber: strings.TrimSpace(req.TrackingNumber),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if req.ShippingAddress != nil {
		order.ShippingAddress = *req.ShippingAddress
	}
	if h.requiresFullPrepaymentForOrder(c.Request.Context(), order.ShippingAddress.Country, order.UserID) &&
		requiresPaidBeforeExecution(order.Status) &&
		!isPaidInFull(order.PaymentStatus) {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "payment_policy_violation")
		return
	}

	var createErr error
	if order.StockReserved {
		createErr = h.services.Order.CreateOrderWithStockReservation(c.Request.Context(), order)
	} else {
		createErr = h.services.Order.CreateOrder(c.Request.Context(), order)
	}
	if createErr != nil {
		if errors.Is(createErr, modelsOrder.ErrInsufficientStock) {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "inventory_violation")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		return
	}

	c.JSON(http.StatusCreated, order)
}

// AdminUpdateOrder updates an order.
func (h *Handler) AdminUpdateOrder(c *gin.Context) {
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
	previousStatus := strings.ToLower(strings.TrimSpace(order.Status))
	originalItems := append(modelsOrder.OrderItemArray{}, order.Items...)
	origFin := orderSvc.SnapshotOrderFinancial(order)

	var req adminUpdateOrderRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.UserID != nil {
		userID := strings.TrimSpace(*req.UserID)
		if userID == "" {
			response.InvalidResp(c, "invalid_request")
			return
		}
		if _, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr != nil {
			response.ErrorResp(c, http.StatusNotFound, "user_not_found")
			return
		}
		order.UserID = userID
	}

	if req.InquiryID != nil {
		inquiryID := strings.TrimSpace(*req.InquiryID)
		if inquiryID == "" {
			order.InquiryID = nil
			order.Inquiry = nil
		} else {
			if _, inquiryErr := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID); inquiryErr != nil {
				response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
				return
			}
			order.InquiryID = &inquiryID
		}
	}

	if req.OrderNumber != nil {
		orderNo := strings.TrimSpace(*req.OrderNumber)
		if orderNo != "" {
			order.OrderNumber = orderNo
		}
	}
	if req.Status != nil {
		order.Status = strings.TrimSpace(*req.Status)
	}
	if req.PaymentStatus != nil {
		order.PaymentStatus = strings.TrimSpace(*req.PaymentStatus)
	}
	if req.Items != nil {
		order.Items = modelsOrder.OrderItemArray(*req.Items)
	}
	if req.Subtotal != nil {
		order.Subtotal = *req.Subtotal
	}
	if req.TaxAmount != nil {
		order.TaxAmount = *req.TaxAmount
	}
	if req.ShippingAmount != nil {
		order.ShippingAmount = *req.ShippingAmount
	}
	if req.TotalAmount != nil {
		order.TotalAmount = *req.TotalAmount
	}
	if req.Currency != nil {
		currency := strings.TrimSpace(*req.Currency)
		if currency != "" {
			order.Currency = currency
		}
	}
	if req.TrackingNumber != nil {
		order.TrackingNumber = strings.TrimSpace(*req.TrackingNumber)
	}
	if req.ShippingAddress != nil {
		order.ShippingAddress = *req.ShippingAddress
	}
	currentStatus := strings.ToLower(strings.TrimSpace(order.Status))

	// Validate status transition before any side-effects.
	if currentStatus != previousStatus {
		if err := modelsOrder.ValidateOrderStatusTransition(previousStatus, currentStatus); err != nil {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "invalid_status_transition")
			return
		}
	}

	if h.requiresFullPrepaymentForOrder(c.Request.Context(), order.ShippingAddress.Country, order.UserID) &&
		requiresPaidBeforeExecution(order.Status) &&
		!isPaidInFull(order.PaymentStatus) {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "payment_policy_violation")
		return
	}

	// Credit limit check for company NET terms
	company := h.resolveUserCompany(c, order.UserID)
	if company != nil && requiresPaidBeforeExecution(currentStatus) && !isPaidInFull(order.PaymentStatus) {
		if !requiresPrepaymentByTerms(company.PaymentTerms) && company.CreditLimit > 0 && order.TotalAmount > company.CreditLimit {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "credit_limit_exceeded")
			return
		}
	}

	newFin := orderSvc.SnapshotOrderFinancial(order)
	if orderSvc.OrderFinancialChanged(origFin, newFin) && orderSvc.OrderStatusRequiresFinancialReason(previousStatus) {
		if strings.TrimSpace(req.FinancialAdjustmentReason) == "" {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "financial_adjustment_reason_required")
			return
		}
	}

	if previousStatus != "cancelled" && currentStatus == "cancelled" && order.StockReserved {
		order.UpdatedAt = time.Now()
		if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_stock_release_failed")
			return
		}
		h.syncOrderFinancialSideEffects(c, order, origFin, newFin, previousStatus, req.FinancialAdjustmentReason)
		if previousStatus != currentStatus && order.User != nil {
			h.services.Order.SendOrderStatusEmail(order, order.User.Email,
				order.User.FirstName+" "+order.User.LastName, currentStatus)
		}
		c.JSON(http.StatusOK, order)
		return
	}
	if order.StockReserved && req.Items != nil && currentStatus != "cancelled" {
		stockAdjustment := calculateStockAdjustment(originalItems, order.Items)
		order.UpdatedAt = time.Now()
		if err := h.services.Order.UpdateOrderWithStockAdjustment(c.Request.Context(), order, stockAdjustment); err != nil {
			if errors.Is(err, modelsOrder.ErrInsufficientStock) {
				response.ErrorResp(c, http.StatusUnprocessableEntity, "inventory_violation")
				return
			}
			response.ErrorResp(c, http.StatusInternalServerError, "order_update_stock_failed")
			return
		}
		h.syncOrderFinancialSideEffects(c, order, origFin, newFin, previousStatus, req.FinancialAdjustmentReason)
		if previousStatus != currentStatus && order.User != nil {
			h.services.Order.SendOrderStatusEmail(order, order.User.Email,
				order.User.FirstName+" "+order.User.LastName, currentStatus)
		}
		c.JSON(http.StatusOK, order)
		return
	}
	// Reserve stock when confirming an order that was created without stock deduction
	if currentStatus == modelsOrder.OrderStatusConfirmed && !order.StockReserved && len(order.Items) > 0 {
		order.UpdatedAt = time.Now()
		if err := h.services.Order.ReserveOrderStock(c.Request.Context(), order); err != nil {
			if errors.Is(err, modelsOrder.ErrInsufficientStock) {
				response.ErrorResp(c, http.StatusUnprocessableEntity, "inventory_violation")
				return
			}
			response.ErrorResp(c, http.StatusInternalServerError, "order_stock_reserve_failed")
			return
		}
	}

	now := time.Now()
	order.UpdatedAt = now
	if order.Status == "confirmed" && order.ConfirmedAt == nil {
		order.ConfirmedAt = &now
		if order.COGS == 0 && h.services.Supplier != nil {
			var totalCOGS float64
			for _, item := range order.Items {
				avgCost := h.services.Supplier.ComputeWeightedAvgCost(c.Request.Context(), item.ProductID)
				totalCOGS += float64(item.Quantity) * avgCost
			}
			order.COGS = totalCOGS
		}
	}
	if order.Status == "shipped" && order.ShippedAt == nil {
		order.ShippedAt = &now
	}
	if order.Status == "delivered" && order.DeliveredAt == nil {
		order.DeliveredAt = &now
	}

	// Route through UpdateOrderForAdmin to trigger outbox when transitioning to confirmed
	if currentStatus != previousStatus {
		if err := h.services.Order.UpdateOrderForAdmin(c.Request.Context(), order, previousStatus, currentStatus); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_update_failed")
			return
		}
	} else {
		if err := h.services.Order.UpdateOrder(c.Request.Context(), order); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_update_failed")
			return
		}
	}

	h.syncOrderFinancialSideEffects(c, order, origFin, newFin, previousStatus, req.FinancialAdjustmentReason)
	if previousStatus != currentStatus && order.User != nil {
		h.services.Order.SendOrderStatusEmail(order, order.User.Email,
			order.User.FirstName+" "+order.User.LastName, currentStatus)
	}
	h.dispatchOrderWebhook(c, order, previousStatus, currentStatus)
	c.JSON(http.StatusOK, order)
}

// dispatchOrderWebhook fires outgoing webhooks for order status transitions.
func (h *Handler) dispatchOrderWebhook(c *gin.Context, order *modelsOrder.Order, prevStatus, newStatus string) {
	if h.services == nil || h.services.Webhook == nil || prevStatus == newStatus {
		return
	}
	var eventType string
	switch newStatus {
	case modelsOrder.OrderStatusConfirmed:
		eventType = modelsOrder.WebhookEventOrderConfirmed
	case modelsOrder.OrderStatusShipped:
		eventType = modelsOrder.WebhookEventOrderShipped
	case modelsOrder.OrderStatusDelivered:
		eventType = modelsOrder.WebhookEventOrderDelivered
	default:
		return
	}
	h.services.Webhook.Dispatch(c.Request.Context(), eventType, order.ID, order)
}

// syncOrderFinancialSideEffects 订单金额变更后同步贸易主单并写审计
func (h *Handler) syncOrderFinancialSideEffects(c *gin.Context, order *modelsOrder.Order, origFin, newFin orderSvc.OrderFinancialSnapshot, previousStatus, reason string) {
	if h.services == nil || order == nil {
		return
	}
	if !orderSvc.OrderFinancialChanged(origFin, newFin) {
		return
	}
	if err := h.services.Trade.SyncTradeTotalFromOrder(c.Request.Context(), order); err != nil {
		log.Printf("Warning: failed to sync trade total for order %s: %v", order.ID, err)
	}
	if orderSvc.OrderStatusRequiresFinancialReason(previousStatus) && strings.TrimSpace(reason) != "" {
		actor := adminActorID(c)
		if err := h.services.Invoice.RecordOrderFinancialAdjustment(c.Request.Context(), order.ID, actor, reason, origFin, newFin); err != nil {
			log.Printf("Warning: failed to record order financial adjustment for order %s: %v", order.ID, err)
		}
	}
}

// AdminDeleteOrder deletes an order.
func (h *Handler) AdminDeleteOrder(c *gin.Context) {
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

	var deleteErr error
	if order.StockReserved {
		deleteErr = h.services.Order.DeleteOrderWithStockRestore(c.Request.Context(), order)
	} else {
		deleteErr = h.services.Order.DeleteOrder(c.Request.Context(), orderID)
	}
	if deleteErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_delete_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order deleted successfully",
		"id":      orderID,
	})
}

func calculateStockAdjustment(oldItems, newItems []modelsOrder.OrderItem) map[string]int {
	oldQty := buildItemQuantityMap(oldItems)
	newQty := buildItemQuantityMap(newItems)

	adjustment := make(map[string]int, len(oldQty)+len(newQty))
	for productID, qty := range newQty {
		adjustment[productID] = qty - oldQty[productID]
	}
	for productID, qty := range oldQty {
		if _, exists := newQty[productID]; exists {
			continue
		}
		adjustment[productID] = -qty
	}
	return adjustment
}

func buildItemQuantityMap(items []modelsOrder.OrderItem) map[string]int {
	qtyByProduct := make(map[string]int, len(items))
	for _, item := range items {
		productID := strings.TrimSpace(item.ProductID)
		if productID == "" || item.Quantity < 1 {
			continue
		}
		qtyByProduct[productID] += item.Quantity
	}
	return qtyByProduct
}
