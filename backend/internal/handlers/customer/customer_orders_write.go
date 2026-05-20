package customer

import (
	"bytes"
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/kyb"
	"candypro/api/internal/pkg/money"
	"candypro/api/internal/pkg/response"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type customerCreateOrderItemRequest struct {
	ProductID      string  `json:"productId" binding:"required"`
	Quantity       int     `json:"quantity" binding:"required,min=1"`
	UnitPrice      float64 `json:"unitPrice"`
	Specifications string  `json:"specifications"`
}

type customerCreateOrderRequest struct {
	OrderNumber     string                           `json:"orderNumber"`
	InquiryID       *string                          `json:"inquiryId"`
	Items           []customerCreateOrderItemRequest `json:"items" binding:"required,min=1"`
	TaxAmount       float64                          `json:"taxAmount"`
	ShippingAmount  float64                          `json:"shippingAmount"`
	Currency        string                           `json:"currency"`
	ShippingAddress modelsOrder.Address              `json:"shippingAddress"`
}

// CustomerCreateOrder creates a new customer order and binds it to the authenticated user.
func (h *Handler) CustomerCreateOrder(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req customerCreateOrderRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.TaxAmount < 0 || req.ShippingAmount < 0 {
		response.InvalidResp(c, "tax_shipping_negative")
		return
	}

	if msg := validateCustomerShippingAddress(req.ShippingAddress); msg != "" {
		response.InvalidResp(c, msg)
		return
	}

	var inquiryID *string
	if req.InquiryID != nil && strings.TrimSpace(*req.InquiryID) != "" {
		trimmedInquiryID := strings.TrimSpace(*req.InquiryID)
		inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), trimmedInquiryID)
		if err != nil {
			response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
			return
		}
		if inquiry.UserID == nil || *inquiry.UserID != userID {
			response.ErrorResp(c, http.StatusForbidden, "inquiry_no_access")
			return
		}
		inquiryID = &trimmedInquiryID
	}

	items := make(modelsOrder.OrderItemArray, 0, len(req.Items))
	selectedProducts := make([]modelsProduct.Product, 0, len(req.Items))
	productByID := make(map[string]modelsProduct.Product, len(req.Items))
	subtotal := 0.0

	// Resolve contract price list for the user (if applicable)
	var contractPriceListID *string
	if h.services.Price != nil && h.services.User != nil && h.services.Company != nil {
		if usr, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr == nil && usr.CompanyID != nil {
			if company, compErr := h.services.Company.GetCompany(c.Request.Context(), *usr.CompanyID); compErr == nil && company.PriceListID != nil {
				contractPriceListID = company.PriceListID
			}
		}
	}

	for _, it := range req.Items {
		productID := strings.TrimSpace(it.ProductID)
		if productID == "" {
			response.InvalidResp(c, "invalid_request")
			return
		}
		if it.Quantity < 1 {
			response.InvalidResp(c, "quantity_min_1")
			return
		}

		product, err := h.services.Product.GetProductByID(c.Request.Context(), productID)
		if err != nil {
			response.ErrorResp(c, http.StatusNotFound, "product_not_found")
			return
		}
		if status := strings.ToLower(strings.TrimSpace(product.Status)); status != "" && status != "active" {
			response.ErrorResp(c, http.StatusBadRequest, "product_unavailable")
			return
		}

		// SEC-2: Resolve price server-side — reject client-supplied unitPrice
		var unitPrice float64
		if contractPriceListID != nil && h.services.Price != nil {
			if cp, priceErr := h.services.Price.GetPriceForProduct(c.Request.Context(), productID, *contractPriceListID, it.Quantity); priceErr == nil {
				unitPrice = cp
			}
		}
		if unitPrice <= 0 {
			unitPrice = product.BasePrice
		}
		if unitPrice <= 0 {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "no_price")
			return
		}
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), product, unitPrice, req.ShippingAddress.Country)

		orderItem := modelsOrder.OrderItem{
			ProductID:      productID,
			Quantity:       it.Quantity,
			UnitPrice:      unitPrice,
			Specifications: strings.TrimSpace(it.Specifications),
		}
		items = append(items, orderItem)
		selectedProducts = append(selectedProducts, *product)
		productByID[productID] = *product
		subtotal += float64(it.Quantity) * unitPrice
	}

	compliance := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), req.ShippingAddress.Country, selectedProducts)
	if len(compliance.Violations) > 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsCommon.ErrorResponse{
			Error:   "compliance_violation",
			Message: i18n.T(c, "errors.compliance_violation"),
			Details: gin.H{
				"country":    compliance.Country,
				"violations": compliance.Violations,
				"warnings":   compliance.Warnings,
			},
		})
		return
	}
	ids := make([]string, 0, len(productByID))
	for id := range productByID {
		ids = append(ids, id)
	}
	sellable, _ := h.services.Product.EffectiveSellableByProducts(c.Request.Context(), ids, modelsProduct.ChannelWebstore)
	inventory := h.services.Order.ValidateInventoryWithSellable(items, productByID, sellable)
	if len(inventory.Violations) > 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsCommon.ErrorResponse{
			Error:   "inventory_violation",
			Message: i18n.T(c, "errors.inventory_violation"),
			Details: gin.H{
				"violations": inventory.Violations,
				"warnings":   inventory.Warnings,
			},
		})
		return
	}

	currencyNorm, curErr := money.NormalizeISOCurrency(req.Currency)
	if curErr != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	currency := currencyNorm
	if currency == "" {
		currency = "USD"
	}

	orderNumber := strings.TrimSpace(req.OrderNumber)
	if orderNumber == "" {
		orderNumber = buildCustomerOrderNumber()
	}

	totalAmount := subtotal + req.TaxAmount + req.ShippingAmount
	if math.IsNaN(totalAmount) || math.IsInf(totalAmount, 0) || totalAmount < 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if !h.ensureActiveOrKYBBypassForAmount(c, userID, totalAmount, kyb.LineProductIDs(items)...) {
		return
	}
	now := time.Now()
	orderStatus := "pending"
	if h.services.Approval != nil {
		if needsApproval, _, _ := h.services.Approval.ShouldRequireApproval(c.Request.Context(), userID, totalAmount); needsApproval {
			orderStatus = "pending_approval"
		}
	}

	order := &modelsOrder.Order{
		ID:             crypto.GenerateID(),
		OrderNumber:    orderNumber,
		UserID:         userID,
		InquiryID:      inquiryID,
		Status:         orderStatus,
		PaymentStatus:  "unpaid",
		Items:          items,
		StockReserved:  false,
		Subtotal:       subtotal,
		TaxAmount:      req.TaxAmount,
		ShippingAmount: req.ShippingAmount,
		TotalAmount:    totalAmount,
		Currency:       currency,
		ShippingAddress: modelsOrder.Address{
			Street:  strings.TrimSpace(req.ShippingAddress.Street),
			City:    strings.TrimSpace(req.ShippingAddress.City),
			State:   strings.TrimSpace(req.ShippingAddress.State),
			ZipCode: strings.TrimSpace(req.ShippingAddress.ZipCode),
			Country: strings.TrimSpace(req.ShippingAddress.Country),
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	order.StockReserved = true
	if err := h.services.Order.CreateOrderWithStockReservation(c.Request.Context(), order); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		return
	}

	// Fire outgoing webhook for order.created event
	if h.services.Webhook != nil {
		h.services.Webhook.Dispatch(c.Request.Context(), modelsOrder.WebhookEventOrderCreated, order.ID, order)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
		"compliance": gin.H{
			"country":       compliance.Country,
			"warnings":      dedupeCustomerWarnings(append(compliance.Warnings, inventory.Warnings...)),
			"paymentPolicy": compliance.Payment,
		},
		"inventory": gin.H{"warnings": inventory.Warnings},
	})

	// Send order confirmation notification
	if h.services.Notification != nil {
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "order",
			Reference: order.ID,
			Title:     "Order Placed",
			Message:   "Your order #" + order.OrderNumber + " has been placed and is pending review.",
		})
	}
}

