package admin

import (
	"net/http"

	"candypro/api/internal/pkg/eino"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AdminAIGenerateProduct uses AI to generate a complete product from a natural-language description
// and auto-translates text fields to all supported locales.
// @Summary AI generate product
// @Tags admin-products
// @Accept json
// @Produce json
// @Router /admin/products/ai-generate [post]
func (h *Handler) AdminAIGenerateProduct(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	var req struct {
		Description string `json:"description" binding:"required"`
		Language    string `json:"language"` // "en" or "zh", default "zh"
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Language == "" {
		req.Language = "zh"
	}

	product, rawFallback, err := h.aiService.GenerateProduct(c.Request.Context(), req.Description, req.Language)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_error")
		return
	}

	if product == nil && rawFallback != "" {
		c.JSON(http.StatusOK, gin.H{
			"product":    rawFallback,
			"parseError": true,
		})
		return
	}

	// Auto-translate text fields to all non-source locales
	sourceData := eino.ProductFieldsForTranslation(product)
	targetLocales := eino.LocalesExcluding(req.Language)
	transResult, _ := h.batchTranslateContent(c.Request.Context(), sourceData, targetLocales)

	c.JSON(http.StatusOK, gin.H{
		"product":      product,
		"translations": transResult.Fields,
	})
}
