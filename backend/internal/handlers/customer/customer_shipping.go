package customer

import (
	"net/http"
	"strconv"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// CustomerGetShippingRates returns active shipping rates for a destination.
func (h *Handler) CustomerGetShippingRates(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	destination := c.Query("destination")
	if destination == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	rates, err := h.services.Shipping.GetDestinationRates(c.Request.Context(), destination)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}

// CustomerGetShippingEstimate returns a shipping cost estimate.
func (h *Handler) CustomerGetShippingEstimate(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	destination := c.Query("destination")
	if destination == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	weightKg, _ := strconv.ParseFloat(c.Query("weight_kg"), 64)

	cost, currency, err := h.services.Shipping.CalculateShippingCost(c.Request.Context(), destination, weightKg)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"cost": cost, "currency": currency})
}
