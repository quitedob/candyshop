package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminGetOrderPayments returns all payments for an order.
func (h *Handler) AdminGetOrderPayments(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	orderID := c.Param("id")
	payments, err := h.services.Payment.GetPaymentsByOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch payments")
		return
	}

	c.JSON(http.StatusOK, payments)
}

type adminCreatePaymentRequest struct {
	Amount    float64 `json:"amount" binding:"required"`
	Currency  string `json:"currency"`
	Method    string `json:"method" binding:"required"`
	Reference string `json:"reference"`
	ProofURL  string `json:"proofUrl"`
	Notes     string `json:"notes"`
}

// AdminCreatePayment creates a payment record for an order.
func (h *Handler) AdminCreatePayment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	orderID := c.Param("id")
	// Verify order exists
	if _, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
		return
	}

	var req adminCreatePaymentRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "USD"
	}

	payment := &modelsOrder.Payment{
		ID:        utils.GenerateID(),
		OrderID:   orderID,
		Amount:    req.Amount,
		Currency:  currency,
		Method:    strings.TrimSpace(req.Method),
		Status:    "pending",
		Reference: strings.TrimSpace(req.Reference),
		ProofURL:  strings.TrimSpace(req.ProofURL),
		Notes:     strings.TrimSpace(req.Notes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.services.Payment.CreatePayment(c.Request.Context(), payment); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create payment")
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// AdminConfirmPayment confirms a payment.
func (h *Handler) AdminConfirmPayment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	paymentID := c.Param("paymentId")
	adminID, _ := c.Get("userID")
	adminIDStr := ""
	if id, ok := adminID.(string); ok {
		adminIDStr = id
	}

	if err := h.services.Payment.ConfirmPayment(c.Request.Context(), paymentID, adminIDStr); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "payment_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, modelsProduct.ErrorResponse{
		Error:   "success",
		Message: "Payment confirmed successfully",
	})
}

// AdminRefundPayment refunds a payment.
func (h *Handler) AdminRefundPayment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	paymentID := c.Param("paymentId")
	if err := h.services.Payment.RefundPayment(c.Request.Context(), paymentID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "payment_error", err.Error())
		return
	}

	c.JSON(http.StatusOK, modelsProduct.ErrorResponse{
		Error:   "success",
		Message: "Payment refunded successfully",
	})
}
