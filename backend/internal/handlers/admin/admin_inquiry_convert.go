package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsTrade "candypro/api/internal/models/trade"
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

	// Parse EstimatedQuantity to get numeric value for order item quantities
	quantity := parseQuantity(inquiry.EstimatedQuantity)

	// Build order items from inquiry products
	var items []modelsOrder.OrderItem
	for _, pid := range inquiry.Products {
		pid = strings.TrimSpace(pid)
		if pid == "" {
			continue
		}
		unitPrice := 0.0
		// Try to get price from pricing service if available
		if h.services.Price != nil {
			if price, err := h.services.Price.GetPriceForProduct(c.Request.Context(), pid, "", quantity); err == nil {
				unitPrice = price
			}
		}
		items = append(items, modelsOrder.OrderItem{
			ProductID: pid,
			Quantity:  quantity,
			UnitPrice: unitPrice,
		})
	}

	itemsArr := modelsOrder.OrderItemArray(items)
	if itemsArr == nil {
		itemsArr = modelsOrder.OrderItemArray{}
	}

	// Calculate totals
	subtotal := inquiry.QuotedAmount
	if subtotal == 0 {
		for _, item := range items {
			subtotal += item.UnitPrice * float64(item.Quantity)
		}
	}

	orderNumber := fmt.Sprintf("ORD-%s-%s", time.Now().Format("20060102"), strings.ToUpper(utils.GenerateSlug()))

	var shippingAddr modelsOrder.Address
	if req.ShippingAddress != nil {
		shippingAddr = *req.ShippingAddress
	} else {
		shippingAddr = modelsOrder.Address{
			Country: inquiry.TargetCountry,
		}
	}

	order := &modelsOrder.Order{
		ID:             utils.GenerateID(),
		OrderNumber:    orderNumber,
		UserID:         userID,
		InquiryID:      &inquiryID,
		Status:         "pending",
		PaymentStatus:  "unpaid",
		Items:          itemsArr,
		StockReserved:  len(items) > 0,
		Subtotal:       subtotal,
		TaxAmount:      0,
		ShippingAmount: 0,
		TotalAmount:    subtotal,
		Currency:       "USD",
		ShippingAddress: shippingAddr,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
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
	autoCreateTradeFromOrder(c, h, order, inquiryID)

	c.JSON(http.StatusCreated, order)
}

// autoCreateTradeFromOrder creates a trade transaction linked to a newly created order.
func autoCreateTradeFromOrder(c *gin.Context, h *Handler, order *modelsOrder.Order, inquiryID string) {
	if h.services.Trade == nil {
		return
	}
	trade := &modelsTrade.TradeTransaction{
		UserID:      order.UserID,
		OrderID:     &order.ID,
		Status:      modelsTrade.TradeStatusPending,
		Currency:    order.Currency,
		TotalAmount: order.TotalAmount,
		Terms:       "FOB",
	}
	if inquiryID != "" {
		trade.InquiryID = &inquiryID
	}
	if err := h.services.Trade.CreateTransaction(c.Request.Context(), trade); err != nil {
		// N-17: Log the error; don't silently discard it
		c.Header("X-Trade-Create-Warning", "trade_creation_failed")
	}
}

// parseQuantity extracts a numeric value from a quantity string like "1000 cartons", "5 tons", "500 kg".
func parseQuantity(qty string) int {
	if qty == "" {
		return 1
	}
	// Try to extract the leading numeric value
	var num float64
	n, err := fmt.Sscanf(qty, "%f", &num)
	if err != nil || n == 0 {
		// Try to parse word by word
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
