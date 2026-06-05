package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/money"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/uploadpath"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminGetOrderPayments returns all payments for an order.
func (h *Handler) AdminGetOrderPayments(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	orderID := c.Param("id")
	payments, err := h.services.Payment.GetPaymentsByOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "payment_fetch_failed")
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
		response.ServiceUnavailableResp(c)
		return
	}

	orderID := c.Param("id")
	ord, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}

	var req adminCreatePaymentRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if math.IsNaN(req.Amount) || math.IsInf(req.Amount, 0) || req.Amount < 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}

	currencyNorm, curErr := money.NormalizeISOCurrency(req.Currency)
	if curErr != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	currency := currencyNorm
	if currency == "" {
		currency = "USD"
	}
	orderCur, _ := money.NormalizeISOCurrency(ord.Currency)
	if orderCur == "" {
		orderCur = "USD"
	}
	if currency != orderCur {
		response.InvalidResp(c, "invalid_request")
		return
	}

	method := strings.TrimSpace(req.Method)
	if !modelsOrder.ValidPaymentMethods[method] {
		response.InvalidResp(c, "invalid_request")
		return
	}

	payment := &modelsOrder.Payment{
		ID:        crypto.GenerateID(),
		OrderID:   orderID,
		Amount:    req.Amount,
		Currency:  currency,
		Method:    method,
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
			response.InvalidResp(c, "invalid_request")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "payment_create_failed")
		return
	}

	adminIDStr := c.GetString("userID")
	h.logPaymentActivity(c, "payment_create", orderID, payment.ID, adminIDStr, payment.Amount)

	c.JSON(http.StatusCreated, payment)
}

// AdminConfirmPayment confirms a payment.
func (h *Handler) AdminConfirmPayment(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil {
		response.ErrorResp(c, http.StatusNotFound, "payment_not_found")
		return
	}
	if pay.OrderID != orderID {
		response.InvalidResp(c, "invalid_request")
		return
	}

	adminIDStr := c.GetString("userID")

	var confirmErr error
	if h.services.GatewayPayment != nil && pay.GatewayTransactionID != nil && strings.TrimSpace(*pay.GatewayTransactionID) != "" {
		confirmErr = h.services.GatewayPayment.CaptureAndConfirm(c.Request.Context(), pay, adminIDStr)
	} else {
		confirmErr = h.services.Payment.ConfirmPayment(c.Request.Context(), paymentID, adminIDStr)
	}
	if confirmErr != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	h.logPaymentActivity(c, "payment_confirm", orderID, paymentID, adminIDStr, pay.Amount)

	if h.cfg != nil && h.cfg.Order.AutoInvoiceOnPayment && h.services.Invoice != nil {
		existing, _ := h.services.Invoice.GetByOrderID(c.Request.Context(), orderID)
		if len(existing) == 0 {
			if inv, invErr := h.services.Invoice.CreateInvoiceFromOrder(c.Request.Context(), orderID, adminIDStr); invErr == nil && inv != nil && h.services.ActivityLog != nil {
				details, _ := json.Marshal(map[string]interface{}{"orderId": orderID, "invoiceId": inv.ID, "amount": inv.Amount})
				uid := adminIDStr
				_ = h.services.ActivityLog.LogActivity(c.Request.Context(), &modelsCommon.ActivityLog{
					ID: crypto.GenerateID(), UserID: &uid, Action: "invoice_auto_created",
					EntityType: "order", EntityID: orderID, Details: string(details),
					IPAddress: c.ClientIP(), UserAgent: c.GetHeader("User-Agent"), CreatedAt: time.Now(),
				})
			}
		}
	}

	if h.services.Notification != nil {
		if ord, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err == nil {
			_ = h.services.Notification.Create(c.Request.Context(), &modelsCommon.Notification{
				UserID:    ord.UserID,
				Type:      "order",
				Reference: orderID,
				Title:     "Payment Confirmed",
				Message:   fmt.Sprintf("Your payment of %.2f for order #%s has been confirmed.", pay.Amount, ord.OrderNumber),
			})
			h.emitLifecycleEvent(c, modelsOrder.WebhookEventPaymentConfirmed, orderID, gin.H{
				"orderId": orderID, "paymentId": paymentID, "amount": pay.Amount,
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
		response.ServiceUnavailableResp(c)
		return
	}

	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil {
		response.ErrorResp(c, http.StatusNotFound, "payment_not_found")
		return
	}
	if pay.OrderID != orderID {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if h.services.GatewayPayment != nil && pay.GatewayTransactionID != nil {
		if err := h.services.GatewayPayment.RefundGateway(c.Request.Context(), pay); err != nil {
			response.InvalidResp(c, "invalid_request")
			return
		}
	} else if err := h.services.Payment.RefundPayment(c.Request.Context(), paymentID); err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	adminIDStr := c.GetString("userID")
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
		log.Printf("payment_audit: empty admin user id action=%s orderId=%s paymentId=%s amount=%v ip=%s",
			action, orderID, paymentID, amount, c.ClientIP())
	}
	details, _ := json.Marshal(map[string]interface{}{
		"orderId":   orderID,
		"paymentId": paymentID,
		"amount":    amount,
	})
	var uid *string
	if strings.TrimSpace(adminID) != "" {
		id := adminID
		uid = &id
	}
	now := time.Now()
	base := &modelsCommon.ActivityLog{
		UserID:    uid,
		Action:    action,
		Details:   string(details),
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		CreatedAt: now,
	}
	// 付款实体日志（全局审计）
	_ = h.services.ActivityLog.LogActivity(c.Request.Context(), &modelsCommon.ActivityLog{
		ID: crypto.GenerateID(), UserID: base.UserID, Action: base.Action,
		EntityType: "payment", EntityID: paymentID, Details: base.Details,
		IPAddress: base.IPAddress, UserAgent: base.UserAgent, CreatedAt: base.CreatedAt,
	})
	// 订单实体日志（订单详情时间线）
	_ = h.services.ActivityLog.LogActivity(c.Request.Context(), &modelsCommon.ActivityLog{
		ID: crypto.GenerateID(), UserID: base.UserID, Action: base.Action,
		EntityType: "order", EntityID: orderID, Details: base.Details,
		IPAddress: base.IPAddress, UserAgent: base.UserAgent, CreatedAt: base.CreatedAt,
	})
}

// AdminDownloadPaymentProofFile 管理员下载订单付款凭证文件。
func (h *Handler) AdminDownloadPaymentProofFile(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if _, err := h.services.Order.GetOrder(c.Request.Context(), orderID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	pay, err := h.services.Payment.GetPayment(c.Request.Context(), paymentID)
	if err != nil || pay == nil || pay.OrderID != orderID {
		response.ErrorResp(c, http.StatusNotFound, "payment_not_found")
		return
	}
	if strings.TrimSpace(pay.ProofURL) == "" {
		response.ErrorResp(c, http.StatusNotFound, "proof_file_not_found")
		return
	}
	if h.cfg.Upload.StorageDriver == "s3" || h.cfg.Upload.StorageDriver == "oss" {
		presigned, err := h.storage.GetPresignedURL(c.Request.Context(), pay.ProofURL, 15*time.Minute)
		if err != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "payment_download_link_failed")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presigned)
		return
	}
	local, err := uploadpath.LocalPathFromUploadURL(h.cfg.Upload.UploadPath, pay.ProofURL)
	if err != nil {
		response.InvalidResp(c, "payment_invalid_proof_path")
		return
	}
	if _, statErr := os.Stat(local); statErr != nil {
		response.ErrorResp(c, http.StatusNotFound, "file_not_found")
		return
	}
	c.File(local)
}
