package public

import (
	"github.com/gin-gonic/gin"
)

// GetExchangeRates returns configured exchange rates (base: USD)
// @Summary Get exchange rates
// @Tags exchange-rates
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /exchange-rates [get]
func (h *Handler) GetExchangeRates(c *gin.Context) {
	rates := h.cfg.ExchangeRates
	if rates == nil {
		rates = make(map[string]float64)
	}
	c.JSON(200, gin.H{
		"base":       "USD",
		"rates":      rates,
		"updated":    "manual",
		"disclaimer": "Reference rates only. Actual transaction prices may vary.",
	})
}
