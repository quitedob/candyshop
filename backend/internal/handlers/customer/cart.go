package customer

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	countrypkg "candypro/api/internal/pkg/country"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/kyb"
	"candypro/api/internal/pkg/money"
	"candypro/api/internal/pkg/response"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// buildProductByIDMap builds a map of productID -> Product from a slice of products.
func buildProductByIDMap(products []modelsProduct.Product) map[string]modelsProduct.Product {
	m := make(map[string]modelsProduct.Product, len(products))
	for _, p := range products {
		m[p.ID] = p
	}
	return m
}

// CustomerGetCart returns all cart items for the authenticated customer.
// @Summary Get customer cart
// @Tags customer-cart
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /user/cart [get]
func (h *Handler) CustomerGetCart(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := h.services.Cart.GetCart(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "cart_fetch_failed")
		return
	}
	count, _ := h.services.Cart.ItemCount(c.Request.Context(), userID)

	productByID := buildProductByIDMapFromCartItems(c.Request.Context(), h, items)
	enriched := h.enrichCartItems(c.Request.Context(), items, productByID)
	resp := gin.H{"items": enriched, "itemCount": count}
	destination := countrypkg.NormalizeCountryCode(strings.TrimSpace(c.Query("destination")))
	region := strings.TrimSpace(c.Query("region"))
	incoterms := strings.TrimSpace(c.Query("incoterms"))
	var subtotal float64
	var weightKg float64
	for _, it := range items {
		subtotal += float64(it.Quantity) * it.UnitPrice
	}
	weightKg = sumCartWeightKg(items, productByID)

	taxEst := h.buildCartTaxEstimate(c.Request.Context(), subtotal, destination, region)
	shipEst := h.buildCartShippingEstimate(c.Request.Context(), destination, weightKg, incoterms)
	resp["taxEstimate"] = taxEst
	resp["shippingEstimate"] = shipEst

	summary := gin.H{
		"subtotal": subtotal,
		"pricingScope": "cart_reference", // 购物车阶段均为参考估算
	}
	if taxEst["status"] == feeEstimateComputed {
		summary["taxAmount"] = taxEst["amount"]
		summary["taxStatus"] = feeEstimateComputed
	} else {
		summary["taxStatus"] = taxEst["status"]
	}
	if shipEst["status"] == feeEstimateComputed {
		summary["shippingAmount"] = shipEst["cost"]
		summary["shippingStatus"] = feeEstimateComputed
	} else {
		summary["shippingStatus"] = shipEst["status"]
	}
	resp["summary"] = summary
	c.JSON(http.StatusOK, resp)
}

// enrichCartItems 批量关联 Product 表，补充缩略图、徽章、MOQ、库存上限等展示字段
func (h *Handler) enrichCartItems(ctx context.Context, items []modelsOrder.CartItem, productByID map[string]modelsProduct.Product) []modelsOrder.CartItemResponse {
	if len(items) == 0 {
		return []modelsOrder.CartItemResponse{}
	}
	if productByID == nil {
		productByID = buildProductByIDMapFromCartItems(ctx, h, items)
	}
	productIDs := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, it := range items {
		if it.ProductID == "" {
			continue
		}
		if _, ok := seen[it.ProductID]; ok {
			continue
		}
		seen[it.ProductID] = struct{}{}
		productIDs = append(productIDs, it.ProductID)
	}
	sellable := map[string]int{}
	if h.services.Product != nil && len(productIDs) > 0 {
		if m, serr := h.services.Product.EffectiveSellableByProducts(ctx, productIDs, modelsProduct.ChannelWebstore); serr == nil {
			sellable = m
		}
	}
	out := make([]modelsOrder.CartItemResponse, 0, len(items))
	for _, it := range items {
		row := modelsOrder.CartItemResponseFromItem(it)
		if p, ok := productByID[it.ProductID]; ok {
			if p.Name != "" {
				row.Name = p.Name
				if row.ProductName == "" {
					row.ProductName = p.Name
				}
			}
			row.Thumbnail = p.Thumbnail
			row.HalalCertified = p.HalalCertified
			row.OEMAvailable = p.OEMAvailable
			if p.MOQ > 0 {
				row.MOQ = p.MOQ
			}
			row.Category = p.Category
			row.Translations = p.Translations
			row.MaxQuantity = modelsOrder.ResolveCartMaxQuantity(it.Quantity, sellable[it.ProductID])
		}
		out = append(out, row)
	}
	return out
}