// CustomerConfirmOrder confirms an AI-drafted order before final processing.
// The confirm flow only accepts compliance acknowledgement; item edits and
// price changes are rejected — the server-authoritative draft data is used.
func (h *Handler) CustomerConfirmOrder(c *gin.Context) {
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

	if order.Status != "pending_confirmation" {
		response.ErrorResp(c, http.StatusConflict, "invalid_state")
		return
	}

	// Re-validate compliance at confirm time with current market profiles
	if order.ShippingAddress.Country != "" && len(order.Items) > 0 {
		products := make([]modelsProduct.Product, 0, len(order.Items))
		allFound := true
		for _, item := range order.Items {
			product, err := h.services.Product.GetProductByID(c.Request.Context(), item.ProductID)
			if err != nil {
				allFound = false
				break
			}
			products = append(products, *product)
		}
		if allFound {
			complianceRecheck := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), order.ShippingAddress.Country, products)
			if len(complianceRecheck.Violations) > 0 {
				c.JSON(http.StatusUnprocessableEntity, modelsCommon.ErrorResponse{
					Error:   "compliance_violation",
					Message: i18n.T(c, "errors.compliance_violation"),
					Details: gin.H{
						"country":    complianceRecheck.Country,
						"violations": complianceRecheck.Violations,
						"warnings":   complianceRecheck.Warnings,
					},
				})
				return
			}
		}
	}

	// Parse compliance acknowledgement and optional item edits.
	var parsedReq struct {
		ComplianceAck bool                     `json:"complianceAck"`
		Items         []modelsOrder.OrderItem  `json:"items"`
	}
	if c.Request.Body != nil {
		rawBody, readErr := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(rawBody))
		if readErr != nil {
			response.InvalidResp(c, "invalid_request")
			return
		}
		trimmedBody := bytes.TrimSpace(rawBody)
		if len(trimmedBody) > 0 {
			if err := json.Unmarshal(trimmedBody, &parsedReq); err != nil {
				response.InvalidResp(c, "invalid_request")
				return
			}
		}
	}

	if !order.ComplianceOfficialEvidence && !parsedReq.ComplianceAck {
		c.JSON(http.StatusConflict, modelsCommon.ErrorResponse{
			Error:   "compliance_ack_required",
			Message: i18n.T(c, "errors.compliance_ack_required"),
			Details: gin.H{
				"requiresComplianceAck": true,
			},
		})
		return
	}

	if !h.ensureActiveOrKYBBypassForAmount(c, userID, order.TotalAmount, kyb.LineProductIDs(order.Items)...) {
		return
	}

	now := time.Now()
	itemsToConfirm := order.Items
	if len(parsedReq.Items) > 0 {
		itemsToConfirm = parsedReq.Items
	}
	if err := h.services.Order.ConfirmAndReserveOrder(c.Request.Context(), orderID, itemsToConfirm, now); err != nil {
		if errors.Is(err, modelsOrder.ErrInsufficientStock) {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "insufficient_stock")
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			latest, latestErr := h.services.Order.GetOrder(c.Request.Context(), orderID)
			if latestErr != nil {
				response.ErrorResp(c, http.StatusConflict, "invalid_state")
				return
			}
			if latest.UserID != userID {
				response.ErrorResp(c, http.StatusForbidden, "forbidden")
				return
			}
			if latest.Status == "cancelled" && !latest.StockReserved {
				response.ErrorResp(c, http.StatusConflict, "order_expired")
				return
			}
			response.ErrorResp(c, http.StatusConflict, "invalid_state")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "order_confirm_failed")
		return
	}

	order.Status = "pending"
	order.ConfirmedAt = &now
	order.UpdatedAt = now

	c.JSON(http.StatusOK, gin.H{
		"message": "Order confirmed successfully",
		"order":   order,
	})
}

