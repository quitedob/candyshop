package customer

import (
	"bytes"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
		return
	}

	var req customerCreateOrderRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	if req.TaxAmount < 0 || req.ShippingAmount < 0 {
		utils.InvalidRequestResponse(c, "taxAmount and shippingAmount cannot be negative")
		return
	}

	if msg := validateCustomerShippingAddress(req.ShippingAddress); msg != "" {
		utils.InvalidRequestResponse(c, msg)
		return
	}

	var inquiryID *string
	if req.InquiryID != nil && strings.TrimSpace(*req.InquiryID) != "" {
		trimmedInquiryID := strings.TrimSpace(*req.InquiryID)
		inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), trimmedInquiryID)
		if err != nil {
			c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
				Error:   "not_found",
				Message: "Inquiry not found",
			})
			return
		}
		if inquiry.UserID == nil || *inquiry.UserID != userID {
			c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
				Error:   "forbidden",
				Message: "You do not have access to this inquiry",
			})
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
			utils.InvalidRequestResponse(c, "items.productId is required")
			return
		}
		if it.Quantity < 1 {
			utils.InvalidRequestResponse(c, "items.quantity must be greater than 0")
			return
		}

		product, err := h.services.Product.GetProductByID(c.Request.Context(), productID)
		if err != nil {
			c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
				Error:   "not_found",
				Message: fmt.Sprintf("Product %s not found", productID),
			})
			return
		}
		if status := strings.ToLower(strings.TrimSpace(product.Status)); status != "" && status != "active" {
			utils.InvalidRequestResponse(c, fmt.Sprintf("Product %s is not available for ordering", productID))
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
			utils.InvalidRequestResponse(c, fmt.Sprintf("No price available for product %s. Please contact support.", productID))
			return
		}

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

	compliance := h.services.Order.ValidateCountryCompliance(req.ShippingAddress.Country, selectedProducts)
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
	inventory := h.services.Order.ValidateInventory(items, productByID)
	if len(inventory.Violations) > 0 {
		c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
			Error:   "inventory_violation",
			Message: "Order exceeds inventory policy constraints",
			Details: gin.H{
				"violations": inventory.Violations,
				"warnings":   inventory.Warnings,
			},
		})
		return
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	orderNumber := strings.TrimSpace(req.OrderNumber)
	if orderNumber == "" {
		orderNumber = buildCustomerOrderNumber()
	}

	totalAmount := subtotal + req.TaxAmount + req.ShippingAmount
	now := time.Now()
	order := &modelsOrder.Order{
		ID:             utils.GenerateID(),
		OrderNumber:    orderNumber,
		UserID:         userID,
		InquiryID:      inquiryID,
		Status:         "pending",
		PaymentStatus:  "unpaid",
		Items:          items,
		StockReserved:  true,
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

	if err := h.services.Order.CreateOrderWithStockReservation(c.Request.Context(), order); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "insufficient stock") {
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
}

