package customer

import (
modelsOrder "candypro/api/internal/models/order"
"candypro/api/internal/utils"
"net/http"

"github.com/gin-gonic/gin"
)

// CustomerGetInvoices returns all invoices for the current user orders.
func (h *Handler) CustomerGetInvoices(c *gin.Context) {
if h.services == nil {
utils.ServiceUnavailableResponse(c)
return
}
userID, ok := contextUserID(c)
if !ok {
utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
return
}
orders, err := h.services.Order.GetUserOrders(c.Request.Context(), userID, 1, 100)
if err != nil {
utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch invoices")
return
}
allInvoices := make([]modelsOrder.Invoice, 0)
if orderSlice, ok := orders.Data.([]modelsOrder.Order); ok {
for _, o := range orderSlice {
invoices, err := h.services.Invoice.GetByOrderID(c.Request.Context(), o.ID)
if err == nil {
allInvoices = append(allInvoices, invoices...)
}
}
}
c.JSON(http.StatusOK, gin.H{"data": allInvoices, "total": len(allInvoices)})
}

// CustomerGetInvoice returns a single invoice, verifying it belongs to the current user.
func (h *Handler) CustomerGetInvoice(c *gin.Context) {
if h.services == nil {
utils.ServiceUnavailableResponse(c)
return
}
userID, ok := contextUserID(c)
if !ok {
utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
return
}
id := c.Param("id")
invoice, err := h.services.Invoice.GetInvoice(c.Request.Context(), id)
if err != nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Invoice not found")
return
}
order, err := h.services.Order.GetOrder(c.Request.Context(), invoice.OrderID)
if err != nil || order.UserID != userID {
utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "You do not have access to this invoice")
return
}
c.JSON(http.StatusOK, invoice)
}