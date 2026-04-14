package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetFinancialOverview returns a summary of financial metrics.
func (h *Handler) GetFinancialOverview(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	// Invoice stats by status
	invoiceStats, err := h.services.Invoice.GetStats(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch invoice stats")
		return
	}

	// Sum overdue invoices
	overdueAmount, err := h.services.Invoice.SumOverdue(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch overdue amount")
		return
	}

	// Confirmed payment count
	confirmedPayments, err := h.services.Payment.CountByStatus(c.Request.Context(), "confirmed")
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch payment counts")
		return
	}

	// Total revenue from orders
	totalRevenue, err := h.services.Order.SumSales(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch total revenue")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"invoiceStats":      invoiceStats,
		"overdueAmount":     overdueAmount,
		"confirmedPayments": confirmedPayments,
		"totalRevenue":      totalRevenue,
	})
}

// GetOutstandingInvoices returns paginated overdue/sent invoices.
func (h *Handler) GetOutstandingInvoices(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	page, limit := utils.ParsePagination(c, 20, 100)
	status := strings.TrimSpace(c.Query("status"))
	if status == "" {
		status = "sent"
	}

	invoices, total, err := h.services.Invoice.FindByStatus(c.Request.Context(), status, page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch outstanding invoices")
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

// GetPaymentBreakdown returns payment counts grouped by status.
func (h *Handler) GetPaymentBreakdown(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	breakdown, err := h.services.Payment.StatusBreakdown(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch payment breakdown")
		return
	}
	c.JSON(http.StatusOK, gin.H{"breakdown": breakdown})
}
