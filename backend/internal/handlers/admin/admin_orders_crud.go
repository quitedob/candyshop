package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/request"
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
	Items           []modelsOrder.OrderItem `json:"items" binding:"required,min=1,dive"`
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
	PaymentStatusSnake *string               `json:"payment_status"`
	Items           *[]modelsOrder.OrderItem `json:"items"`
	Subtotal        *float64                 `json:"subtotal"`
	TaxAmount       *float64                 `json:"taxAmount"`
	ShippingAmount  *float64                 `json:"shippingAmount"`
	TotalAmount     *float64                 `json:"totalAmount"`
	Currency        *string                  `json:"currency"`
	TrackingNumber  *string                  `json:"trackingNumber"`
	TrackingNumberSnake *string              `json:"tracking_number"`
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

	country := ""
	if req.ShippingAddress != nil {
		country = strings.TrimSpace(req.ShippingAddress.Country)
	}

	// C-1: Admin-created orders must enforce the same business validations
	// that customer checkout does — MOQ, server-side pricing, inventory,
	// compliance, and credit limit. The previous implementation accepted
	// client-supplied prices and skipped MOQ/inventory/credit checks.
	contractPriceListID := h.resolveAdminContractPriceListID(c, userID)

	items, productByID, products, ok := h.resolveAdminOrderItems(c, req.Items, contractPriceListID, country)
	if !ok {
		return
	}

	// C-2: Refuse to create orders without a destination country when items are present.
	// validateOrderItemsCompliance returns true on empty country (no profile match), so
	// without this guard, market-restricted products (halal-only, etc.) bypass the check.
	if country == "" {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "target_country_required")
		return
	}

	// Server-side recompute subtotal from validated unit prices; ignore client-supplied
	// subtotal/total to prevent admin-side price manipulation (M-14).
	subtotal := 0.0
	for _, item := range items {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}

	taxAmount := req.TaxAmount
	if taxAmount < 0 {
		taxAmount = 0
	}
	shippingAmount := req.ShippingAmount
	if shippingAmount < 0 {
		shippingAmount = 0
	}
	total := subtotal + taxAmount + shippingAmount

	// Compliance check (always runs once items are resolved).
	compliance := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), country, products)
	if len(compliance.Violations) > 0 {
		response.ErrorRespDetail(c, http.StatusUnprocessableEntity, "compliance_violation", gin.H{
			"country":    compliance.Country,
			"violations": compliance.Violations,
			"warnings":   compliance.Warnings,
		})
		return
	}

	if !h.validateAdminInventory(c, items, productByID) {
		return
	}

	if !h.checkAdminCompanyCreditLimit(c, userID, total) {
		return
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
		TaxAmount:      taxAmount,
		ShippingAmount: shippingAmount,
		TotalAmount:    total,
		Currency:       currency,
		TrackingNumber: strings.TrimSpace(req.TrackingNumber),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if req.ShippingAddress != nil {
		order.ShippingAddress = *req.ShippingAddress
	}
	if paymentInsufficientForExecution(order.Status, order.PaymentStatus) {
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

	// G20: admin-supplied prices are validated before any field is mutated. A
	// zero/negative line unit price would persist goods at no cost onto the
	// order and its derived invoice, and a non-positive total or negative
	// amount would manufacture a negative invoice. Previously AdminUpdateOrder
	// wrote these values verbatim.
	if reason := validateAdminOrderPrices(req); reason != "" {
		response.ErrorRespDetail(c, http.StatusBadRequest, "invalid_request", gin.H{"reason": reason})
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
	if ps := request.FirstOptionalNonEmpty(req.PaymentStatus, req.PaymentStatusSnake); ps != nil {
		order.PaymentStatus = *ps
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
	if tn := request.FirstOptionalNonEmpty(req.TrackingNumber, req.TrackingNumberSnake); tn != nil {
		order.TrackingNumber = *tn
	}
	if req.ShippingAddress != nil {
		order.ShippingAddress = *req.ShippingAddress
	}
	if !h.validateOrderItemsCompliance(c, order.ShippingAddress.Country, order.Items) {
		return
	}
	currentStatus := strings.ToLower(strings.TrimSpace(order.Status))

	// Validate status transition before any side-effects.
	if currentStatus != previousStatus {
		if err := modelsOrder.ValidateOrderStatusTransition(previousStatus, currentStatus); err != nil {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "invalid_status_transition")
			return
		}
	}

	if paymentInsufficientForExecution(currentStatus, order.PaymentStatus) {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "payment_policy_violation")
		return
	}

	// G20 (follow-up): the admin confirm path must re-enforce the same policy the
	// customer confirm path does — per-product MOQ, the contract price-list
	// minimum, and the buyer company's CUMULATIVE credit exposure — before stock
	// is committed. order.Items / order.TotalAmount / order.UserID here already
	// reflect the request mutations, so the checks run against the final state.
	// Placed before the stock-adjustment branch so an item edit that also
	// confirms cannot bypass them.
	if currentStatus == modelsOrder.OrderStatusConfirmed && previousStatus != currentStatus {
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

	newFin := orderSvc.SnapshotOrderFinancial(order)
	if orderSvc.OrderFinancialChanged(origFin, newFin) && orderSvc.OrderStatusRequiresFinancialReason(previousStatus) {
		if strings.TrimSpace(req.FinancialAdjustmentReason) == "" {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "financial_adjustment_reason_required")
			return
		}
	}

	if previousStatus != "cancelled" && currentStatus == "cancelled" && order.StockReserved {
		if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_stock_release_failed")
			return
		}
	}
	// H-22: warn (audit) when cancelling an order with confirmed payments still
	// outstanding. Mirrors the same guard in AdminUpdateOrderStatus so the audit
	// trail is consistent across both update paths.
	if previousStatus != "cancelled" && currentStatus == "cancelled" && h.services.Payment != nil {
		count, total, perr := h.services.Payment.CountConfirmedPaymentsForOrder(c.Request.Context(), order.ID)
		if perr == nil && count > 0 {
			h.logOrderAudit(c, "order_cancel_with_unrefunded_payments",
				order.ID, c.GetString("userID"),
				fmt.Sprintf("count=%d", count),
				fmt.Sprintf("total=%.2f %s", total, order.Currency),
			)
		}
	}
	if order.StockReserved && req.Items != nil && currentStatus != "cancelled" {
		stockAdjustment := calculateStockAdjustment(originalItems, order.Items)
		if hasNonZeroStockDelta(stockAdjustment) {
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
	}

	now := time.Now()
	order.UpdatedAt = now
	if currentStatus == modelsOrder.OrderStatusConfirmed && previousStatus != currentStatus && h.services.Product != nil {
		order.COGS = orderSvc.ComputeOrderCOGS(c.Request.Context(), order.Items, h.services.Product)
	}
	if order.Status == "confirmed" && order.ConfirmedAt == nil {
		order.ConfirmedAt = &now
	}
	if order.Status == "shipped" && order.ShippedAt == nil {
		order.ShippedAt = &now
	}
	if order.Status == "delivered" && order.DeliveredAt == nil {
		order.DeliveredAt = &now
	}

	// Route through UpdateOrderForAdmin to trigger outbox when transitioning to confirmed
	if err := h.services.Order.UpdateOrderForAdmin(c.Request.Context(), order, previousStatus, currentStatus); err != nil {
		if errors.Is(err, modelsOrder.ErrInsufficientStock) {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "inventory_violation")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "order_update_failed")
		return
	}

	h.syncOrderFinancialSideEffects(c, order, origFin, newFin, previousStatus, req.FinancialAdjustmentReason)
	if previousStatus != currentStatus {
		h.logOrderAudit(c, "order_status_change", order.ID, c.GetString("userID"), previousStatus, currentStatus)
	}
	if previousStatus != currentStatus && order.User != nil {
		h.services.Order.SendOrderStatusEmail(order, order.User.Email,
			order.User.FirstName+" "+order.User.LastName, currentStatus)
	}
	h.dispatchOrderWebhook(c, order, previousStatus, currentStatus)
	c.JSON(http.StatusOK, order)
}

