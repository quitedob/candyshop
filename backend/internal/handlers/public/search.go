package public

import (
	modelsProduct "candypro/api/internal/models/product"
	apiresp "candypro/api/internal/pkg/response"
	productService "candypro/api/internal/services/product"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var _ modelsProduct.SearchResponse

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
		apiresp.ServiceUnavailableResp(c)
		return
	}

	query := c.Query("q")
	searchType := c.DefaultQuery("type", "all")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	if query == "" {
		apiresp.ErrorResp(c, http.StatusBadRequest, "search_query_required")
		return
	}

	// SEC-22: Cap search query length (consistent with SSE handler 4000 char limit)
	if len(query) > 4000 {
		apiresp.ErrorResp(c, http.StatusBadRequest, "search_query_too_long")
		return
	}

	response, err := h.services.Search.Search(c.Request.Context(), query, searchType, limit)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "search_failed")
		return
	}

	locale, _ := c.Get("locale")
	loc, _ := locale.(string)
	if loc == "" {
		loc = "zh"
	}
	if response != nil && len(response.Products) > 0 {
		productService.ApplyProductTranslationsBatch(response.Products, loc)
	}

	c.JSON(http.StatusOK, response)
}