// buildProductByIDMapFromCartItems 按购物车行批量加载产品（去重）
func buildProductByIDMapFromCartItems(ctx context.Context, h *Handler, items []modelsOrder.CartItem) map[string]modelsProduct.Product {
	m := make(map[string]modelsProduct.Product, len(items))
	if h.services == nil || h.services.Product == nil {
		return m
	}
	for _, it := range items {
		if it.ProductID == "" {
			continue
		}
		if _, ok := m[it.ProductID]; ok {
			continue
		}
		if p, err := h.services.Product.GetProductByID(ctx, it.ProductID); err == nil && p != nil {
			m[it.ProductID] = *p
		}
	}
	return m
}

// CustomerAddToCart adds a product to the cart.
func (h *Handler) CustomerAddToCart(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ProductID          string  `json:"productId" binding:"required"`
		ProductName        string  `json:"productName"`
		Quantity           int     `json:"quantity"`
		UnitPrice          float64 `json:"unitPrice"`
		Currency           string  `json:"currency"`
		Specifications     string  `json:"specifications"`
		DestinationCountry string  `json:"destinationCountry"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// N-02: Validate product exists and use server-side price
	product, err := h.services.Product.GetProductByID(c.Request.Context(), req.ProductID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}
	if product.Status != modelsProduct.ProductStatusActive {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "product_unavailable")
		return
	}

	// Normalize quantity before pricing and persistence
	qty := req.Quantity
	if qty < 1 {
		qty = 1
	}

	// Use server-side price: try contract price first, then product base price
	var unitPrice float64
	if h.services.Price != nil && h.services.User != nil {
		if user, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr == nil && user.CompanyID != nil {
			if company, compErr := h.services.Company.GetCompany(c.Request.Context(), *user.CompanyID); compErr == nil && company.PriceListID != nil {
				if contractPrice, priceErr := h.services.Price.GetPriceForProduct(c.Request.Context(), req.ProductID, *company.PriceListID, qty); priceErr == nil {
					unitPrice = contractPrice
				}
			}
		}
	}
	if unitPrice <= 0 {
		unitPrice = product.BasePrice
	}
	if unitPrice <= 0 {
		response.ErrorResp(c, http.StatusUnprocessableEntity, "no_price")
		return
	}
	if strings.TrimSpace(req.DestinationCountry) != "" {
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), product, unitPrice, req.DestinationCountry)
	}
	unitPrice = h.applyChannelUnitPrice(c, unitPrice)

	productName := req.ProductName
	if productName == "" {
		productName = product.Name
	}
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	item := &modelsOrder.CartItem{
		UserID:         userID,
		ProductID:      req.ProductID,
		ProductName:    productName,
		Quantity:       qty,
		UnitPrice:      unitPrice,
		Currency:       currency,
		Specifications: req.Specifications,
	}
	cartItems, _ := h.services.Cart.GetCart(c.Request.Context(), userID)
	proj := projectedCartUSDAfterAdd(cartItems, req.ProductID, qty, unitPrice)
	pids := kyb.CartLineProductIDs(cartItems, req.ProductID)
	if !h.ensureActiveOrKYBBypassForAmount(c, userID, proj, pids...) {
		return
	}

	created, err := h.services.Cart.AddItem(c.Request.Context(), userID, item)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "cart_add_failed")
		return
	}
	c.JSON(http.StatusCreated, created)
}

// CustomerUpdateCartItem updates the quantity of a cart item.
func (h *Handler) CustomerUpdateCartItem(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	cartItems, _ := h.services.Cart.GetCart(c.Request.Context(), userID)
	proj := projectedCartUSDAfterQtyChange(cartItems, uint(itemID), req.Quantity)
	if !h.ensureActiveOrKYBBypassForAmount(c, userID, proj, kyb.CartLineProductIDs(cartItems)...) {
		return
	}

	updated, err := h.services.Cart.UpdateItem(c.Request.Context(), userID, uint(itemID), req.Quantity)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "cart_update_failed")
		return
	}
	c.JSON(http.StatusOK, updated)
}

// CustomerRemoveCartItem removes a single item from the cart.
func (h *Handler) CustomerRemoveCartItem(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if err := h.services.Cart.RemoveItem(c.Request.Context(), userID, uint(itemID)); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "cart_item_not_found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

// CustomerClearCart removes all items from the cart.
func (h *Handler) CustomerClearCart(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.services.Cart.ClearCart(c.Request.Context(), userID); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "cart_clear_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
}

// CustomerCheckoutCart converts the cart into a pending order.
func (h *Handler) CustomerCheckoutCart(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := h.services.Cart.GetCart(c.Request.Context(), userID)
	if err != nil || len(items) == 0 {
		response.ErrorResp(c, http.StatusBadRequest, "cart_empty")
		return
	}

	var req struct {
		ShippingAddress   modelsOrder.Address `json:"shippingAddress"`
		Notes             string              `json:"notes"`
		Currency          string              `json:"currency"`
		EstimatedWeightKg float64             `json:"estimated_weight_kg"`
		CouponCode        string              `json:"couponCode"`
		Incoterms         string              `json:"incoterms"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// Build order items from cart and collect product details for validation
	orderItems := make(modelsOrder.OrderItemArray, 0, len(items))
	selectedProducts := make([]modelsProduct.Product, 0, len(items))
	productByID := make(map[string]modelsProduct.Product, len(items))
	var subtotal float64
	currency := "USD"
	if req.Currency != "" {
		currency = req.Currency
	}

	// Resolve contract price list for the user (if applicable)
	var contractPriceListID *string
	if h.services.Price != nil && h.services.User != nil && h.services.Company != nil {
		if usr, userErr := h.services.User.GetByID(c.Request.Context(), userID); userErr == nil && usr.CompanyID != nil {
			if company, compErr := h.services.Company.GetCompany(c.Request.Context(), *usr.CompanyID); compErr == nil && company.PriceListID != nil {
				contractPriceListID = company.PriceListID
			}
		}
	}

	var priceChangeWarnings []string
	for _, item := range items {
		product, productErr := h.services.Product.GetProductByID(c.Request.Context(), item.ProductID)
		if productErr != nil {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "product_not_found")
			return
		}
		if status := strings.ToLower(strings.TrimSpace(product.Status)); status != "" && status != "active" {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "product_unavailable")
			return
		}

		// Re-resolve price server-side at checkout time (not stale cart price)
		var unitPrice float64
		if contractPriceListID != nil && h.services.Price != nil {
			if cp, priceErr := h.services.Price.GetPriceForProduct(c.Request.Context(), item.ProductID, *contractPriceListID, item.Quantity); priceErr == nil {
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
		if !h.validateLineMinQuantity(c, item.ProductID, item.Quantity, contractPriceListID) {
			return
		}
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), product, unitPrice, req.ShippingAddress.Country)
		unitPrice = h.applyChannelUnitPrice(c, unitPrice)
		if item.UnitPrice > 0 && unitPrice != item.UnitPrice {
			productName := strings.TrimSpace(product.Name)
			if productName == "" {
				productName = strings.TrimSpace(product.Slug)
			}
			if productName == "" {
				productName = item.ProductID
			}
			priceChangeWarnings = append(priceChangeWarnings, i18n.TWithVars(c, "messages.price_updated", map[string]string{
				"productName": productName,
				"oldPrice":    fmt.Sprintf("%.2f", item.UnitPrice),
				"newPrice":    fmt.Sprintf("%.2f", unitPrice),
			}))
		}

		orderItems = append(orderItems, modelsOrder.OrderItem{
			ProductID:      item.ProductID,
			Quantity:       item.Quantity,
			UnitPrice:      unitPrice,
			Specifications: item.Specifications,
		})
		selectedProducts = append(selectedProducts, *product)
		productByID[item.ProductID] = *product
		subtotal += float64(item.Quantity) * unitPrice
		if currency == "USD" && item.Currency != "" {
			currency = item.Currency
		}
	}

	// H1: Require shipping country for compliance validation
	if strings.TrimSpace(req.ShippingAddress.Country) == "" {
		response.InvalidResp(c, "target_country_required")
		return
	}
	if strings.TrimSpace(req.ShippingAddress.Street) == "" {
		response.InvalidResp(c, "shipping_street_required")
		return
	}
	if strings.TrimSpace(req.ShippingAddress.City) == "" {
		response.InvalidResp(c, "shipping_city_required")
		return
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
	inventory := h.services.Order.ValidateInventoryWithSellable(orderItems, productByID, sellable)
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

	pricing := h.computeCheckoutPricing(c.Request.Context(), checkoutPricingInput{
		Items:             orderItems,
		ProductByID:       productByID,
		ShippingAddress:   req.ShippingAddress,
		Incoterms:         req.Incoterms,
		EstimatedWeightKg: req.EstimatedWeightKg,
		Subtotal:          subtotal,
		Currency:          currency,
	})
	shippingAmount := pricing.ShippingAmount
	shippingCurrency := currency
	if pricing.Currency != "" {
		shippingCurrency = pricing.Currency
	}
	taxAmount := pricing.TaxAmount
	if h.services.Channel != nil {
		webstoreCh, _ := h.services.Channel.ResolveWebstoreChannel(c.Request.Context())
		if taxAmount <= 0 {
			taxAmount = h.services.Channel.ComputeTaxAmount(subtotal, webstoreCh, taxAmount)
		}
	}
	taxAmount = money.RoundMoney(taxAmount)
	shippingAmount = money.RoundMoney(shippingAmount)
	checkoutTotal := money.RoundMoney(subtotal + taxAmount + shippingAmount)
	hasOfficialEvidence := len(compliance.Violations) == 0 && len(compliance.Warnings) == 0
	couponCode := strings.TrimSpace(req.CouponCode)
	if couponCode != "" && h.services.Coupon != nil {
		if _, _, cerr := h.services.Coupon.PreviewCouponDiscount(c.Request.Context(), couponCode, userID, checkoutTotal); cerr != nil {
			response.ErrorResp(c, http.StatusBadRequest, "coupon_apply_failed")
			return
		}
	}
	if !h.checkCompanyCreditLimit(c, userID, checkoutTotal) {
		return
	}
	destCountry := countrypkg.NormalizeCountryCode(strings.TrimSpace(req.ShippingAddress.Country))
	incoterms := strings.ToUpper(strings.TrimSpace(req.Incoterms))
	if incoterms == "" {
		incoterms = "FOB"
	}
	estimatedWeightKg := req.EstimatedWeightKg
	if estimatedWeightKg <= 0 {
		for _, item := range orderItems {
			if product, ok := productByID[item.ProductID]; ok && product.GrossWeightPerCarton > 0 {
				estimatedWeightKg += float64(item.Quantity) * product.GrossWeightPerCarton
			}
		}
	}
	orderStatus := modelsOrder.OrderStatusPending
	if h.services.Approval != nil {
		if needsApproval, _, _ := h.services.Approval.ShouldRequireApproval(c.Request.Context(), userID, checkoutTotal); needsApproval {
			orderStatus = modelsOrder.OrderStatusPendingApproval
		}
	}

	now := time.Now()
	order := &modelsOrder.Order{
		ID:                         generateCartOrderID(),
		OrderNumber:                generateCartOrderNumber(),
		UserID:                     userID,
		Source:                     modelsOrder.OrderSourceCart,
		Status:                     orderStatus,
		PaymentStatus:              "unpaid",
		StockReserved:              false,
		ComplianceOfficialEvidence: hasOfficialEvidence,
		Items:                      orderItems,
		Subtotal:                   subtotal,
		TaxAmount:                  taxAmount,
		ShippingAmount:             shippingAmount,
		TotalAmount:                checkoutTotal,
		Currency:                   shippingCurrency,
		ShippingAddress:            req.ShippingAddress,
	}
	if orderStatus == modelsOrder.OrderStatusPending {
		order.ConfirmedAt = &now
	}
	h.assignOrderWarehouseID(c, &order.WarehouseID)

	if !h.ensureActiveOrKYBBypassForAmount(c, userID, subtotal, kyb.CartLineProductIDs(items)...) {
		return
	}

	var createErr error
	if orderStatus == modelsOrder.OrderStatusPending {
		createErr = h.services.Order.CreateOrderWithStockReservation(c.Request.Context(), order)
	} else {
		createErr = h.services.Order.CreateOrder(c.Request.Context(), order)
	}
	if createErr != nil {
		if errors.Is(createErr, modelsOrder.ErrInsufficientStock) {
			response.ErrorResp(c, http.StatusUnprocessableEntity, "insufficient_stock")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		return
	}

	if couponCode != "" && h.services.Coupon != nil {
		if _, cerr := h.services.Coupon.ApplyCouponToCart(c.Request.Context(), order.ID, userID, couponCode); cerr != nil {
			slog.Warn("checkout coupon apply failed after order create", "orderId", order.ID, "error", cerr)
		}
	}

	if orderStatus == modelsOrder.OrderStatusPendingApproval {
		h.notifyOrderApprovers(c, order)
	}

	// N-10: Log cart clear errors but don't fail the checkout — order is already created
	if clearErr := h.services.Cart.ClearCart(c.Request.Context(), userID); clearErr != nil {
		c.Header("X-Cart-Clear-Warning", "cart_clear_failed")
	}

	checkoutTaxSt := checkoutTaxStatus(c.Request.Context(), h, destCountry, req.ShippingAddress.State, subtotal, taxAmount)
	checkoutShipSt := checkoutShippingStatus(c.Request.Context(), h, destCountry, estimatedWeightKg, incoterms)

	c.JSON(http.StatusCreated, gin.H{
		"id":          order.ID,
		"orderNumber": order.OrderNumber,
		"status":      order.Status,
		"totalAmount": order.TotalAmount,
		"currency":    order.Currency,
		"pricing": gin.H{
			"scope":          "order_snapshot",
			"subtotal":       order.Subtotal,
			"taxAmount":      order.TaxAmount,
			"shippingAmount": order.ShippingAmount,
			"totalAmount":    order.TotalAmount,
			"taxStatus":      checkoutTaxSt,
			"shippingStatus": checkoutShipSt,
			"notice":         pricingNoticeForCheckout(checkoutTaxSt, checkoutShipSt),
		},
		"compliance": gin.H{
			"country":  compliance.Country,
			"warnings": compliance.Warnings,
		},
		"inventory": gin.H{"warnings": append(formatInventoryWarnings(c, inventory.Warnings, productByID), priceChangeWarnings...)},
	})

	// Send order confirmation notification
	if h.services.Notification != nil {
		vars := map[string]string{"orderNumber": order.OrderNumber}
		// H-24: log notification failures so silent drops are detectable.
		// We don't fail the checkout because the order is already committed.
		if err := h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
			UserID:    userID,
			Type:      "order",
			Reference: order.ID,
			Title:     i18n.TWithVars(c, "notifications.order_placed_title", vars),
			Message:   i18n.TWithVars(c, "notifications.order_placed_message", vars),
		}); err != nil {
			slog.Warn("checkout: failed to create order notification",
				"orderID", order.ID,
				"userID", userID,
				"error", err,
			)
		}
	}
}

func generateCartOrderID() string {
	return crypto.GenerateID()
}

func generateCartOrderNumber() string {
	return fmt.Sprintf("ORD-%s-%s", time.Now().UTC().Format("20060102"), crypto.GenerateSlug()[:6])
}
