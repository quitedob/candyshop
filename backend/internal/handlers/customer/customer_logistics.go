package customer

import (
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CustomerGetShipmentTimeline returns the tracking event timeline for a specific shipment.
func (h *Handler) CustomerGetShipmentTimeline(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Verify trade ownership
	tradeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	trans, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(tradeID))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return
	}
	if trans.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	shipmentID, err := strconv.ParseUint(c.Param("shipmentId"), 10, 32)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	events, err := h.services.Logistics.GetShipmentTimeline(c.Request.Context(), uint(shipmentID))
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

// CustomerGetTradeTimeline returns all shipments and events for a trade transaction.
func (h *Handler) CustomerGetTradeTimeline(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	userID, ok := contextUserID(c)
	if !ok {
		response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	tradeID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	trans, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(tradeID))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return
	}
	if trans.UserID != userID {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	shipments, events, err := h.services.Logistics.GetTransactionTimeline(c.Request.Context(), uint(tradeID))
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"shipments": shipments,
		"events":    events,
	})
}
