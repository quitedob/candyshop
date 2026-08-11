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

	pricing, perr := h.computeCheckoutPricing(c.Request.Context(), checkoutPricingInput{
		Items:             items,
		ProductByID:       productByID,
		ShippingAddress:   req.ShippingAddress,
		Incoterms:         req.Incoterms,
		EstimatedWeightKg: req.EstimatedWeightKg,
		Subtotal:          subtotal,
		Currency:          currency,
	})
	if perr != nil {
		// G24c: a tax/shipping rate-lookup failure fails CLOSED — never book the
		// order with 0 tax/shipping on a transient DB failure.
		response.ErrorResp(c, http.StatusInternalServerError, "checkout_pricing_failed")
		return
	}
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
		ID:                         crypto.GenerateID(),
		OrderNumber:                orderNumber,
		UserID:                     userID,
		InquiryID:                  inquiryID,
		Source:                     modelsOrder.OrderSourceCart,
		Status:                     orderStatus,
		PaymentStatus:              paymentStatus,
		Items:                      items,
		StockReserved:              false,
		ComplianceOfficialEvidence: hasOfficialEvidence,
		Subtotal:                   subtotal,
		TaxAmount:                  taxAmount,
		ShippingAmount:             shippingAmount,
		TotalAmount:                totalAmount,
		Currency:                   currency,
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

	// G20: never skip the destination-market compliance recheck at confirm.
	// Previously the recheck was gated on a non-empty shipping country, so
	// bulk/requisition/reorder drafts (created without an address) sailed past
	// destination-market validation. A draft with line items that carries no
	// official compliance evidence must declare a shipping country at confirm —
	// without a destination we cannot evaluate market rules (halal-only bans,
	// ingredient restrictions, prepayment policy). Orders with official
	// compliance evidence were already validated against a declared destination
	// at intake, so the recheck below is supplementary and tolerates a missing
	// country (it degrades to generic labeling warnings, never a false pass on
	// market-restricted products that were validated earlier).
	if len(order.Items) > 0 && !order.ComplianceOfficialEvidence && strings.TrimSpace(order.ShippingAddress.Country) == "" {
		c.JSON(http.StatusUnprocessableEntity, modelsCommon.ErrorResponse{
			Error:   "target_country_required",
			Message: i18n.T(c, "errors.target_country_required"),
			Details: gin.H{
				"reason": "a shipping country is required to validate destination-market compliance at confirm",
			},
		})
		return
	}

	if len(order.Items) > 0 {
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
		ComplianceAck bool                    `json:"complianceAck"`
		Items         []modelsOrder.OrderItem `json:"items"`
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

	// M-19 override detection (G20 r3): the customer may supply an items array at
	// confirm that changes line quantities from the server-authoritative draft.
	// An overridden quantity is a NEW quantity that never passed intake
	// validation, so it must be re-checked against product MOQ and the contract
	// price-list minimum below. Unchanged lines keep the validation (or the
	// negotiation, for H10 inquiry drafts) they received at intake. A product
	// absent from the draft (a client-added line) always counts as overridden —
	// draft quantity 0 != override quantity.
	overriddenQty := make(map[string]bool)
	if len(parsedReq.Items) > 0 {
		draftQty := make(map[string]int, len(order.Items))
		for _, it := range order.Items {
			pid := strings.TrimSpace(it.ProductID)
			if pid == "" {
				continue
			}
			draftQty[pid] += it.Quantity
		}
		for _, it := range parsedReq.Items {
			pid := strings.TrimSpace(it.ProductID)
			if pid == "" {
				continue
			}
			if draftQty[pid] != it.Quantity {
				overriddenQty[pid] = true
			}
		}
	}

	// H1/H2: re-price every line server-side from the catalog / contract price
	// list. Bulk, requisition and reorder drafts are created with UnitPrice=0 —
	// trusting order.Subtotal at confirm would ship goods at ~zero cost. Confirm
	// overrides are honored for quantity but their unit prices are discarded.
	//
	// H10: inquiry drafts are already server-priced at intake — an accepted
	// negotiation offer's unit price, or the catalog price for a no-offer
	// conversion (and OEM project conversions resolve their own unit price). The
	// catalog / contract price list must NOT override these agreed prices at
	// confirm, or the negotiated total is lost and a BasePrice=0 OEM-only offer
	// would 422 (no_price) on a validly-priced order. Inquiry drafts keep their
	// prices (client-supplied confirm prices are still discarded).
	var repricedItems []modelsOrder.OrderItem
	var repricedProducts map[string]modelsProduct.Product
	var repricedSubtotal float64
	var priceErr error
	if order.Source == modelsOrder.OrderSourceInquiry {
		repricedItems, repricedProducts, repricedSubtotal, priceErr = h.priceInquiryConfirmItems(c, userID, order.Items, itemsToConfirm, order.ShippingAddress.Country)
	} else {
		repricedItems, repricedProducts, repricedSubtotal, priceErr = h.repriceOrderItems(c, userID, itemsToConfirm, order.ShippingAddress.Country)
	}
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

	// G20: re-enforce per-product MOQ at confirm. Draft paths (bulk CSV,
	// requisition, reorder) never validated MOQ, and a confirm-time M-19
	// quantity override can reduce a line below MOQ. Enforce before stock is
	// committed so a below-MOQ order cannot reserve inventory.
	//
	// G20 r3: inquiry (H10) drafts keep their negotiated line set authoritative —
	// an accepted below-MOQ deal (e.g. a sample order the agent negotiated) must
	// not 422 at confirm. Only lines the customer quantity-overrode (M-19) are
	// re-checked on inquiry drafts; unchanged lines were settled by the
	// negotiation. All other sources validate every line.
	if h.services.Order != nil {
		moqCheck := itemsToConfirm
		if order.Source == modelsOrder.OrderSourceInquiry {
			moqCheck = filterOverriddenItems(itemsToConfirm, overriddenQty)
		}
		if moqViolations := h.services.Order.ValidateMOQ(moqCheck, productByID); len(moqViolations) > 0 {
			c.JSON(http.StatusUnprocessableEntity, modelsCommon.ErrorResponse{
				Error:   "min_quantity_not_met",
				Message: i18n.T(c, "errors.min_quantity_not_met"),
				Details: gin.H{"violations": moqViolations},
			})
			return
		}
	}

	// G20: bulk/requisition/reorder drafts bypass the contract price-list minimum
	// quantity at creation (only the product-level MOQ is checked there — the
	// price-list min is a separate B2B tier constraint, enforced on the cart path
	// via validateLineMinQuantity). Enforce the price-list min at confirm for
	// those drafts so a below-contract line cannot be confirmed.
	//
	// G20 r3: the recheck now also covers M-19 quantity overrides on non-bulk
	// sources. A cart line validated at intake at qty 100 (>= min 50) could be
	// confirmed at an overridden qty 20 (< min 50) — the override is a new
	// quantity that never passed intake validation, so overridden lines are
	// re-checked on every source. Unchanged inquiry lines stay excluded: a
	// negotiated offer may legitimately sit below the list tier by design (H10),
	// and an unchanged line was already validated (or negotiated) at intake.
	if h.services.Price != nil && h.services.User != nil && h.services.Company != nil {
		var contractPriceListID *string
		if usr, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr == nil && usr.CompanyID != nil {
			if company, compErr := h.services.Company.GetCompany(c.Request.Context(), *usr.CompanyID); compErr == nil && company.PriceListID != nil {
				contractPriceListID = company.PriceListID
			}
		}
		if contractPriceListID != nil {
			// G20 (follow-up): enforce the contract price-list minimum PER LINE at
			// confirm so a split / per-line redistribution cannot bypass the
			// sub-min recheck. A draft line A qty 100 (>= min 50) confirmed as two
			// lines A qty 30 + A qty 70 keeps the per-product aggregate at 100 —
			// MOQ (aggregate) is fine, but each line must independently clear the
			// contract min. Only the H10 negotiated-inquiry path stays gated to
			// M-19-overridden lines (an unchanged negotiated deal may legitimately
			// sit below a contract tier); every other source re-checks every line
			// so the min gate no longer depends on the override map at all.
			minCheckItems := itemsToConfirm
			if order.Source == modelsOrder.OrderSourceInquiry {
				minCheckItems = filterOverriddenItems(itemsToConfirm, overriddenQty)
			}
			for _, it := range minCheckItems {
				if it.Quantity < 1 {
					continue
				}
				if !h.validateLineMinQuantity(c, it.ProductID, it.Quantity, contractPriceListID) {
					return
				}
			}
		}
	}

	// 确认时重新计算税/运费（地址或费率可能已变化），基于重定价后的行与 subtotal。
	confirmPricing, perr := h.computeCheckoutPricing(c.Request.Context(), checkoutPricingInput{
		Items:           itemsToConfirm,
		ProductByID:     productByID,
		ShippingAddress: order.ShippingAddress,
		Incoterms:       "FOB",
		Subtotal:        repricedSubtotal,
		Currency:        order.Currency,
	})
	if perr != nil {
		// G24c: a tax/shipping rate-lookup failure fails the confirm CLOSED — never
		// confirm with 0 tax/shipping on a transient DB failure.
		response.ErrorResp(c, http.StatusInternalServerError, "order_confirm_failed")
		return
	}
	if h.services.Channel != nil && confirmPricing.TaxAmount <= 0 {
		webstoreCh, _ := h.services.Channel.ResolveWebstoreChannel(c.Request.Context())
		confirmPricing.TaxAmount = h.services.Channel.ComputeTaxAmount(repricedSubtotal, webstoreCh, confirmPricing.TaxAmount)
	}
	// H10 follow-up: inquiry drafts carry the intake contract amount on
	// order.TotalAmount — the accepted negotiation offer's total, or the
	// line-price sum for a no-offer conversion. When the confirm line set
	// matches the draft (no M-19 quantity override / added line) and the draft
	// total is non-zero, that total IS the authoritative contract amount.
	// Rebuilding it as subtotal+tax+shipping would discard a deal-level
	// negotiated total that differs from the line-price sum (e.g. a
	// round-number / discounted offer), and the confirmed order would then
	// disagree with the trade created at intake (which carries order.TotalAmount).
	preserveNegotiatedTotal := order.Source == modelsOrder.OrderSourceInquiry &&
		order.TotalAmount > 0 &&
		confirmLinesMatchDraft(order.Items, itemsToConfirm)

	taxAmount := money.RoundMoney(confirmPricing.TaxAmount)
	shippingAmount := money.RoundMoney(confirmPricing.ShippingAmount)
	subtotal := money.RoundMoney(repricedSubtotal)
	totalAmount := money.RoundMoney(subtotal + taxAmount + shippingAmount)
	if preserveNegotiatedTotal {
		// The negotiated total is the all-inclusive contract value captured at
		// intake (intake stores tax/shipping as 0 and builds the trade from this
		// total). Keep tax/shipping at zero AND rescale the confirmed lines so
		// the persisted Subtotal equals the negotiated total exactly — the
		// invoice-derivation path (services/order/invoice_policy.go
		// CreateInvoiceFromOrder) derives Amount and TotalAmount from
		// Subtotal/TaxAmount/ShippingAmount and assumes
		// TotalAmount == Subtotal + Tax + Shipping. Persisting the intake draft's
		// divergence (line-sum Subtotal 2500 on a 2000 negotiated total) would
		// make the auto-derived invoice over-bill a discounted deal by exactly
		// the discount. Scaling restores the invariant so the confirmed order,
		// the derived invoice, the payment amount (order.TotalAmount) and the
		// intake trade all agree on the negotiated total.
		itemsToConfirm, subtotal = scaleInquiryLinesToTotal(itemsToConfirm, money.RoundMoney(order.TotalAmount))
		totalAmount = money.RoundMoney(order.TotalAmount)
		taxAmount = 0
		shippingAmount = 0
	}
	// G20: enforce the buyer company's credit limit cumulatively at confirm. The
	// per-order check at draft creation cannot catch a buyer stacking multiple
	// under-limit orders, and bulk/requisition/reorder drafts bypass the creation
	// check entirely, so the cumulative exposure (this order's total plus all
	// other open order totals) is validated here, before stock is committed.
	if !h.checkCompanyCreditLimitCumulative(c, userID, order.ID, totalAmount) {
		return
	}
	if !h.ensureActiveOrKYBBypassForAmount(c, userID, totalAmount, order.Currency, kyb.LineProductIDs(itemsToConfirm)...) {
		return
	}

	// H-15: compute financials BEFORE the atomic confirm so they land in the
	// same transaction as the status flip and stock reservation. Previously the
	// confirm tx and the financial update were separate, leaving stock reserved
	// with cogs=0 if the second update failed.
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

// confirmLinesMatchDraft reports whether the confirm-time line set (the server
// draft, or any M-19 client override after quantities/specifications are
// honored) matches the server draft by product and quantity. It gates H10 total
// preservation: an inquiry draft's negotiated TotalAmount is authoritative only
// for the exact negotiated line set — a quantity change or an added line
// invalidates it, so the total falls back to a recompute from the preserved (and
// for added lines, catalog) unit prices.
func confirmLinesMatchDraft(draft, confirm []modelsOrder.OrderItem) bool {
	if len(draft) != len(confirm) {
		return false
	}
	type key struct {
		productID string
		quantity  int
	}
	count := func(items []modelsOrder.OrderItem) map[key]int {
		out := make(map[key]int, len(items))
		for _, it := range items {
			pid := strings.TrimSpace(it.ProductID)
			if pid == "" || it.Quantity < 1 {
				continue
			}
			out[key{productID: pid, quantity: it.Quantity}]++
		}
		return out
	}
	draftCounts := count(draft)
	confirmCounts := count(confirm)
	if len(draftCounts) != len(confirmCounts) {
		return false
	}
	for k, v := range draftCounts {
		if confirmCounts[k] != v {
			return false
		}
	}
	return true
}

// filterOverriddenItems returns the subset of items whose productID was
// quantity-overridden by the customer (M-19) at confirm. Used to gate the
// confirm-time MOQ / contract price-list minimum re-checks on non-bulk drafts:
// an unchanged line already passed intake validation (or was negotiated, for
// H10 inquiry drafts) and must not be re-judged against catalog thresholds,
// while an overridden quantity is new and must be re-validated. Returns nil when
// nothing was overridden so the caller's ValidateMOQ short-circuits.
func filterOverriddenItems(items []modelsOrder.OrderItem, overridden map[string]bool) []modelsOrder.OrderItem {
	if len(overridden) == 0 {
		return nil
	}
	out := make([]modelsOrder.OrderItem, 0, len(items))
	for _, it := range items {
		if overridden[strings.TrimSpace(it.ProductID)] {
			out = append(out, it)
		}
	}
	return out
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
