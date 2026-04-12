package customer

import (
"candypro/api/internal/utils"
"net/http"
"strconv"

"github.com/gin-gonic/gin"
)

// CustomerGetTradeShipments returns shipment tracking for a trade owned by the user.
func (h *Handler) CustomerGetTradeShipments(c *gin.Context) {
if h.services == nil {
utils.ServiceUnavailableResponse(c)
return
}
userID, ok := contextUserID(c)
if !ok {
utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
return
}
tradeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
if err != nil {
utils.InvalidRequestResponse(c, "Invalid trade ID")
return
}
trans, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(tradeID))
if err != nil {
utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Trade transaction not found")
return
}
if trans.UserID != userID {
utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "You do not have access to this trade")
return
}
shipments, err := h.services.Shipment.GetByTransactionID(c.Request.Context(), uint(tradeID))
if err != nil {
utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch shipments")
return
}
c.JSON(http.StatusOK, gin.H{"data": shipments})
}