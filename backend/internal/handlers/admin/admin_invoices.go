package admin

import (
	"errors"
	"fmt"

	modelsOrder "candypro/api/internal/models/order"
	modelsProduct "candypro/api/internal/models/product"
	orderSvc "candypro/api/internal/services/order"
	"candypro/api/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminGetInvoices returns paginated invoices.
func (h *Handler) AdminGetInvoices(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	page, limit := utils.ParsePagination(c, 20, 100)
	status := strings.TrimSpace(c.Query("status"))

	invoices, total, err := h.services.Invoice.ListInvoices(c.Request.Context(), page, limit, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch invoices")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: invoices,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// AdminGetInvoice returns a single invoice by ID.
func (h *Handler) AdminGetInvoice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id := c.Param("id")
	invoice, err := h.services.Invoice.GetInvoice(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Invoice not found")
		return
	}
	c.JSON(http.StatusOK, invoice)
}

type adminCreateInvoiceRequest struct {
	OrderID   string   `json:"orderId"`
	TradeID   *uint    `json:"tradeId"`
	Type      string   `json:"type"`
	Amount    float64  `json:"amount" binding:"required"`
	TaxAmount float64  `json:"taxAmount"`
	Currency  string   `json:"currency"`
	DueDate   *string  `json:"dueDate"`
	Notes     string   `json:"notes"`
	Items     string   `json:"items"` // JSON string for line items
}

// AdminCreateInvoice creates a new invoice.
func (h *Handler) AdminCreateInvoice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	var req adminCreateInvoiceRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	if req.Amount < 0 || req.TaxAmount < 0 {
		utils.InvalidRequestResponse(c, "amount and taxAmount cannot be negative")
		return
	}

	invoice := &modelsOrder.Invoice{
		OrderID:   req.OrderID,
		TradeID:   req.TradeID,
		Type:      req.Type,
		Amount:    req.Amount,
		TaxAmount: req.TaxAmount,
		Currency:  req.Currency,
		Notes:     req.Notes,
		Items:     req.Items,
	}
	if req.DueDate != nil {
		parsed, parseErr := parseOptionalNullableTime(*req.DueDate, "dueDate")
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
				Error:   "invalid_request",
				Message: parseErr.Error(),
			})
			return
		}
		invoice.DueDate = parsed
	}

	if err := h.services.Invoice.CreateInvoice(c.Request.Context(), invoice); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create invoice")
		return
	}
	c.JSON(http.StatusCreated, invoice)
}

// AdminUpdateInvoice updates an existing invoice.
func (h *Handler) AdminUpdateInvoice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id := c.Param("id")
	invoice, err := h.services.Invoice.GetInvoice(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Invoice not found")
		return
	}

	var req struct {
		Amount           *float64 `json:"amount"`
		TaxAmount        *float64 `json:"taxAmount"`
		Currency         string   `json:"currency"`
		Notes            string   `json:"notes"`
		Items            string   `json:"items"`
		DueDate          *string  `json:"dueDate"`
		AdjustmentReason string   `json:"adjustmentReason"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	before := *invoice
	after := *invoice

	// Update amount: allow any non-negative value including 0
	if req.Amount != nil {
		if *req.Amount < 0 {
			utils.InvalidRequestResponse(c, "amount cannot be negative")
			return
		}
		after.Amount = *req.Amount
	}
	if req.TaxAmount != nil {
		if *req.TaxAmount < 0 {
			utils.InvalidRequestResponse(c, "taxAmount cannot be negative")
			return
		}
		after.TaxAmount = *req.TaxAmount
	}
	if req.Currency != "" {
		after.Currency = req.Currency
	}
	// Notes/Items: empty string = clear (intentional)
	if req.Notes != "" || (req.Notes == "" && len(c.Request.URL.Query()) > 0) {
		after.Notes = req.Notes
	}
	if req.Items != "" {
		after.Items = req.Items
	}
	// DueDate: omitted = unchanged, "" = clear, RFC3339 = set, invalid = 400
	if req.DueDate != nil {
		parsed, parseErr := parseOptionalNullableTime(*req.DueDate, "dueDate")
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
				Error:   "invalid_request",
				Message: parseErr.Error(),
			})
			return
		}
		after.DueDate = parsed
	}

	actor := ""
	if v, ok := c.Get("userID"); ok {
		if s, ok := v.(string); ok {
			actor = strings.TrimSpace(s)
		}
	}
	if err := h.services.Invoice.ValidateAndPersistInvoiceUpdate(c.Request.Context(), &before, &after, actor, strings.TrimSpace(req.AdjustmentReason)); err != nil {
		if errors.Is(err, orderSvc.ErrAdjustmentReasonRequired) {
			c.JSON(http.StatusUnprocessableEntity, modelsProduct.ErrorResponse{
				Error:   "adjustment_reason_required",
				Message: err.Error(),
			})
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update invoice")
		return
	}
	c.JSON(http.StatusOK, &after)
}

// AdminSendInvoice marks an invoice as sent.
func (h *Handler) AdminSendInvoice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id := c.Param("id")
	invoice, err := h.services.Invoice.SendInvoice(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	c.JSON(http.StatusOK, invoice)
}

// AdminDeleteInvoice removes a draft invoice.
func (h *Handler) AdminDeleteInvoice(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id := c.Param("id")
	if err := h.services.Invoice.DeleteInvoice(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invoice deleted", "id": id})
}

// AdminGetInvoiceStats returns invoice counts by status.
func (h *Handler) AdminGetInvoiceStats(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	stats, err := h.services.Invoice.GetStats(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch invoice stats")
		return
	}
	c.JSON(http.StatusOK, stats)
}

// parseOptionalNullableTime handles the tri-state date semantics:
//   - empty string → nil (clear the field)
//   - valid RFC3339 → parsed time
//   - invalid non-empty string → error
func parseOptionalNullableTime(raw string, fieldName string) (*time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return nil, fmt.Errorf("%s must be a valid RFC3339 timestamp or empty string, got: %q", fieldName, raw)
	}
	return &t, nil
}
