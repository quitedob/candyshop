package customer

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type customerUploadPaymentRequest struct {
	Amount    float64 `json:"amount" binding:"required"`
	Method    string  `json:"method" binding:"required"`
	Reference string  `json:"reference"`
	ProofURL  string  `json:"proofUrl"`
	Notes     string  `json:"notes"`
}

// CustomerUploadPaymentProof creates a payment record with proof of payment.
func (h *Handler) CustomerUploadPaymentProof(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	orderID := c.Param("id")
	userID, _ := c.Get("userID")
	userIDStr := ""
	if id, ok := userID.(string); ok {
		userIDStr = id
	}

	// Verify order belongs to this user
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
		return
	}
	if order.UserID != userIDStr {
		utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "You do not have access to this order")
		return
	}

	var req customerUploadPaymentRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	payment := &modelsOrder.Payment{
		ID:        utils.GenerateID(),
		OrderID:   orderID,
		Amount:    req.Amount,
		Currency:  order.Currency,
		Method:    strings.TrimSpace(req.Method),
		Status:    "pending",
		Reference: strings.TrimSpace(req.Reference),
		ProofURL:  strings.TrimSpace(req.ProofURL),
		Notes:     strings.TrimSpace(req.Notes),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.services.Payment.CreatePayment(c.Request.Context(), payment); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to submit payment proof")
		return
	}

	c.JSON(http.StatusCreated, payment)
}
