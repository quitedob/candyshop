package public

import (
	"candypro/api/internal/pkg/pagination"
	apiresp "candypro/api/internal/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ===== Content =====

// GetPosts returns paginated blog posts
// @Summary Get blog posts
// @Tags blog
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param category query string false "Filter by category"
// @Success 200 {object} modelsProduct.PaginatedResponse
// @Router /posts [get]
func (h *Handler) GetPosts(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 10, 50)
	category := c.Query("category")

	response, err := h.services.Content.GetPosts(c.Request.Context(), page, limit, category)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "posts_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPost returns a blog post by slug
// @Summary Get blog post by slug
// @Tags blog
// @Produce json
// @Param slug path string true "Post slug"
// @Success 200 {object} modelsProduct.BlogPost
// @Failure 404 {object} modelsCommon.ErrorResponse
// @Router /posts/{slug} [get]
func (h *Handler) GetPost(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	slug := c.Param("slug")

	post, err := h.services.Content.GetPostBySlug(c.Request.Context(), slug)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusNotFound, "post_not_found")
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetRelatedPosts returns related blog posts
// @Summary Get related blog posts
// @Tags blog
// @Produce json
// @Param slug path string true "Post slug"
// @Param limit query int false "Number of posts" default(3)
// @Success 200 {array} modelsProduct.BlogPost
// @Router /posts/{slug}/related [get]
func (h *Handler) GetRelatedPosts(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	slug := c.Param("slug")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "3"))
	if err != nil || limit < 1 || limit > 20 {
		limit = 3
	}

	posts, err := h.services.Content.GetRelatedPosts(c.Request.Context(), slug, limit)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "related_posts_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, posts)
}

// GetCases returns paginated case studies
// @Summary Get case studies
// @Tags cases
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param industry query string false "Filter by industry"
// @Success 200 {object} modelsProduct.PaginatedResponse
// @Router /cases [get]
func (h *Handler) GetCases(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 10, 50)
	industry := c.Query("industry")

	response, err := h.services.Content.GetCases(c.Request.Context(), page, limit, industry)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "cases_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetRelatedCases returns related case studies
// @Summary Get related case studies
// @Tags cases
// @Produce json
// @Param slug path string true "Case study slug"
// @Param limit query int false "Number of cases" default(3)
// @Success 200 {array} modelsProduct.CaseStudy
// @Router /cases/{slug}/related [get]
func (h *Handler) GetRelatedCases(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	slug := c.Param("slug")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "3"))
	if err != nil || limit < 1 || limit > 20 {
		limit = 3
	}

	cases, err := h.services.Content.GetRelatedCases(c.Request.Context(), slug, limit)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "related_cases_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, cases)
}

// GetCase returns a case study by slug
// @Summary Get case study by slug
// @Tags cases
// @Produce json
// @Param slug path string true "Case study slug"
// @Success 200 {object} modelsProduct.CaseStudy
// @Failure 404 {object} modelsCommon.ErrorResponse
// @Router /cases/{slug} [get]
func (h *Handler) GetCase(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	slug := c.Param("slug")

	caseStudy, err := h.services.Content.GetCaseBySlug(c.Request.Context(), slug)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusNotFound, "case_not_found")
		return
	}

	c.JSON(http.StatusOK, caseStudy)
}
