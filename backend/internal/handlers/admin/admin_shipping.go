package admin

import (
	"net/http"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type adminCreateShippingRateRequest struct {
	Destination   string  `json:"destination" binding:"required"`
	MinWeightKg   float64 `json:"min_weight_kg"`
	MaxWeightKg   float64 `json:"max_weight_kg"`
	BaseCost      float64 `json:"base_cost"`
	CostPerKg     float64 `json:"cost_per_kg"`
	Currency      string  `json:"currency"`
	Carrier       string  `json:"carrier"`
	EstimatedDays int     `json:"estimated_days"`
	IsActive      bool    `json:"is_active"`
}

type adminUpdateShippingRateRequest struct {
	Destination   *string  `json:"destination"`
	MinWeightKg   *float64 `json:"min_weight_kg"`
	MaxWeightKg   *float64 `json:"max_weight_kg"`
	BaseCost      *float64 `json:"base_cost"`
	CostPerKg     *float64 `json:"cost_per_kg"`
	Currency      *string  `json:"currency"`
	Carrier       *string  `json:"carrier"`
	EstimatedDays *int     `json:"estimated_days"`
	IsActive      *bool    `json:"is_active"`
}

// AdminGetShippingRates returns all shipping rates.
func (h *Handler) AdminGetShippingRates(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	rates, err := h.services.Shipping.GetAllRates(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}

// AdminCreateShippingRate creates a new shipping rate.
func (h *Handler) AdminCreateShippingRate(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req adminCreateShippingRateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	rate := modelsOrder.ShippingRate{
		Destination:   req.Destination,
		MinWeightKg:   req.MinWeightKg,
		MaxWeightKg:   req.MaxWeightKg,
		BaseCost:      req.BaseCost,
		CostPerKg:     req.CostPerKg,
		Currency:      req.Currency,
		Carrier:       req.Carrier,
		EstimatedDays: req.EstimatedDays,
		IsActive:      req.IsActive,
	}
	if err := h.services.Shipping.CreateRate(c.Request.Context(), &rate); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "create_failed")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "rate": rate})
}

// AdminUpdateShippingRate updates an existing shipping rate.
func (h *Handler) AdminUpdateShippingRate(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")

	existing, err := h.services.Shipping.GetRate(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	var req adminUpdateShippingRateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Destination != nil {
		existing.Destination = *req.Destination
	}
	if req.MinWeightKg != nil {
		existing.MinWeightKg = *req.MinWeightKg
	}
	if req.MaxWeightKg != nil {
		existing.MaxWeightKg = *req.MaxWeightKg
	}
	if req.BaseCost != nil {
		existing.BaseCost = *req.BaseCost
	}
	if req.CostPerKg != nil {
		existing.CostPerKg = *req.CostPerKg
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.Carrier != nil {
		existing.Carrier = *req.Carrier
	}
	if req.EstimatedDays != nil {
		existing.EstimatedDays = *req.EstimatedDays
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if err := h.services.Shipping.UpdateRate(c.Request.Context(), existing); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "rate": existing})
}

// AdminDeleteShippingRate deletes a shipping rate.
func (h *Handler) AdminDeleteShippingRate(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if err := h.services.Shipping.DeleteRate(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
