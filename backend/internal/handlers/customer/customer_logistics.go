package customer

import (
	"net/http"
	"strconv"

	"candypro/api/internal/utils"

	"github.com/gin-gonic/gin"
)

// CustomerGetShipmentTimeline returns the tracking event timeline for a specific shipment.
func (h *Handler) CustomerGetShipmentTimeline(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized", "User not identified")
		return
	}

	// Verify trade ownership
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

	shipmentID, err := strconv.ParseUint(c.Param("shipmentId"), 10, 32)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	events, err := h.services.Logistics.GetShipmentTimeline(c.Request.Context(), uint(shipmentID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch timeline")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

// CustomerGetTradeTimeline returns all shipments and events for a trade transaction.
func (h *Handler) CustomerGetTradeTimeline(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
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

	shipments, events, err := h.services.Logistics.GetTransactionTimeline(c.Request.Context(), uint(tradeID))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch timeline")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"shipments": shipments,
		"events":    events,
	})
}
