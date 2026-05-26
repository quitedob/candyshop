package customer

import (
	"candypro/api/internal/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CustomerGetInvoices returns all invoices for the current user orders.
func (h *Handler) CustomerGetInvoices(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	allInvoices, err := h.services.Invoice.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "invoice_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": allInvoices, "total": len(allInvoices)})
}

// CustomerGetInvoice returns a single invoice, verifying it belongs to the current user.
func (h *Handler) CustomerGetInvoice(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := c.Param("id")
	invoice, err := h.services.Invoice.GetInvoice(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "invoice_not_found")
		return
	}
	order, err := h.services.Order.GetOrder(c.Request.Context(), invoice.OrderID)
	if err != nil || order.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	c.JSON(http.StatusOK, invoice)
}
