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
	orderRepo "candypro/api/internal/repository/order"
	orderSvc "candypro/api/internal/services/order"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
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
	OrderNumber       string                           `json:"orderNumber"`
	InquiryID         *string                          `json:"inquiryId"`
	Items             []customerCreateOrderItemRequest `json:"items" binding:"required,min=1"`
	TaxAmount         float64                          `json:"taxAmount"`
	ShippingAmount    float64                          `json:"shippingAmount"`
	Currency          string                           `json:"currency"`
	Incoterms         string                           `json:"incoterms"`
	EstimatedWeightKg float64                          `json:"estimatedWeightKg"`
	ShippingAddress   modelsOrder.Address              `json:"shippingAddress"`
}

// CustomerCreateOrder creates a new customer order and binds it to the authenticated user.
// @Summary Create customer order
// @Tags customer-orders
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Router /user/orders [post]
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

	// R2 E-5: batch product lookups up front. Previously every requested line
	// triggered an individual GetProductByID call — N round-trips on the
	// primary customer checkout path. We still validate per-line below, but
	// the resolves all come from a single FindByIDs query.
	uniqueIDs := make([]string, 0, len(req.Items))
	idSeen := make(map[string]struct{}, len(req.Items))
	for _, it := range req.Items {
		pid := strings.TrimSpace(it.ProductID)
		if pid == "" {
			continue
		}
		if _, ok := idSeen[pid]; ok {
			continue
		}
		idSeen[pid] = struct{}{}
		uniqueIDs = append(uniqueIDs, pid)
	}
	productLookup := make(map[string]modelsProduct.Product, len(uniqueIDs))
	if len(uniqueIDs) > 0 {
		batch, batchErr := h.services.Product.GetProductsByIDs(c.Request.Context(), uniqueIDs)
		if batchErr != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "product_lookup_failed")
			return
		}
		for i := range batch {
			productLookup[batch[i].ID] = batch[i]
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
		if !h.validateLineMinQuantity(c, productID, it.Quantity, contractPriceListID) {
			return
		}

		product, found := productLookup[productID]
		if !found {
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
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), &product, unitPrice, req.ShippingAddress.Country)
		unitPrice = h.applyChannelUnitPrice(c, unitPrice)

		orderItem := modelsOrder.OrderItem{
			ProductID:      productID,
			Quantity:       it.Quantity,
			UnitPrice:      unitPrice,
			Specifications: strings.TrimSpace(it.Specifications),
		}
		items = append(items, orderItem)
		selectedProducts = append(selectedProducts, product)
		productByID[productID] = product
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
				"warnings":   formatInventoryWarnings(c, inventory.Warnings, productByID),
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

	pricing := h.computeCheckoutPricing(c.Request.Context(), checkoutPricingInput{
		Items:             items,
		ProductByID:       productByID,
		ShippingAddress:   req.ShippingAddress,
		Incoterms:         req.Incoterms,
		EstimatedWeightKg: req.EstimatedWeightKg,
		Subtotal:          subtotal,
		Currency:          currency,
	})
	taxAmount := pricing.TaxAmount
	shippingAmount := pricing.ShippingAmount
	if pricing.Currency != "" {
		currency = pricing.Currency
	}

	var webstoreCh *modelsProduct.Channel
	if h.services.Channel != nil {
		webstoreCh, _ = h.services.Channel.ResolveWebstoreChannel(c.Request.Context())
		if taxAmount <= 0 {
			taxAmount = h.services.Channel.ComputeTaxAmount(subtotal, webstoreCh, taxAmount)
		}
	}

	taxAmount = money.RoundMoney(taxAmount)
	shippingAmount = money.RoundMoney(shippingAmount)

	totalAmount := money.RoundMoney(subtotal + taxAmount + shippingAmount)
	if math.IsNaN(totalAmount) || math.IsInf(totalAmount, 0) || totalAmount < 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if !h.checkCompanyCreditLimit(c, userID, totalAmount) {
		return
	}
	if !h.ensureActiveOrKYBBypassForAmount(c, userID, totalAmount, currency, kyb.LineProductIDs(items)...) {
		return
	}

	now := time.Now()
	needsApproval := false
	if h.services.Approval != nil {
		var err error
		needsApproval, _, err = h.services.Approval.ShouldRequireApproval(c.Request.Context(), userID, totalAmount)
		if err != nil {
			needsApproval = false
		}
	}
	orderStatus := modelsOrder.OrderStatusPendingConfirm
	if h.services.Channel != nil {
		orderStatus = h.services.Channel.ResolveInitialOrderStatus(webstoreCh, needsApproval)
	} else if needsApproval {
		orderStatus = modelsOrder.OrderStatusPendingApproval
	}

	paymentStatus := "unpaid"
	if h.services.Channel != nil {
		paymentStatus = h.services.Channel.ResolvePaymentStatus(webstoreCh)
	}

	hasOfficialEvidence := len(compliance.Violations) == 0 && len(compliance.Warnings) == 0

	order := &modelsOrder.Order{
		ID:             crypto.GenerateID(),
		OrderNumber:    orderNumber,
		UserID:         userID,
		InquiryID:      inquiryID,
		Source:         modelsOrder.OrderSourceCart,
		Status:         orderStatus,
		PaymentStatus:  paymentStatus,
		Items:          items,
		StockReserved:  false,
		ComplianceOfficialEvidence: hasOfficialEvidence,
		Subtotal:       subtotal,
		TaxAmount:      taxAmount,
		ShippingAmount: shippingAmount,
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
	h.assignOrderWarehouseID(c, &order.WarehouseID)

	if err := h.services.Order.CreateOrder(c.Request.Context(), order); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		return
	}

	if orderStatus == modelsOrder.OrderStatusPendingApproval {
		h.notifyOrderApprovers(c, order)
	}

	// Fire lifecycle event for order.created
	h.emitLifecycleEvent(c, modelsOrder.WebhookEventOrderCreated, order.ID, order)

	inventoryWarnings := formatInventoryWarnings(c, inventory.Warnings, productByID)

	c.JSON(http.StatusCreated, gin.H{
		"message": i18n.T(c, "messages.order_created_success"),
		"order":   order,
		"compliance": gin.H{
			"country":       compliance.Country,
			"warnings":      dedupeCustomerWarnings(append(compliance.Warnings, inventoryWarnings...)),
			"paymentPolicy": compliance.Payment,
		},
		"inventory": gin.H{"warnings": inventoryWarnings},
	})

	// Send order confirmation notification
	if h.services.Notification != nil {
		vars := map[string]string{"orderNumber": order.OrderNumber}
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "order",
			Reference: order.ID,
			Title:     i18n.TWithVars(c, "notifications.order_placed_title", vars),
			Message:   i18n.TWithVars(c, "notifications.order_placed_message", vars),
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

	// Re-validate compliance at confirm time with current market profiles.
	// H-2/H-3: batch product lookups in a single query rather than one per line.
	productByID := make(map[string]modelsProduct.Product, len(order.Items))
	if len(order.Items) > 0 {
		ids := make([]string, 0, len(order.Items))
		seen := make(map[string]struct{}, len(order.Items))
		for _, item := range order.Items {
			pid := strings.TrimSpace(item.ProductID)
			if pid == "" {
				continue
			}
			if _, ok := seen[pid]; ok {
				continue
			}
			seen[pid] = struct{}{}
			ids = append(ids, pid)
		}
		batch, batchErr := h.services.Product.GetProductsByIDs(c.Request.Context(), ids)
		if batchErr == nil {
			for i := range batch {
				productByID[batch[i].ID] = batch[i]
			}
		}
	}

	if order.ShippingAddress.Country != "" && len(order.Items) > 0 {
		products := make([]modelsProduct.Product, 0, len(productByID))
		allFound := true
		for _, item := range order.Items {
			p, ok := productByID[item.ProductID]
			if !ok {
				allFound = false
				break
			}
			products = append(products, p)
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

	now := time.Now()
	itemsToConfirm := order.Items
	if len(parsedReq.Items) > 0 {
		itemsToConfirm = parsedReq.Items
		// M-19: customer-supplied items may override the server-authoritative
		// draft. Quantities/specifications are honored, but unit prices are
		// ALWAYS re-priced server-side below (H2) — a client-supplied unitPrice
		// can never reach the persisted order. Each diff is still recorded so
		// reconciliation tooling can detect post-quote line tampering.
		if itemsDifferFromOrder(order.Items, parsedReq.Items) {
			slog.Warn("customer confirmed order with overridden items",
				"orderID", orderID,
				"userID", userID,
				"originalCount", len(order.Items),
				"overrideCount", len(parsedReq.Items),
				"originalItems", summarizeOrderItems(order.Items),
				"overrideItems", summarizeOrderItems(parsedReq.Items),
			)
		}
	}

	// H1/H2: re-price every line server-side from the catalog / contract price
	// list. Bulk, requisition and reorder drafts are created with UnitPrice=0 —
	// trusting order.Subtotal at confirm would ship goods at ~zero cost. Confirm
	// overrides are honored for quantity but their unit prices are discarded.
	repricedItems, repricedProducts, repricedSubtotal, priceErr := h.repriceOrderItems(c, userID, itemsToConfirm, order.ShippingAddress.Country)
	if priceErr != nil {
		switch {
		case errors.Is(priceErr, errRepriceNoPrice):
			response.ErrorResp(c, http.StatusUnprocessableEntity, "no_price")
		case errors.Is(priceErr, errRepriceProductNotFound):
			response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		default:
			response.ErrorResp(c, http.StatusInternalServerError, "order_confirm_failed")
		}
		return
	}
	itemsToConfirm = repricedItems
	productByID = repricedProducts

	// 确认时重新计算税/运费（地址或费率可能已变化），基于重定价后的行与 subtotal。
	confirmPricing := h.computeCheckoutPricing(c.Request.Context(), checkoutPricingInput{
		Items:           itemsToConfirm,
		ProductByID:     productByID,
		ShippingAddress: order.ShippingAddress,
		Incoterms:       "FOB",
		Subtotal:        repricedSubtotal,
		Currency:        order.Currency,
	})
	if h.services.Channel != nil && confirmPricing.TaxAmount <= 0 {
		webstoreCh, _ := h.services.Channel.ResolveWebstoreChannel(c.Request.Context())
		confirmPricing.TaxAmount = h.services.Channel.ComputeTaxAmount(repricedSubtotal, webstoreCh, confirmPricing.TaxAmount)
	}
	confirmTotal := repricedSubtotal + confirmPricing.TaxAmount + confirmPricing.ShippingAmount
	if !h.ensureActiveOrKYBBypassForAmount(c, userID, confirmTotal, order.Currency, kyb.LineProductIDs(itemsToConfirm)...) {
		return
	}

	// H-15: compute financials BEFORE the atomic confirm so they land in the
	// same transaction as the status flip and stock reservation. Previously the
	// confirm tx and the financial update were separate, leaving stock reserved
	// with cogs=0 if the second update failed.
	taxAmount := money.RoundMoney(confirmPricing.TaxAmount)
	shippingAmount := money.RoundMoney(confirmPricing.ShippingAmount)
	subtotal := money.RoundMoney(repricedSubtotal)
	totalAmount := money.RoundMoney(subtotal + taxAmount + shippingAmount)
	currency := order.Currency
	if confirmPricing.Currency != "" {
		currency = confirmPricing.Currency
	}
	cogs := 0.0
	if h.services.Product != nil {
		cogs = orderSvc.ComputeOrderCOGS(c.Request.Context(), itemsToConfirm, h.services.Product)
	}
	itemsArray := modelsOrder.OrderItemArray(itemsToConfirm)
	confirmFin := &orderRepo.OrderConfirmFinancials{
		Items:          &itemsArray,
		COGS:           &cogs,
		Subtotal:       &subtotal,
		TaxAmount:      &taxAmount,
		ShippingAmount: &shippingAmount,
		TotalAmount:    &totalAmount,
		Currency:       &currency,
	}
	if err := h.services.Order.ConfirmAndReserveOrderWithFinancials(c.Request.Context(), orderID, itemsToConfirm, now, confirmFin); err != nil {
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

	// Reload the order so the response reflects the just-committed financials.
	order, err = h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_confirm_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(c, "messages.order_confirmed_success"),
		"order":   order,
	})
}

func validateCustomerShippingAddress(address modelsOrder.Address) string {
	if strings.TrimSpace(address.Street) == "" {
		return "shipping_street_required"
	}
	if strings.TrimSpace(address.City) == "" {
		return "shipping_city_required"
	}
	if strings.TrimSpace(address.Country) == "" {
		return "target_country_required"
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

// itemsDifferFromOrder reports whether the customer-supplied confirm payload
// describes a different line set than the server-authoritative draft. Used to
// gate the M-19 audit log so we only record real overrides (not echoed identical
// payloads).
func itemsDifferFromOrder(serverItems modelsOrder.OrderItemArray, clientItems []modelsOrder.OrderItem) bool {
	if len(serverItems) != len(clientItems) {
		return true
	}
	type key struct {
		id    string
		qty   int
		price float64
	}
	count := func(items []modelsOrder.OrderItem) map[key]int {
		out := make(map[key]int, len(items))
		for _, it := range items {
			out[key{id: it.ProductID, qty: it.Quantity, price: it.UnitPrice}]++
		}
		return out
	}
	serverCounts := count(serverItems)
	clientCounts := count(clientItems)
	if len(serverCounts) != len(clientCounts) {
		return true
	}
	for k, v := range serverCounts {
		if clientCounts[k] != v {
			return true
		}
	}
	return false
}

// summarizeOrderItems renders an item slice as a compact "<productID>:<qty>@<price>"
// list suitable for audit log fields without dumping full JSON.
func summarizeOrderItems(items []modelsOrder.OrderItem) string {
	parts := make([]string, 0, len(items))
	for _, it := range items {
		parts = append(parts, fmt.Sprintf("%s:%d@%.2f", it.ProductID, it.Quantity, it.UnitPrice))
	}
	return strings.Join(parts, ",")
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
		"message": i18n.T(c, "messages.order_cancelled_success"),
		"orderId": order.ID,
		"status":  "cancelled",
	})

	// Notify customer of cancellation
	if h.services.Notification != nil {
		vars := map[string]string{"orderNumber": order.OrderNumber}
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "order",
			Reference: order.ID,
			Title:     i18n.TWithVars(c, "notifications.order_cancelled_title", vars),
			Message:   i18n.TWithVars(c, "notifications.order_cancelled_message", vars),
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
		vars := map[string]string{"orderNumber": order.OrderNumber}
		_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "order",
			Reference: order.ID,
			Title:     i18n.TWithVars(c, "notifications.customer_nudge_title", vars),
			Message:   i18n.TWithVars(c, "notifications.customer_nudge_message", vars),
		})
	}

	// Notify all admin users
	if h.services.User != nil && h.services.Notification != nil {
		adminUsers, userErr := h.services.User.FindAdminUsers(c.Request.Context())
		if userErr == nil {
			vars := map[string]string{"orderNumber": order.OrderNumber}
			for _, admin := range adminUsers {
				_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
					UserID:    admin.ID,
					Type:      "order",
					Reference: order.ID,
					Title:     i18n.TWithVars(c, "notifications.admin_customer_nudge_title", vars),
					Message:   i18n.TWithVars(c, "notifications.admin_customer_nudge_message", vars),
				})
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": i18n.T(c, "messages.nudge_sent_success"),
	})
}
