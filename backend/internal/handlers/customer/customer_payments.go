package customer

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/storage"
	"candypro/api/internal/utils"

	"github.com/gin-gonic/gin"
)

type customerUploadPaymentRequest struct {
	Amount    float64 `json:"amount" binding:"required"`
	Method    string  `json:"method" binding:"required"`
	Reference string  `json:"reference"`
	ProofURL  string  `json:"proofUrl"`
	Notes     string  `json:"notes"`
}

// paymentRemainingDue 计算订单剩余应付：总额减去已确认与待审核的付款申报（防止重复超额申报）。
func paymentRemainingDue(order *modelsOrder.Order, payments []modelsOrder.Payment) float64 {
	if order == nil {
		return 0
	}
	allocated := 0.0
	for _, p := range payments {
		switch p.Status {
		case "confirmed", "pending":
			allocated += p.Amount
		}
	}
	rem := order.TotalAmount - allocated
	if rem < 0 {
		return 0
	}
	return rem
}

func sanitizeProofURL(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	low := strings.ToLower(s)
	if strings.HasPrefix(low, "javascript:") || strings.HasPrefix(low, "data:") || strings.HasPrefix(low, "vbscript:") {
		return ""
	}
	return s
}

// CustomerUploadPaymentProof 创建付款记录：支持 multipart（与前端一致：method + proof 文件）或 JSON（amount/method/proofUrl 等）。
func (h *Handler) CustomerUploadPaymentProof(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	orderID := c.Param("id")
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}

	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Order not found")
		return
	}
	if order.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "You do not have access to this order")
		return
	}

	payments, payErr := h.services.Payment.GetPaymentsByOrder(c.Request.Context(), orderID)
	if payErr != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to load payments")
		return
	}

	remaining := paymentRemainingDue(order, payments)
	if remaining <= utils.MoneyEpsilon {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", "No remaining balance to pay for this order")
		return
	}

	ct := c.GetHeader("Content-Type")
	var method string
	var amount float64
	var reference string
	var proofURL string
	var notes string
	var amountSet bool

	if strings.Contains(strings.ToLower(ct), "multipart/form-data") {
		method = strings.TrimSpace(c.PostForm("method"))
		if v := c.PostForm("amount"); v != "" {
			a, err := strconv.ParseFloat(v, 64)
			if err != nil || math.IsNaN(a) || math.IsInf(a, 0) {
				utils.InvalidRequestResponse(c, "Invalid amount")
				return
			}
			amount = a
			amountSet = true
		}
		reference = strings.TrimSpace(c.PostForm("reference"))
		notes = strings.TrimSpace(c.PostForm("notes"))

		fileHeader, ferr := c.FormFile("proof")
		if ferr != nil || fileHeader == nil {
			utils.InvalidRequestResponse(c, "Payment proof file is required")
			return
		}
		file, oerr := fileHeader.Open()
		if oerr != nil {
			utils.InvalidRequestResponse(c, "Failed to open payment proof file")
			return
		}
		defer file.Close()
		saved, serr := h.storage.Upload(c.Request.Context(), file, storage.UploadOptions{
			Folder:   "payment-proofs",
			FileName: fileHeader.Filename,
		})
		if serr != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to save payment proof")
			return
		}
		proofURL = saved
	} else {
		var req customerUploadPaymentRequest
		if !utils.BindJSONOrInvalidRequest(c, &req) {
			return
		}
		method = strings.TrimSpace(req.Method)
		amount = req.Amount
		amountSet = true
		reference = strings.TrimSpace(req.Reference)
		proofURL = sanitizeProofURL(req.ProofURL)
		notes = strings.TrimSpace(req.Notes)
	}

	if method == "" {
		utils.InvalidRequestResponse(c, "Payment method is required")
		return
	}
	if !amountSet {
		amount = remaining
	}
	if amount <= 0 {
		utils.InvalidRequestResponse(c, "Amount must be greater than zero")
		return
	}
	if amount > remaining+1e-4 {
		utils.InvalidRequestResponse(c, fmt.Sprintf("Amount exceeds remaining balance (%.2f)", remaining))
		return
	}
	payment := &modelsOrder.Payment{
		ID:        utils.GenerateID(),
		OrderID:   orderID,
		Amount:    amount,
		Currency:  order.Currency,
		Method:    method,
		Status:    "pending",
		Reference: reference,
		ProofURL:  proofURL,
		Notes:     notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.services.Payment.CreatePayment(c.Request.Context(), payment); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to submit payment proof")
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// CustomerDownloadPaymentProofFile 通过认证流式返回本订单付款凭证文件（禁止直链 /uploads/payment-proofs）。
func (h *Handler) CustomerDownloadPaymentProofFile(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.InvalidRequestResponse(c, "order id and payment id are required")
		return
	}
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil || order.UserID != userID {
		utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "You do not have access to this order")
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

func isAllowedPaymentProofType(ct string) bool {
	ct = strings.TrimSpace(strings.ToLower(ct))
	switch {
	case strings.HasPrefix(ct, "application/pdf"):
		return true
	case strings.HasPrefix(ct, "image/jpeg"):
		return true
	case strings.HasPrefix(ct, "image/png"):
		return true
	default:
		return false
	}
}
