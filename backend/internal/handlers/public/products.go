package public

import (
	"candypro/api/internal/pkg/pagination"
	apiresp "candypro/api/internal/pkg/response"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ===== Products =====

// GetProducts returns paginated products
// @Summary Get products
// @Tags products
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(12)
// @Param category query string false "Filter by category slug"
// @Success 200 {object} modelsProduct.PaginatedResponse
// @Router /products [get]
func (h *Handler) GetProducts(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 12, 100)
	category := c.Query("category")

	// M7: Parse all filter params the frontend sends
	halal := c.Query("halal") == "true"
	oemOnly := c.Query("oemOnly") == "true"
	featuredOnly := c.Query("featured") == "true"
	search := c.Query("search")
	sort := c.Query("sort")

	var minMOQ, maxMOQ int
	if v := c.Query("minMoq"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			minMOQ = n
		}
	}
	if v := c.Query("maxMoq"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxMOQ = n
		}
	}

	// Use filtered query when any filter is active
	if halal || oemOnly || featuredOnly || search != "" || sort != "" || minMOQ > 0 || maxMOQ > 0 {
		response, err := h.services.Product.GetProductsFiltered(c.Request.Context(), page, limit, halal, oemOnly, featuredOnly, search, sort, minMOQ, maxMOQ)
		if err != nil {
			apiresp.ErrorResp(c, http.StatusInternalServerError, "product_fetch_failed")
			return
		}
		c.JSON(http.StatusOK, response)
		return
	}

	response, err := h.services.Product.GetProducts(c.Request.Context(), page, limit, category)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "product_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct returns a single product by slug
// @Summary Get product by slug
// @Tags products
// @Produce json
// @Param slug path string true "Product slug"
// @Success 200 {object} modelsProduct.Product
// @Failure 404 {object} modelsCommon.ErrorResponse
// @Router /products/{slug} [get]
func (h *Handler) GetProduct(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	slug := c.Param("slug")

	product, err := h.services.Product.GetProduct(c.Request.Context(), slug)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}

	// Increment view count (best-effort, non-blocking)
	if h.services != nil && h.services.Product != nil {
		_ = h.services.Product.IncrementViewCount(c.Request.Context(), product.ID)
	}

	c.JSON(http.StatusOK, product)
}

// GetFeaturedProducts returns featured products
// @Summary Get featured products
// @Tags products
// @Produce json
// @Param limit query int false "Number of products" default(8)
// @Success 200 {array} modelsProduct.Product
// @Router /products/featured [get]
func (h *Handler) GetFeaturedProducts(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "8"))
	if err != nil || limit < 1 || limit > 50 {
		limit = 8
	}

	products, err := h.services.Product.GetFeaturedProducts(c.Request.Context(), limit)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "featured_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetRelatedProducts returns related products
// @Summary Get related products
// @Tags products
// @Produce json
// @Param slug path string true "Product slug"
// @Param limit query int false "Number of products" default(4)
// @Success 200 {array} modelsProduct.Product
// @Router /products/{slug}/related [get]
func (h *Handler) GetRelatedProducts(c *gin.Context) {
	if !(h.services != nil) {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	slug := c.Param("slug")
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "4"))
	if err != nil || limit < 1 || limit > 20 {
		limit = 4
	}

	products, err := h.services.Product.GetRelatedProducts(c.Request.Context(), slug, limit)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "related_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, products)
}

// GetProductVariants returns active SKU variants for a product.
func (h *Handler) GetProductVariants(c *gin.Context) {
	if h.services == nil {
		apiresp.ServiceUnavailableResp(c)
		return
	}

	productID := c.Param("slug") // reuse slug param — accepts slug or ID
	product, err := h.services.Product.GetProduct(c.Request.Context(), productID)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}

	variants, err := h.services.Product.GetProductVariants(c.Request.Context(), product.ID)
	if err != nil {
		apiresp.ErrorResp(c, http.StatusInternalServerError, "variant_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, variants)
}
