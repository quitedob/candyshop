package customer

import (
	"net/http"
	"strconv"
	"strings"

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
	weightRaw := strings.TrimSpace(c.Query("weight_kg"))
	if weightRaw == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	// R2 A-7: previously the error was swallowed and malformed input silently
	// became 0.0, producing wrong/free shipping estimates.
	weightKg, err := strconv.ParseFloat(weightRaw, 64)
	if err != nil || weightKg <= 0 {
		response.InvalidResp(c, "invalid_weight")
		return
	}
	incoterms := strings.TrimSpace(c.Query("incoterms"))

	cost, currency, err := h.services.Shipping.CalculateShippingCostWithIncoterms(c.Request.Context(), destination, weightKg, incoterms)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"cost": cost, "currency": currency, "incoterms": incoterms})
}
