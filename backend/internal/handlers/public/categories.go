package public

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// swag 类型引用（避免 unused import）
var (
	_ modelsProduct.Category
	_ modelsCommon.ErrorResponse
)

// ===== Categories =====

// GetCategories returns all categories
// @Summary Get all categories
// @Tags categories
// @Produce json
// @Success 200 {array} modelsProduct.Category
// @Router /categories [get]
func (h *Handler) GetCategories(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	categories, err := h.services.Category.GetCategories(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "category_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, categories)
}

// GetCategory returns a category by slug
// @Summary Get category by slug
// @Tags categories
// @Produce json
// @Param slug path string true "Category slug"
// @Success 200 {object} modelsProduct.Category
// @Failure 404 {object} modelsCommon.ErrorResponse
// @Router /categories/{slug} [get]
func (h *Handler) GetCategory(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	slug := c.Param("slug")

	category, err := h.services.Category.GetCategory(c.Request.Context(), slug)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "category_not_found")
		return
	}

	c.JSON(http.StatusOK, category)
}