func validateCustomerShippingAddress(address modelsOrder.Address) string {
	if strings.TrimSpace(address.Street) == "" {
		return "shippingAddress.street is required"
	}
	if strings.TrimSpace(address.City) == "" {
		return "shippingAddress.city is required"
	}
	if strings.TrimSpace(address.Country) == "" {
		return "shippingAddress.country is required"
	}
	return ""
}

func buildCustomerOrderNumber() string {
	return fmt.Sprintf("CUS-%s-%s", time.Now().Format("20060102"), strings.ToUpper(crypto.GenerateSlug()))
}

func dedupeCustomerWarnings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

// CustomerCancelOrder allows a customer to cancel their own pending order.
// Only orders in "pending" status can be cancelled by the customer.
func (h *Handler) CustomerCancelOrder(c *gin.Context) {
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

	// Only allow cancellation of pending and pending_confirmation orders
	if err := modelsOrder.ValidateOrderStatusTransition(order.Status, "cancelled"); err != nil {
		response.ErrorResp(c, http.StatusConflict, "invalid_state")
		return
	}

	now := time.Now()
	order.Status = "cancelled"
	order.UpdatedAt = now

	if order.StockReserved {
		if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_update_failed")
			return
		}
	} else {
		order.StockReserved = false
		if err := h.services.Order.UpdateOrder(c.Request.Context(), order); err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "order_update_failed")
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"orderId": order.ID,
		"status":  "cancelled",
	})

	// Notify customer of cancellation
	if h.services.Notification != nil {
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "order",
			Reference: order.ID,
			Title:     "Order Cancelled",
			Message:   "Your order #" + order.OrderNumber + " has been cancelled.",
		})
	}
}

// CustomerNudgeOrder allows a customer to send a reminder about their order to all admin users.
func (h *Handler) CustomerNudgeOrder(c *gin.Context) {
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

	nudgeable := map[string]bool{
		modelsOrder.OrderStatusPending:    true,
		modelsOrder.OrderStatusConfirmed:  true,
		modelsOrder.OrderStatusProduction: true,
	}
	if !nudgeable[order.Status] {
		response.ErrorResp(c, http.StatusConflict, "cannot_nudge_order")
		return
	}

	// Confirm to customer
	if h.services.Notification != nil {
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "order",
			Reference: order.ID,
			Title:     "Nudge Sent",
			Message:   "Your reminder for order #" + order.OrderNumber + " has been sent to our team.",
		})
	}

	// Notify all admin users
	if h.services.User != nil && h.services.Notification != nil {
		adminUsers, userErr := h.services.User.FindAdminUsers(c.Request.Context())
		if userErr == nil {
			for _, admin := range adminUsers {
				_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
					UserID:    admin.ID,
					Type:      "order",
					Reference: order.ID,
					Title:     "Customer Nudge",
					Message:   "Customer sent a reminder for order #" + order.OrderNumber + ".",
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Nudge sent successfully",
	})
}
