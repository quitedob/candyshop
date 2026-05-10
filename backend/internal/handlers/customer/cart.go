package customer

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/kyb"
	"candypro/api/internal/pkg/response"
	"fmt"
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
	c.JSON(http.StatusOK, gin.H{"items": items, "itemCount": count})
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
		ShippingAddress modelsOrder.Address `json:"shippingAddress"`
		Notes           string              `json:"notes"`
		Currency        string              `json:"currency"`
	}
	_ = c.ShouldBindJSON(&req)

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
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), product, unitPrice, req.ShippingAddress.Country)
		if item.UnitPrice > 0 && unitPrice != item.UnitPrice {
			priceChangeWarnings = append(priceChangeWarnings, fmt.Sprintf("Price for %s updated from %.2f to %.2f", item.ProductID, item.UnitPrice, unitPrice))
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
		if item.Currency != "" && req.Currency == "" {
			currency = item.Currency
		}
	}

	// H1: Require shipping country for compliance validation
	if strings.TrimSpace(req.ShippingAddress.Country) == "" {
		response.InvalidResp(c, "target_country_required")
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
				"warnings":   inventory.Warnings,
			},
		})
		return
	}

	order := &modelsOrder.Order{
		ID:              generateCartOrderID(),
		OrderNumber:     generateCartOrderNumber(),
		UserID:          userID,
		Status:          "pending",
		PaymentStatus:   "unpaid",
		StockReserved:   true,
		Items:           orderItems,
		Subtotal:        subtotal,
		TotalAmount:     subtotal,
		Currency:        currency,
		ShippingAddress: req.ShippingAddress,
	}

	if !h.ensureActiveOrKYBBypassForAmount(c, userID, subtotal, kyb.CartLineProductIDs(items)...) {
		return
	}

	if err := h.services.Order.CreateOrderWithStockReservation(c.Request.Context(), order); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_create_failed")
		return
	}

	// N-10: Log cart clear errors but don't fail the checkout — order is already created
	if clearErr := h.services.Cart.ClearCart(c.Request.Context(), userID); clearErr != nil {
		c.Header("X-Cart-Clear-Warning", "cart_clear_failed")
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          order.ID,
		"orderNumber": order.OrderNumber,
		"status":      order.Status,
		"totalAmount": order.TotalAmount,
		"currency":    order.Currency,
		"compliance": gin.H{
			"country":  compliance.Country,
			"warnings": compliance.Warnings,
		},
		"inventory": gin.H{"warnings": append(inventory.Warnings, priceChangeWarnings...)},
	})
}

func generateCartOrderID() string {
	return crypto.GenerateID()
}

func generateCartOrderNumber() string {
	return fmt.Sprintf("ORD-%s-%s", time.Now().UTC().Format("20060102"), crypto.GenerateSlug()[:6])
}
