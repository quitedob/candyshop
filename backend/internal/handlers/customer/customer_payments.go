package customer

import (
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

// CustomerGetOrderPayments returns all payments for the customer's order.
func (h *Handler) CustomerGetOrderPayments(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	orderID := c.Param("id")
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	if order.UserID != userID {
		utils.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	payments, err := h.services.Payment.GetPaymentsByOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "payment_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, payments)
}

// CustomerUploadPaymentProof creates payment record: supports multipart (method + proof file) or JSON (amount/method/proofUrl etc).
func (h *Handler) CustomerUploadPaymentProof(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}

	orderID := c.Param("id")
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "order_not_found")
		return
	}
	if order.UserID != userID {
		utils.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	payments, payErr := h.services.Payment.GetPaymentsByOrder(c.Request.Context(), orderID)
	if payErr != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "payment_fetch_failed")
		return
	}

	remaining := paymentRemainingDue(order, payments)
	if remaining <= utils.MoneyEpsilon {
		utils.ErrorResp(c, http.StatusBadRequest, "payment_no_balance")
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
				utils.InvalidResp(c, "invalid_request")
				return
			}
			amount = a
			amountSet = true
		}
		reference = strings.TrimSpace(c.PostForm("reference"))
		notes = strings.TrimSpace(c.PostForm("notes"))

		fileHeader, ferr := c.FormFile("proof")
		if ferr != nil || fileHeader == nil {
			utils.InvalidResp(c, "payment_proof_required")
			return
		}
		file, oerr := fileHeader.Open()
		if oerr != nil {
			utils.InvalidResp(c, "file_open_failed")
			return
		}
		defer file.Close()
		saved, serr := h.storage.Upload(c.Request.Context(), file, storage.UploadOptions{
			Folder:   "payment-proofs",
			FileName: fileHeader.Filename,
		})
		if serr != nil {
			utils.ErrorResp(c, http.StatusInternalServerError, "payment_upload_failed")
			return
		}
		proofURL = saved
	} else {
		var req customerUploadPaymentRequest
		if !utils.BindJSONOrInvalid(c, &req) {
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
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if !amountSet {
		amount = remaining
	}
	if amount <= 0 {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if amount > remaining+1e-4 {
		utils.InvalidResp(c, "invalid_request")
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
		utils.ErrorResp(c, http.StatusInternalServerError, "payment_create_failed")
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// CustomerDownloadPaymentProofFile returns payment proof file via authenticated stream.
func (h *Handler) CustomerDownloadPaymentProofFile(c *gin.Context) {
	if h.services == nil || h.cfg == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID := strings.TrimSpace(c.Param("id"))
	paymentID := strings.TrimSpace(c.Param("paymentId"))
	if orderID == "" || paymentID == "" {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	order, err := h.services.Order.GetOrder(c.Request.Context(), orderID)
	if err != nil || order.UserID != userID {
		utils.ErrorResp(c, http.StatusForbidden, "forbidden")
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
			utils.ErrorResp(c, http.StatusInternalServerError, "internal_error")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, presigned)
		return
	}
	local, err := utils.LocalPathFromUploadURL(h.cfg.Upload.UploadPath, pay.ProofURL)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	if _, statErr := os.Stat(local); statErr != nil {
		utils.ErrorResp(c, http.StatusNotFound, "file_not_found")
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
