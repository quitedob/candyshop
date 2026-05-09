package admin

import (
	"net/http"
	"strconv"
	"time"

	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"

	"github.com/gin-gonic/gin"
)

// AdminDispatchShipment dispatches a shipment (goods-issue from warehouse).
func (h *Handler) AdminDispatchShipment(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	operatorID, _ := c.Get("userID")
	opStr, _ := operatorID.(string)

	if err := h.services.Logistics.DispatchShipment(c.Request.Context(), uint(id), opStr); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "dispatch_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Shipment dispatched", "shipmentId": id})
}

// AdminAddTrackingEvent adds a tracking event to a shipment.
func (h *Handler) AdminAddTrackingEvent(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	var req struct {
		EventType   string `json:"eventType" binding:"required"`
		Location    string `json:"location"`
		Description string `json:"description"`
		EventTime   string `json:"eventTime"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	eventTime := time.Now()
	if req.EventTime != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, req.EventTime); parseErr == nil {
			eventTime = parsed
		}
	}
	operatorID, _ := c.Get("userID")
	opStr, _ := operatorID.(string)

	shipment, err := h.services.Logistics.AddTrackingEvent(c.Request.Context(), uint(id), req.EventType, req.Location, req.Description, opStr, eventTime)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "event_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, shipment)
}

// AdminConfirmDelivery confirms delivery with proof of delivery.
func (h *Handler) AdminConfirmDelivery(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	var req struct {
		DeliveryProofURL string `json:"deliveryProofUrl"`
		SignedBy         string `json:"signedBy"`
		DeliveredAt      string `json:"deliveredAt"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	deliveredAt := time.Now()
	if req.DeliveredAt != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, req.DeliveredAt); parseErr == nil {
			deliveredAt = parsed
		}
	}
	if err := h.services.Logistics.ConfirmDelivery(c.Request.Context(), uint(id), req.DeliveryProofURL, req.SignedBy, deliveredAt); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "confirm_failed", err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Delivery confirmed", "shipmentId": id})
}

// AdminGetShipmentTimeline returns the tracking event timeline for a shipment.
func (h *Handler) AdminGetShipmentTimeline(c *gin.Context) {
	if h.services == nil || h.services.Logistics == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	events, err := h.services.Logistics.GetShipmentTimeline(c.Request.Context(), uint(id))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch timeline")
		return
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: events,
	})
}
