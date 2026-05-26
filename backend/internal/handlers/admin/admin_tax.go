package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type adminTaxRateRequest struct {
	Country  string  `json:"country" binding:"required"`
	Region   string  `json:"region"`
	Rate     float64 `json:"rate" binding:"required"`
	Name     string  `json:"name"`
	IsActive *bool   `json:"isActive"`
}

// AdminGetTaxRates handles GET /api/v1/admin/tax-rates
func (h *Handler) AdminGetTaxRates(c *gin.Context) {
	if h.services == nil || h.services.Tax == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	rates, err := h.services.Tax.GetAllRates(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "tax_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, rates)
}

// AdminCreateTaxRate handles POST /api/v1/admin/tax-rates
func (h *Handler) AdminCreateTaxRate(c *gin.Context) {
	if h.services == nil || h.services.Tax == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req adminTaxRateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	rate := &modelsOrder.TaxRate{
		Country:  strings.ToUpper(strings.TrimSpace(req.Country)),
		Region:   strings.TrimSpace(req.Region),
		Rate:     req.Rate,
		Name:     strings.TrimSpace(req.Name),
		IsActive: true,
	}
	if req.IsActive != nil {
		rate.IsActive = *req.IsActive
	}
	if err := h.services.Tax.CreateRate(c.Request.Context(), rate); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "tax_create_failed")
		return
	}
	c.JSON(http.StatusCreated, rate)
}

// AdminUpdateTaxRate handles PUT /api/v1/admin/tax-rates/:id
func (h *Handler) AdminUpdateTaxRate(c *gin.Context) {
	if h.services == nil || h.services.Tax == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	existing, err := h.services.Tax.GetRate(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "tax_not_found")
		return
	}
	var req adminTaxRateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	existing.Country = strings.ToUpper(strings.TrimSpace(req.Country))
	existing.Region = strings.TrimSpace(req.Region)
	existing.Rate = req.Rate
	existing.Name = strings.TrimSpace(req.Name)
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	if err := h.services.Tax.UpdateRate(c.Request.Context(), existing); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "tax_update_failed")
		return
	}
	c.JSON(http.StatusOK, existing)
}

// AdminDeleteTaxRate handles DELETE /api/v1/admin/tax-rates/:id
func (h *Handler) AdminDeleteTaxRate(c *gin.Context) {
	if h.services == nil || h.services.Tax == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if _, err := h.services.Tax.GetRate(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "tax_not_found")
		return
	}
	if err := h.services.Tax.DeleteRate(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "tax_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tax rate deleted", "id": id})
}
