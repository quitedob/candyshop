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

	invoice := &modelsOrder.Invoice{
		OrderID:   req.OrderID,
		Type:      req.Type,
		Amount:    req.Amount,
		TaxAmount: req.TaxAmount,
		Currency:  req.Currency,
		Notes:     req.Notes,
		Items:     req.Items,
	}
	if req.DueDate != nil {
		if t, err := time.Parse(time.RFC3339, *req.DueDate); err == nil {
			invoice.DueDate = &t
		}
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
		Amount    *float64 `json:"amount"`
		TaxAmount *float64 `json:"taxAmount"`
		Currency  string   `json:"currency"`
		Notes     string   `json:"notes"`
		Items     string   `json:"items"`
		DueDate   *string  `json:"dueDate"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	// N-12: Only update fields that are explicitly provided (use pointers)
	if req.Amount != nil && *req.Amount > 0 {
		invoice.Amount = *req.Amount
	}
	if req.TaxAmount != nil {
		invoice.TaxAmount = *req.TaxAmount
	}
	if req.Currency != "" {
		invoice.Currency = req.Currency
	}
	if req.Notes != "" {
		invoice.Notes = req.Notes
	}
	if req.Items != "" {
		invoice.Items = req.Items
	}
	if req.DueDate != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.DueDate); parseErr == nil {
			invoice.DueDate = &t
		}
	}

	if err := h.services.Invoice.UpdateInvoice(c.Request.Context(), invoice); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update invoice")
		return
	}
	c.JSON(http.StatusOK, invoice)
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
