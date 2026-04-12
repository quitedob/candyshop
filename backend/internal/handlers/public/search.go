package public

import (
	"candypro/api/internal/utils"

	modelsProduct "candypro/api/internal/models/product"
)

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ===== Search =====

// Search searches across all content
// @Summary Search content
// @Tags search
// @Produce json
// @Param q query string true "Search query"
// @Param type query string false "Content type (products, posts, cases, all)" default(all)
// @Param limit query int false "Number of results per type" default(10)
// @Success 200 {object} modelsProduct.SearchResponse
// @Router /search [get]
func (h *Handler) Search(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	query := c.Query("q")
	searchType := c.DefaultQuery("type", "all")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	if query == "" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "bad_request",
			Message: "Search query is required",
		})
		return
	}

	// SEC-22: Cap search query length (consistent with SSE handler 4000 char limit)
	if len(query) > 4000 {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "bad_request",
			Message: "Search query is too long (max 4000 characters)",
		})
		return
	}

	response, err := h.services.Search.Search(c.Request.Context(), query, searchType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Search failed",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
