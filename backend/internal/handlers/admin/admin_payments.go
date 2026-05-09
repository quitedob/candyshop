package admin

import (
	"encoding/json"
	"log"
	"math"
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"os"
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
	ord, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
		return
	}

	var req adminCreatePaymentRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	if math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) || req.Amount < 0 {
		utils.InvalidRequestResponse(c, "amount must be a non-negative finite number")
		return
	}

	currencyNorm, curErr := utils.NormalizeISOCurrency(req.Currency)
	if curErr != nil {
		utils.InvalidRequestResponse(c, curErr.Error())
		return
	}
	currency := currencyNorm
	if currency == "" {
		currency = "USD"
	}
	orderCur, _ := utils.NormalizeISOCurrency(ord.Currency)
	if orderCur == "" {
		orderCur = "USD"
	}
	if currency != orderCur {
		utils.InvalidRequestResponse(c, "payment currency must match the order currency")
		return
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

	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Order id and payment id are required")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Payment not found")
		return
	}
	if pay.OrderID != orderID {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Payment does not belong to this order")
		return
	}

	adminID, _ := c.Get("userID")
	adminIDStr := ""
	if id, ok := adminID.(string); ok {
		adminIDStr = id
	}

	if err := h.services.Payment.ConfirmPayment(c.Request.Context(), paymentID, adminIDStr); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "payment_error", err.Error())
		return
	}

	h.logPaymentActivity(c, "payment_confirm", orderID, paymentID, adminIDStr, pay.Amount)

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

	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Order id and payment id are required")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Payment not found")
		return
	}
	if pay.OrderID != orderID {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Payment does not belong to this order")
		return
	}

	if err := h.services.Payment.RefundPayment(c.Request.Context(), paymentID); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "payment_error", err.Error())
		return
	}

	adminID, _ := c.Get("userID")
	adminIDStr := ""
	if id, ok := adminID.(string); ok {
		adminIDStr = id
	}
	h.logPaymentActivity(c, "payment_refund", orderID, paymentID, adminIDStr, pay.Amount)

	c.JSON(http.StatusOK, modelsProduct.ErrorResponse{
		Error:   "success",
		Message: "Payment refunded successfully",
	})
}

func (h *Handler) logPaymentActivity(c *gin.Context, action, orderID, paymentID, adminID string, amount float64) {
	if h.services == nil || h.services.ActivityLog == nil {
		return
	}
	if strings.TrimSpace(adminID) == "" {
		log.Printf("payment_audit_skipped: empty admin user id action=%s orderId=%s paymentId=%s amount=%v ip=%s",
			action, orderID, paymentID, amount, c.ClientIP())
		return
	}
	details, _ := json.Marshal(map[string]interface{}{
		"orderId":   orderID,
		"paymentId": paymentID,
		"amount":    amount,
	})
	_ = h.services.ActivityLog.LogActivity(c.Request.Context(), &modelsCommon.ActivityLog{
		ID:         utils.GenerateID(),
		UserID:     &adminID,
		Action:     action,
		EntityType: "payment",
		EntityID:   paymentID,
		Details:    string(details),
		IPAddress:  c.ClientIP(),
		UserAgent:  c.GetHeader("User-Agent"),
		CreatedAt:  time.Now(),
	})
}

// AdminDownloadPaymentProofFile 管理员下载订单付款凭证文件。
func (h *Handler) AdminDownloadPaymentProofFile(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Order id and payment id are required")
		return
	}
	if _, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil || pay.OrderID != orderID {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Payment not found")
		return
	}
	if strings.TrimSpace(pay.ProofURL) == "" {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "No proof file for this payment")
		return
	}
	if h.cfg.Upload.StorageDriver == "s3" {
		presigned, err := h.storage.GetPresignedURL(c.Request.Context(), pay.ProofURL, 15*time.Minute)
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to generate download link")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presigned)
		return
	}
	local, err := utils.LocalPathFromUploadURL(h.cfg.Upload.UploadPath, pay.ProofURL)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid proof path")
		return
	}
	if _, statErr := os.Stat(local); statErr != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Proof file not found on server")
		return
	}
	c.File(local)
}
