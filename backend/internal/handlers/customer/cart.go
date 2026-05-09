package customer

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/kyb"
	"candypro/api/internal/utils"
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
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{Error: "unauthorized", Message: "User not identified"})
		return
	}
	items, err := h.services.Cart.GetCart(c.Request.Context(), userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch cart")
		return
	}
	count, _ := h.services.Cart.ItemCount(c.Request.Context(), userID)
	c.JSON(http.StatusOK, gin.H{"items": items, "itemCount": count})
}

// CustomerAddToCart adds a product to the cart.
func (h *Handler) CustomerAddToCart(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{Error: "unauthorized", Message: "User not identified"})
		return
	}

	var req struct {
		ProductID          string  `json:"productId" binding:"required"`
		ProductName        string  `json:"productName"`
		Quantity           int     `json:"quantity"`
		UnitPrice          float64 `json:"unitPrice"`
		Currency           string  `json:"currency"`
		Specifications     string  `json:"specifications"`
		DestinationCountry string  `json:"destinationCountry"` // 可选：用于套目的国成本栈展示价
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	// N-02: Validate product exists and use server-side price
	product, err := h.services.Product.GetProductByID(c.Request.Context(), req.ProductID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{Error: "not_found", Message: "Product not found"})
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
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "no_price",
			Message: fmt.Sprintf("No price available for product %s. Please contact support.", req.ProductID),
		})
		return
	}
	// 若提供目的国，则在基础/合同价上叠加 ProductMarketCostStack
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to add to cart")
		return
	}
	c.JSON(http.StatusCreated, created)
}

// CustomerUpdateCartItem updates the quantity of a cart item.
func (h *Handler) CustomerUpdateCartItem(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{Error: "unauthorized", Message: "User not identified"})
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid item ID")
		return
	}

	var req struct {
		Quantity int `json:"quantity" binding:"required,min=1"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	cartItems, _ := h.services.Cart.GetCart(c.Request.Context(), userID)
	proj := projectedCartUSDAfterQtyChange(cartItems, uint(itemID), req.Quantity)
	if !h.ensureActiveOrKYBBypassForAmount(c, userID, proj, kyb.CartLineProductIDs(cartItems)...) {
		return
	}

	updated, err := h.services.Cart.UpdateItem(c.Request.Context(), userID, uint(itemID), req.Quantity)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	c.JSON(http.StatusOK, updated)
}

// CustomerRemoveCartItem removes a single item from the cart.
func (h *Handler) CustomerRemoveCartItem(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{Error: "unauthorized", Message: "User not identified"})
		return
	}

	itemIDStr := c.Param("itemId")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 64)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid item ID")
		return
	}

	if err := h.services.Cart.RemoveItem(c.Request.Context(), userID, uint(itemID)); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Cart item not found")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart"})
}

// CustomerClearCart removes all items from the cart.
func (h *Handler) CustomerClearCart(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{Error: "unauthorized", Message: "User not identified"})
		return
	}
	if err := h.services.Cart.ClearCart(c.Request.Context(), userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to clear cart")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
}

// CustomerCheckoutCart converts the cart into a pending order.
func (h *Handler) CustomerCheckoutCart(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{Error: "unauthorized", Message: "User not identified"})
		return
	}

	items, err := h.services.Cart.GetCart(c.Request.Context(), userID)
	if err != nil || len(items) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "cart_empty", "Cart is empty")
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
			utils.ErrorResponse(c, http.StatusUnprocessableEntity, "product_not_found",
				fmt.Sprintf("Product %s in cart no longer exists", item.ProductID))
			return
		}
		// H2: Check product status — reject inactive/draft products
		if status := strings.ToLower(strings.TrimSpace(product.Status)); status != "" && status != "active" {
			utils.ErrorResponse(c, http.StatusUnprocessableEntity, "product_unavailable",
				fmt.Sprintf("Product %s is no longer available for ordering", item.ProductID))
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
			utils.ErrorResponse(c, http.StatusUnprocessableEntity, "no_price",
				fmt.Sprintf("No price available for product %s. Please contact support.", item.ProductID))
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
		utils.InvalidRequestResponse(c, "shippingAddress.country is required for compliance validation")
		return
	}

	// Validate country compliance (same as CustomerCreateOrder，含市场画像)
	compliance := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), req.ShippingAddress.Country, selectedProducts)
	if len(compliance.Violations) > 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "compliance_violation",
			Message: "Cart contains products that violate destination-country compliance requirements",
			Details: gin.H{
				"country":    compliance.Country,
				"violations": compliance.Violations,
				"warnings":   compliance.Warnings,
			},
		})
		return
	}

	// Validate inventory（含 OMS 自建站渠道可售量封顶）
	ids := make([]string, 0, len(productByID))
	for id := range productByID {
		ids = append(ids, id)
	}
	sellable, _ := h.services.Product.EffectiveSellableByProducts(c.Request.Context(), ids, modelsProduct.ChannelWebstore)
	inventory := h.services.Order.ValidateInventoryWithSellable(orderItems, productByID, sellable)
	if len(inventory.Violations) > 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "inventory_violation",
			Message: "Cart contains items that exceed inventory policy constraints",
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
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create order from cart")
		return
	}

	// N-10: Log cart clear errors but don't fail the checkout — order is already created
	if clearErr := h.services.Cart.ClearCart(c.Request.Context(), userID); clearErr != nil {
		// Order succeeded; log the error for ops visibility but return success to client
		// The cart will appear stale until the next session or manual clear
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
	return utils.GenerateID()
}

func generateCartOrderNumber() string {
	return fmt.Sprintf("ORD-%s-%s", time.Now().UTC().Format("20060102"), utils.GenerateSlug()[:6])
}