// CustomerConfirmOrder confirms an AI-drafted order before final processing.
func (h *Handler) CustomerConfirmOrder(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	if orderID == "" {
		utils.InvalidRequestResponse(c, "order id is required")
		return
	}

	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}
	if order.UserID != userID {
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have access to this order",
		})
		return
	}

	if order.Status != "pending_confirmation" || !order.StockReserved {
		c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
			Error:   "invalid_state",
			Message: "Only pending_confirmation orders with reserved stock can be confirmed",
		})
		return
	}
	confirmReq := struct {
		ComplianceAck bool `json:"complianceAck"`
		Items         []struct {
			ProductID      string  `json:"productId"`
			Quantity       int     `json:"quantity"`
			UnitPrice      float64 `json:"unitPrice"`
			Specifications string  `json:"specifications"`
		} `json:"items"`
	}{}
	if c.Request.Body != nil {
		rawBody, readErr := io.ReadAll(c.Request.Body)
		if readErr != nil {
			c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
				Error:   "invalid_request",
				Message: "Failed to read request body",
			})
			return
		}
		trimmedBody := bytes.TrimSpace(rawBody)
		if len(trimmedBody) > 0 {
			if err := json.Unmarshal(trimmedBody, &confirmReq); err != nil {
				c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
					Error:   "invalid_request",
					Message: "Invalid JSON body",
				})
				return
			}
		}
	}

	// M1: Apply edited items if provided by the user
	if len(confirmReq.Items) > 0 {
		updatedItems := make(modelsOrder.OrderItemArray, 0, len(confirmReq.Items))
		var newSubtotal float64
		for _, it := range confirmReq.Items {
			pid := strings.TrimSpace(it.ProductID)
			if pid == "" || it.Quantity < 1 {
				continue
			}
			qty := it.Quantity
			price := it.UnitPrice
			if price < 0 {
				price = 0
			}
			updatedItems = append(updatedItems, modelsOrder.OrderItem{
				ProductID:      pid,
				Quantity:       qty,
				UnitPrice:      price,
				Specifications: strings.TrimSpace(it.Specifications),
			})
			newSubtotal += float64(qty) * price
		}
		if len(updatedItems) > 0 {
			order.Items = updatedItems
			order.Subtotal = newSubtotal
			order.TotalAmount = newSubtotal + order.TaxAmount + order.ShippingAmount
		}
	}
	if !order.ComplianceOfficialEvidence && !confirmReq.ComplianceAck {
		c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
			Error:   "compliance_ack_required",
			Message: "No official compliance evidence found. Explicit compliance acknowledgement is required before confirming this draft.",
			Details: gin.H{
				"requiresComplianceAck": true,
			},
		})
		return
	}

	now := time.Now()
	if err := h.services.Order.ConfirmPendingOrder(c.Request.Context(), orderID, now); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			latest, latestErr := h.services.Order.GetOrder(c.Request.Context(), orderID)
			if latestErr != nil {
				c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
					Error:   "invalid_state",
					Message: "Order state changed during confirmation, please refresh and retry",
				})
				return
			}
			if latest.UserID != userID {
				c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
					Error:   "forbidden",
					Message: "You do not have access to this order",
				})
				return
			}
			if latest.Status == "cancelled" && !latest.StockReserved {
				c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
					Error:   "order_expired",
					Message: "This AI draft expired and was cancelled. Please create a new draft order.",
				})
				return
			}
			c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
				Error:   "invalid_state",
				Message: "Order state changed during confirmation, please refresh and retry",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to confirm order",
		})
		return
	}

	order.Status = "pending"
	order.ConfirmedAt = &now
	order.UpdatedAt = now

	// Auto-create a trade transaction linked to this confirmed order
	if h.services.Trade != nil {
		trade := &modelsTrade.TradeTransaction{
			UserID:      order.UserID,
			OrderID:     &order.ID,
			Status:      modelsTrade.TradeStatusPending,
			Currency:    order.Currency,
			TotalAmount: order.TotalAmount,
			Terms:       "FOB",
		}
		if order.InquiryID != nil {
			trade.InquiryID = order.InquiryID
		}
		_ = h.services.Trade.CreateTransaction(c.Request.Context(), trade)
	}

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
	return fmt.Sprintf("CUS-%s-%s", time.Now().Format("20060102"), strings.ToUpper(utils.GenerateSlug()))
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	userID, ok := contextUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, modelsProduct.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not identified",
		})
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	if orderID == "" {
		utils.InvalidRequestResponse(c, "order id is required")
		return
	}

	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Order not found",
		})
		return
	}

	if order.UserID != userID {
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have access to this order",
		})
		return
	}

	// Only allow cancellation of pending orders
	if order.Status != "pending" {
		c.JSON(http.StatusConflict, modelsProduct.ErrorResponse{
			Error:   "invalid_state",
			Message: fmt.Sprintf("Cannot cancel order in '%s' status. Only pending orders can be cancelled.", order.Status),
		})
		return
	}

	// Release reserved stock and update status to cancelled
	order.Status = "cancelled"
	order.StockReserved = false
	now := time.Now()
	order.UpdatedAt = now

	if err := h.services.Order.ReleaseOrderStock(c.Request.Context(), order); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to cancel order",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
		"orderId": order.ID,
		"status":  "cancelled",
	})
}
