package public

import (
	"candypro/api/internal/utils"

	modelsProduct "candypro/api/internal/models/product"
)

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	categories, err := h.services.Category.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch categories",
		})
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
// @Failure 404 {object} modelsProduct.ErrorResponse
// @Router /categories/{slug} [get]
func (h *Handler) GetCategory(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	slug := c.Param("slug")

	category, err := h.services.Category.GetCategory(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Category not found",
		})
		return
	}

	c.JSON(http.StatusOK, category)
}
