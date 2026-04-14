package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"fmt"
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
}

// AdminCreateOrder creates a new order.
func (h *Handler) AdminCreateOrder(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req adminCreateOrderRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		utils.InvalidRequestResponse(c, "userId is required")
		return
	}

	if _, err := h.services.User.GetByID(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "User not found",
		})
		return
	}

	var inquiryID *string
	if req.InquiryID != nil && strings.TrimSpace(*req.InquiryID) != "" {
		trimmed := strings.TrimSpace(*req.InquiryID)
		if _, err := h.services.Inquiry.GetInquiry(c.Request.Context(), trimmed); err != nil {
			c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
				Error:   "not_found",
				Message: "Inquiry not found",
			})
			return
		}
		inquiryID = &trimmed
	}

	orderNumber := strings.TrimSpace(req.OrderNumber)
	if orderNumber == "" {
		orderNumber = fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), strings.ToUpper(utils.GenerateSlug()))
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
		ID:             utils.GenerateID(),
		OrderNumber:    orderNumber,
		UserID:         userID,
		InquiryID:      inquiryID,
		Status:         status,
		PaymentStatus:  paymentStatus,
		Items:          items,
		StockReserved:  len(items) > 0,
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
	if requiresFullPrepaymentCountry(order.ShippingAddress.Country) &&
		requiresPaidBeforeExecution(order.Status) &&
		!isPaidInFull(order.PaymentStatus) {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "payment_policy_violation",
			Message: "Full prepayment is required before moving India/Pakistan orders to confirmed/production/shipping stages",
		})
		return
	}

	var createErr error
	if order.StockReserved {
		createErr = h.services.Order.CreateOrderWithStockReservation(c.Request.Context(), order)
	} else {
		createErr = h.services.Order.CreateOrder(c.Request.Context(), order)
	}
	if createErr != nil {
		if strings.Contains(strings.ToLower(createErr.Error()), "insufficient stock") {
			c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
				Error:   "inventory_violation",
				Message: "Inventory changed while creating order. Please retry with latest stock.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create order",
		})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// AdminUpdateOrder updates an order.
func (h *Handler) AdminUpdateOrder(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	orderID := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}
	previousStatus := strings.ToLower(strings.TrimSpace(order.Status))
	originalItems := append(modelsOrder.OrderItemArray{}, order.Items...)

	var req adminUpdateOrderRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	if req.UserID != nil {
		userID := strings.TrimSpace(*req.UserID)
		if userID == "" {
			utils.InvalidRequestResponse(c, "userId cannot be empty")
			return
		}
		if _, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr != nil {
			c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
				Error:   "not_found",
				Message: "User not found",
			})
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
				c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
					Error:   "not_found",
					Message: "Inquiry not found",
				})
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
		if err := validateStatusTransition(previousStatus, currentStatus); err != nil {
			c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
				Error:   "invalid_status_transition",
				Message: err.Error(),
			})
			return
		}
	}

	if requiresFullPrepaymentCountry(order.ShippingAddress.Country) &&
		requiresPaidBeforeExecution(order.Status) &&
		!isPaidInFull(order.PaymentStatus) {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "payment_policy_violation",
			Message: "Full prepayment is required before moving India/Pakistan orders to confirmed/production/shipping stages",
		})
		return
	}
	if previousStatus != "cancelled" && currentStatus == "cancelled" && order.StockReserved {
		order.UpdatedAt = time.Now()
		if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
			c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to release reserved stock for cancelled order",
			})
			return
		}
		c.JSON(http.StatusOK, order)
		return
	}
	if order.StockReserved && req.Items != nil && currentStatus != "cancelled" {
		stockAdjustment := calculateStockAdjustment(originalItems, order.Items)
		order.UpdatedAt = time.Now()
		if err := h.services.Order.UpdateOrderWithStockAdjustment(c.Request.Context(), order, stockAdjustment); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "insufficient stock") {
				c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
					Error:   "inventory_violation",
					Message: "Inventory changed while updating order items. Please retry with latest stock.",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
				Error:   "internal_error",
				Message: "Failed to update order with inventory adjustment",
			})
			return
		}
		c.JSON(http.StatusOK, order)
		return
	}

	now := time.Now()
	order.UpdatedAt = now
	if order.Status == "confirmed" && order.ConfirmedAt == nil {
		order.ConfirmedAt = &now
	}
	if order.Status == "shipped" && order.ShippedAt == nil {
		order.ShippedAt = &now
	}
	if order.Status == "delivered" && order.DeliveredAt == nil {
		order.DeliveredAt = &now
	}

	if err := h.services.Order.UpdateOrder(c.Request.Context(), order); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update order",
		})
		return
	}

	c.JSON(http.StatusOK, order)
}

// AdminDeleteOrder deletes an order.
func (h *Handler) AdminDeleteOrder(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	orderID := c.Param("id")
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}

	var deleteErr error
	if order.StockReserved {
		deleteErr = h.services.Order.DeleteOrderWithStockRestore(c.Request.Context(), order)
	} else {
		deleteErr = h.services.Order.DeleteOrder(c.Request.Context(), orderID)
	}
	if deleteErr != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete order",
		})
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

// validStatusTransitions defines the allowed order status flow.
// Each key maps to the set of statuses it can transition to.
var validStatusTransitions = map[string]map[string]bool{
	"pending":              {"confirmed": true, "cancelled": true},
	"pending_confirmation": {"pending": true, "confirmed": true, "cancelled": true},
	"confirmed":            {"production": true, "cancelled": true},
	"production":           {"shipped": true, "cancelled": true},
	"shipped":              {"delivered": true},
	"delivered":            {},
	"cancelled":            {},
}

// validateStatusTransition checks whether moving from prev to next is allowed.
func validateStatusTransition(prev, next string) error {
	allowed, known := validStatusTransitions[prev]
	if !known {
		// Unknown current status — allow any transition to avoid blocking legacy data.
		return nil
	}
	if allowed[next] {
		return nil
	}
	return fmt.Errorf("cannot transition order from '%s' to '%s'", prev, next)
}