// dispatchOrderWebhook fires outgoing webhooks for order status transitions.
func (h *Handler) dispatchOrderWebhook(c *gin.Context, order *modelsOrder.Order, prevStatus, newStatus string) {
	if h.services == nil || prevStatus == newStatus {
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
	h.emitLifecycleEvent(c, eventType, order.ID, order)
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

func hasNonZeroStockDelta(deltas map[string]int) bool {
	for _, delta := range deltas {
		if delta != 0 {
			return true
		}
	}
	return false
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

// validateAdminOrderPrices rejects client-supplied order prices that are never
// legitimate on an admin edit: a negative line unit price or a negative
// subtotal/tax/shipping/total (would manufacture a negative invoice on the order
// and its derived invoice). Returns a human-readable reason, empty when valid.
// G20: previously AdminUpdateOrder wrote these values verbatim, so a negative
// TotalAmount or a negative-price line persisted untouched.
//
// Zero values are deliberately ALLOWED. Bulk / requisition / reorder drafts are
// created unpriced by design — every line UnitPrice, Subtotal and TotalAmount are
// 0 — and the admin order-edit modal echoes those zero values back on the next
// save. Rejecting them would deadlock the draft workflow: the customer cannot
// confirm an empty-country draft (target_country_required) and the admin cannot
// edit that same draft to add a shipping country (the guard above would 400 on
// its own zero prices). Draft prices are never authoritative anyway: the customer
// confirm path re-prices every line server-side (CustomerConfirmOrder →
// repriceOrderItems), and admin-created orders are server-priced in
// AdminCreateOrder. Placeholder lines with quantity < 1 are ignored — they carry
// no sellable value and are rejected by the quantity guards elsewhere.
func validateAdminOrderPrices(req adminUpdateOrderRequest) string {
	if req.Items != nil {
		for i, it := range *req.Items {
			if it.Quantity < 1 {
				continue
			}
			if it.UnitPrice < 0 {
				return fmt.Sprintf("items[%d].unitPrice must not be negative", i)
			}
		}
	}
	if req.Subtotal != nil && *req.Subtotal < 0 {
		return "subtotal must not be negative"
	}
	if req.TaxAmount != nil && *req.TaxAmount < 0 {
		return "taxAmount must not be negative"
	}
	if req.ShippingAmount != nil && *req.ShippingAmount < 0 {
		return "shippingAmount must not be negative"
	}
	if req.TotalAmount != nil && *req.TotalAmount < 0 {
		return "totalAmount must not be negative"
	}
	return ""
}
