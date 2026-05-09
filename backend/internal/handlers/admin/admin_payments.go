package admin

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
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
		utils.ServiceUnavailableResp(c)
		return
	}

	orderID := c.Param("id")
	payments, err := h.services.Payment.GetPaymentsByOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "payment_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, payments)
}

type adminCreatePaymentRequest struct {
	Amount    float64 `json:"amount" binding:"required"`
	Currency  string  `json:"currency"`
	Method    string  `json:"method" binding:"required"`
	Reference string  `json:"reference"`
	ProofURL  string  `json:"proofUrl"`
	Notes     string  `json:"notes"`
}

// AdminCreatePayment creates a payment record for an order.
func (h *Handler) AdminCreatePayment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	orderID := c.Param("id")
	ord, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	var req adminCreatePaymentRequest
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	if math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) || req.Amount < 0 {
		utils.InvalidResp(c, "invalid_request")
		return
	}

	currencyNorm, curErr := utils.NormalizeISOCurrency(req.Currency)
	if curErr != nil {
		utils.InvalidResp(c, "invalid_request")
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
		utils.InvalidResp(c, "invalid_request")
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

	// Atomic balance check + create within a transaction
	if err := h.services.Payment.CreatePaymentWithBalanceCheck(c.Request.Context(), ord.TotalAmount, payment); err != nil {
		if strings.Contains(err.Error(), "exceeds remaining balance") {
			utils.InvalidResp(c, "invalid_request")
			return
		}
		utils.ErrorResp(c, http.StatusInternalServerError, "payment_create_failed")
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// AdminConfirmPayment confirms a payment.
func (h *Handler) AdminConfirmPayment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil {
		utils.ErrorResp(c, http.StatusNotFound, "payment_not_found")
		return
	}
	if pay.OrderID != orderID {
		utils.InvalidResp(c, "invalid_request")
		return
	}

	adminID, _ := c.Get("userID")
	adminIDStr := ""
	if id, ok := adminID.(string); ok {
		adminIDStr = id
	}

	if err := h.services.Payment.ConfirmPayment(c.Request.Context(), paymentID, adminIDStr); err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}

	h.logPaymentActivity(c, "payment_confirm", orderID, paymentID, adminIDStr, pay.Amount)

	if h.services.Notification != nil {
		if ord, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err == nil {
			_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
				UserID:    ord.UserID,
				Type:      "order",
				Reference: orderID,
				Title:     "Payment Confirmed",
				Message:   fmt.Sprintf("Your payment of %.2f for order #%s has been confirmed.", pay.Amount, ord.OrderNumber),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   "success",
		"message": "Payment confirmed successfully",
	})
}

// AdminRefundPayment refunds a payment.
func (h *Handler) AdminRefundPayment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil {
		utils.ErrorResp(c, http.StatusNotFound, "payment_not_found")
		return
	}
	if pay.OrderID != orderID {
		utils.InvalidResp(c, "invalid_request")
		return
	}

	if err := h.services.Payment.RefundPayment(c.Request.Context(), paymentID); err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}

	adminID, _ := c.Get("userID")
	adminIDStr := ""
	if id, ok := adminID.(string); ok {
		adminIDStr = id
	}
	h.logPaymentActivity(c, "payment_refund", orderID, paymentID, adminIDStr, pay.Amount)

	if h.services.Notification != nil {
		if ord, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err == nil {
			_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
				UserID:    ord.UserID,
				Type:      "order",
				Reference: orderID,
				Title:     "Payment Refunded",
				Message:   fmt.Sprintf("Your payment of %.2f for order #%s has been refunded.", pay.Amount, ord.OrderNumber),
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   "success",
		"message": "Payment refunded successfully",
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
		utils.ServiceUnavailableResp(c)
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if _, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil || pay.OrderID != orderID {
		utils.ErrorResp(c, http.StatusNotFound, "payment_not_found")
		return
	}
	if strings.TrimSpace(pay.ProofURL) == "" {
		utils.ErrorResp(c, http.StatusNotFound, "proof_file_not_found")
		return
	}
	if h.cfg.Upload.StorageDriver == "s3" {
		presigned, err := h.storage.GetPresignedURL(c.Request.Context(), pay.ProofURL, 15*time.Minute)
		if err != nil {
			utils.ErrorResp(c, http.StatusInternalServerError, "payment_download_link_failed")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presigned)
		return
	}
	local, err := utils.LocalPathFromUploadURL(h.cfg.Upload.UploadPath, pay.ProofURL)
	if err != nil {
		utils.InvalidResp(c, "payment_invalid_proof_path")
		return
	}
	if _, statErr := os.Stat(local); statErr != nil {
		utils.ErrorResp(c, http.StatusNotFound, "file_not_found")
		return
	}
	c.File(local)
}
