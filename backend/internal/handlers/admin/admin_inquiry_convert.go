package admin

import (
	"errors"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type adminConvertInquiryRequest struct {
	ShippingAddress *modelsOrder.Address `json:"shippingAddress"`
	Notes           string               `json:"notes"`
}

// AdminConvertInquiryToOrder converts a quoted/won inquiry into a draft order.
func (h *Handler) AdminConvertInquiryToOrder(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Inquiry not found",
		})
		return
	}

	// Only quoted or won inquiries can be converted
	if inquiry.Status != "quoted" && inquiry.Status != "won" && inquiry.Status != "negotiating" {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "invalid_status",
			Message: "Only inquiries with status 'quoted', 'won', or 'negotiating' can be converted to orders",
		})
		return
	}

	// Must have a user assigned
	if inquiry.UserID == nil || strings.TrimSpace(*inquiry.UserID) == "" {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "missing_user",
			Message: "Inquiry must have an assigned user before converting to order",
		})
		return
	}

	var req adminConvertInquiryRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	userID := strings.TrimSpace(*inquiry.UserID)

	// Determine destination country
	destCountry := ""
	if req.ShippingAddress != nil {
		destCountry = strings.TrimSpace(req.ShippingAddress.Country)
	}
	if destCountry == "" {
		destCountry = strings.TrimSpace(inquiry.TargetCountry)
	}
	if destCountry == "" {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "missing_country",
			Message: "Destination country is required for order conversion (provide shippingAddress.country or set inquiry targetCountry)",
		})
		return
	}

	// Parse EstimatedQuantity to get numeric value for order item quantities
	quantity := parseQuantity(inquiry.EstimatedQuantity)

	// Build order items from inquiry products with full validation
	var items []modelsOrder.OrderItem
	selectedProducts := make([]modelsProduct.Product, 0, len(inquiry.Products))
	productByID := make(map[string]modelsProduct.Product, len(inquiry.Products))

	for _, pid := range inquiry.Products {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}

		// Validate product exists and is active
		product, prodErr := h.services.Product.GetProductByID(c.Request.Context(), pid)
		if prodErr != nil {
			c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
				Error:   "product_not_found",
				Message: fmt.Sprintf("Product %s in inquiry no longer exists", pid),
			})
			return
		}
		if status := strings.ToLower(strings.TrimSpace(product.Status)); status != "" && status != "active" {
			c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
				Error:   "product_unavailable",
				Message: fmt.Sprintf("Product %s is no longer available for ordering", pid),
			})
			return
		}

		// Resolve price: try price service, then product base price
		var unitPrice float64
		if h.services.Price != nil {
			if price, priceErr := h.services.Price.GetPriceForProduct(c.Request.Context(), pid, "", quantity); priceErr == nil {
				unitPrice = price
			}
		}
		if unitPrice <= 0 {
			unitPrice = product.BasePrice
		}
		if unitPrice <= 0 {
			c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
				Error:   "no_price",
				Message: fmt.Sprintf("No price available for product %s. Set a base price or contract price before converting.", pid),
			})
			return
		}

		// Apply market cost stack if available
		unitPrice = h.services.Product.ResolveCheckoutUnitPrice(c.Request.Context(), product, unitPrice, destCountry)

		items = append(items, modelsOrder.OrderItem{
			ProductID: pid,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		})
		selectedProducts = append(selectedProducts, *product)
		productByID[pid] = *product
	}

	if len(items) == 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "no_items",
			Message: "Inquiry has no valid products to convert",
		})
		return
	}

	// Validate destination-country compliance
	compliance := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), destCountry, selectedProducts)
	if len(compliance.Violations) > 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "compliance_violation",
			Message: "Order violates destination-country compliance requirements",
			Details: gin.H{
				"country":    compliance.Country,
				"violations": compliance.Violations,
				"warnings":   compliance.Warnings,
			},
		})
		return
	}

	// Validate inventory
	ids := make([]string, 0, len(productByID))
	for id := range productByID {
		ids = append(ids, id)
	}
	sellable, _ := h.services.Product.EffectiveSellableByProducts(c.Request.Context(), ids, modelsProduct.ChannelWebstore)
	inventory := h.services.Order.ValidateInventoryWithSellable(items, productByID, sellable)
	if len(inventory.Violations) > 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "inventory_violation",
			Message: "Insufficient stock or inventory policy violation for one or more products",
			Details: gin.H{
				"violations": inventory.Violations,
				"warnings":   inventory.Warnings,
			},
		})
		return
	}

	// Calculate totals from validated items
	var subtotal float64
	for _, item := range items {
		subtotal += item.UnitPrice * float64(item.Quantity)
	}

	orderNumber := fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), strings.ToUpper(utils.GenerateSlug()))

	var shippingAddr modelsOrder.Address
	if req.ShippingAddress != nil {
		shippingAddr = *req.ShippingAddress
	} else {
		shippingAddr = modelsOrder.Address{
			Country: destCountry,
		}
	}

	order := &modelsOrder.Order{
		ID:              utils.GenerateID(),
		OrderNumber:     orderNumber,
		UserID:          userID,
		InquiryID:       &inquiryID,
		Status:          "pending",
		PaymentStatus:   "unpaid",
		Items:           items,
		StockReserved:   true,
		Subtotal:        subtotal,
		TaxAmount:       0,
		ShippingAmount:  0,
		TotalAmount:     subtotal,
		Currency:        "USD",
		ShippingAddress: shippingAddr,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := h.services.Order.CreateOrderWithStockReservation(c.Request.Context(), order); err != nil {
		if errors.Is(err, modelsOrder.ErrInsufficientStock) {
			c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
				Error:   "inventory_violation",
				Message: "Insufficient stock for one or more products in the inquiry",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create order from inquiry",
		})
		return
	}

	// Auto-create a trade transaction linked to this order
	autoCreateTradeFromOrder(c, h, order, inquiry)

	c.JSON(http.StatusCreated, order)
}

// autoCreateTradeFromOrder creates a trade transaction linked to a newly created order.
func autoCreateTradeFromOrder(c *gin.Context, h *Handler, order *modelsOrder.Order, inquiry *modelsProduct.Inquiry) {
	if h.services.Trade == nil {
		return
	}
	trade := orderSvc.BuildTradeTransactionFromInquiry(order, inquiry)
	if err := h.services.Trade.CreateTransaction(c.Request.Context(), trade); err != nil {
		c.Header("X-Trade-Create-Warning", "trade_creation_failed")
	}
}

// parseQuantity extracts a numeric value from a quantity string like "1000 cartons", "5 tons", "500 kg".
func parseQuantity(qty string) int {
	if qty == "" {
		return 1
	}
	var num float64
	n, err := fmt.Sscanf(qty, "%f", &num)
	if err != nil || n == 0 {
		parts := strings.Fields(qty)
		for _, p := range parts {
			n2, err2 := fmt.Sscanf(p, "%f", &num)
			if err2 == nil && n2 > 0 {
				break
			}
		}
	}
	if num <= 0 {
		return 1
	}
	return int(num)
}
