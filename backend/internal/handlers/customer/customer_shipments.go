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
		utils.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	tradeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	trans, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(tradeID))
	if err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return
	}
	if trans.UserID != userID {
		utils.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	shipments, err := h.services.Shipment.GetByTransactionID(c.Request.Context(), uint(tradeID))
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": shipments})
}
