package admin

import (
	modelsTrade "candypro/api/internal/models/trade"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminGetShipments returns paginated shipment records.
func (h *Handler) AdminGetShipments(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	page, limit := utils.ParsePagination(c, 20, 100)
	status := strings.TrimSpace(c.Query("status"))

	shipments, total, err := h.services.Shipment.ListShipments(c.Request.Context(), page, limit, status)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch shipments")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: shipments,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// AdminGetShipment returns a single shipment by ID.
func (h *Handler) AdminGetShipment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	shipment, err := h.services.Shipment.GetShipment(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Shipment not found")
		return
	}
	c.JSON(http.StatusOK, shipment)
}

type adminCreateShipmentRequest struct {
	TransactionID   uint    `json:"transactionId"`
	BillOfLadingNo  string  `json:"billOfLadingNo"`
	CarrierName     string  `json:"carrierName"`
	VesselFlight    string  `json:"vesselFlight"`
	PortOfLoading   string  `json:"portOfLoading"`
	PortOfDischarge string  `json:"portOfDischarge"`
	ETD             *string `json:"etd"`
	ETA             *string `json:"eta"`
	Status          string  `json:"status"`
}

// AdminCreateShipment creates a new shipment tracking record.
func (h *Handler) AdminCreateShipment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	var req adminCreateShipmentRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	shipment := &modelsTrade.ShipmentTracking{
		TransactionID:   req.TransactionID,
		BillOfLadingNo:  req.BillOfLadingNo,
		CarrierName:     req.CarrierName,
		VesselFlight:    req.VesselFlight,
		PortOfLoading:   req.PortOfLoading,
		PortOfDischarge: req.PortOfDischarge,
		Status:          req.Status,
	}
	if req.ETD != nil {
		if t, err := time.Parse(time.RFC3339, *req.ETD); err == nil {
			shipment.ETD = &t
		}
	}
	if req.ETA != nil {
		if t, err := time.Parse(time.RFC3339, *req.ETA); err == nil {
			shipment.ETA = &t
		}
	}

	if err := h.services.Shipment.CreateShipment(c.Request.Context(), shipment); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create shipment")
		return
	}
	c.JSON(http.StatusCreated, shipment)
}

// AdminUpdateShipment updates an existing shipment.
func (h *Handler) AdminUpdateShipment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	shipment, err := h.services.Shipment.GetShipment(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Shipment not found")
		return
	}

	var req adminCreateShipmentRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	if req.BillOfLadingNo != "" {
		shipment.BillOfLadingNo = req.BillOfLadingNo
	}
	if req.CarrierName != "" {
		shipment.CarrierName = req.CarrierName
	}
	if req.VesselFlight != "" {
		shipment.VesselFlight = req.VesselFlight
	}
	if req.PortOfLoading != "" {
		shipment.PortOfLoading = req.PortOfLoading
	}
	if req.PortOfDischarge != "" {
		shipment.PortOfDischarge = req.PortOfDischarge
	}
	if req.Status != "" {
		shipment.Status = req.Status
	}
	if req.ETD != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ETD); parseErr == nil {
			shipment.ETD = &t
		}
	}
	if req.ETA != nil {
		if t, parseErr := time.Parse(time.RFC3339, *req.ETA); parseErr == nil {
			shipment.ETA = &t
		}
	}

	if err := h.services.Shipment.UpdateShipment(c.Request.Context(), shipment); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update shipment")
		return
	}
	c.JSON(http.StatusOK, shipment)
}

// AdminDeleteShipment removes a shipment record.
func (h *Handler) AdminDeleteShipment(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id, err := parseUintParam(c, "id")
	if err != nil {
		utils.InvalidRequestResponse(c, "Invalid shipment ID")
		return
	}
	if _, fetchErr := h.services.Shipment.GetShipment(c.Request.Context(), id); fetchErr != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Shipment not found")
		return
	}
	if err := h.services.Shipment.DeleteShipment(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to delete shipment")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Shipment deleted", "id": id})
}

// parseUintParam parses a gin path parameter as uint64.
func parseUintParam(c *gin.Context, name string) (uint, error) {
	val, err := strconv.ParseUint(c.Param(name), 10, 64)
	return uint(val), err
}
