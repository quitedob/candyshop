package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/response"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

type adminProductUpdateRequest struct {
	Slug           *string   `json:"slug"`
	Name           *string   `json:"name"`
	Summary        *string   `json:"summary"`
	Description    *string   `json:"description"`
	Category       *string   `json:"category"`
	CategorySlug   *string   `json:"categorySlug"`
	Thumbnail      *string   `json:"thumbnail"`
	Images         *[]string `json:"images"`
	OEMAvailable   *bool     `json:"oemAvailable"`
	HalalCertified *bool     `json:"halalCertified"`
	Certifications *[]string `json:"certifications"`
	MOQ            *int      `json:"moq"`
	StockQuantity  *int      `json:"stockQuantity"`
	LeadTime       *string   `json:"leadTime"`
	Featured       *bool     `json:"featured"`
	Flavors        *[]string `json:"flavors"`
	Shapes         *[]string `json:"shapes"`
	Ingredients    *string   `json:"ingredients"`
	Allergens      *string   `json:"allergens"`
	ShelfLife      *string   `json:"shelfLife"`
	Storage        *string   `json:"storage"`
	Status         *string   `json:"status"`
}

// AdminCreateProduct creates a new product
// @Summary Admin create product
// @Tags admin-products
// @Produce json
// @Router /admin/products [post]
func (h *Handler) AdminCreateProduct(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	var product modelsProduct.Product
	if !response.BindJSONOrInvalid(c, &product) {
		return
	}

	product.Name = strings.TrimSpace(product.Name)
	if product.Name == "" {
		response.InvalidResp(c, "product_name_required")
		return
	}

	if strings.TrimSpace(product.ID) == "" {
		product.ID = crypto.GenerateID()
	}
	if strings.TrimSpace(product.Slug) == "" {
		product.Slug = buildProductSlug(product.Name)
	}
	if strings.TrimSpace(product.CategorySlug) == "" {
		product.CategorySlug = normalizeSlug(product.Category)
	}
	if strings.TrimSpace(product.Status) == "" {
		product.Status = modelsProduct.ProductStatusActive
	} else if !modelsProduct.IsValidProductStatus(strings.TrimSpace(product.Status)) {
		response.InvalidResp(c, "product_status_invalid")
		return
	}
	if product.StockQuantity < 0 {
		response.InvalidResp(c, "product_stock_negative")
		return
	}

	now := time.Now()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	product.UpdatedAt = now

	if err := h.services.Product.CreateProduct(c.Request.Context(), &product); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "product_slug_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "product_create_failed")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created successfully",
		"product": product,
	})
}

// AdminUpdateProduct updates a product
// @Summary Admin update product
// @Tags admin-products
// @Produce json
// @Param id path string true "Product ID"
// @Router /admin/products/{id} [put]
func (h *Handler) AdminUpdateProduct(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	product, err := h.services.Product.GetProductByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}

	var req adminProductUpdateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.StockQuantity != nil && *req.StockQuantity < 0 {
		response.InvalidResp(c, "product_stock_negative")
		return
	}
	if req.Status != nil && !modelsProduct.IsValidProductStatus(strings.TrimSpace(*req.Status)) {
		response.InvalidResp(c, "product_status_invalid")
		return
	}

	applyProductPatch(product, req)
	product.UpdatedAt = time.Now()
	if product.Name == "" {
		response.InvalidResp(c, "product_name_empty")
		return
	}

	if err := h.services.Product.UpdateProduct(c.Request.Context(), product); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "product_slug_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "product_update_failed")
		return
	}

	out := gin.H{
		"message": "Product updated successfully",
		"product": product,
	}
	if country := strings.TrimSpace(c.Query("complianceSuggestCountry")); country != "" && h.aiService != nil {
		q := strings.TrimSpace(product.Name) + " " + strings.TrimSpace(product.Ingredients) + " " + strings.TrimSpace(product.Allergens)
		out["complianceCopilot"] = gin.H{
			"disclaimer": "RAG output is not legal advice; hard rules and profiles still govern release.",
			"lookup":     h.aiService.LookupCompliance(country, q, 5),
		}
	}
	c.JSON(http.StatusOK, out)
}

// AdminDeleteProduct deletes a product
// @Summary Admin delete product
// @Tags admin-products
// @Produce json
// @Param id path string true "Product ID"
// @Router /admin/products/{id} [delete]
func (h *Handler) AdminDeleteProduct(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	if _, err := h.services.Product.GetProductByID(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}

	if err := h.services.Product.DeleteProduct(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "product_delete_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product deleted successfully",
		"id":      id,
	})
}

func applyProductPatch(product *modelsProduct.Product, req adminProductUpdateRequest) {
	if req.Slug != nil {
		product.Slug = normalizeSlug(*req.Slug)
	}
	if req.Name != nil {
		product.Name = strings.TrimSpace(*req.Name)
		if product.Slug == "" {
			product.Slug = buildProductSlug(product.Name)
		}
	}
	if req.Summary != nil {
		product.Summary = strings.TrimSpace(*req.Summary)
	}
	if req.Description != nil {
		product.Description = strings.TrimSpace(*req.Description)
	}
	if req.Category != nil {
		product.Category = strings.TrimSpace(*req.Category)
	}
	if req.CategorySlug != nil {
		product.CategorySlug = normalizeSlug(*req.CategorySlug)
	}
	if req.Thumbnail != nil {
		product.Thumbnail = strings.TrimSpace(*req.Thumbnail)
	}
	if req.Images != nil {
		product.Images = modelsCommon.StringArray(*req.Images)
	}
	if req.OEMAvailable != nil {
		product.OEMAvailable = *req.OEMAvailable
	}
	if req.HalalCertified != nil {
		product.HalalCertified = *req.HalalCertified
	}
	if req.Certifications != nil {
		product.Certifications = modelsCommon.StringArray(*req.Certifications)
	}
	if req.MOQ != nil {
		product.MOQ = *req.MOQ
	}
	if req.StockQuantity != nil {
		product.StockQuantity = *req.StockQuantity
	}
	if req.LeadTime != nil {
		product.LeadTime = strings.TrimSpace(*req.LeadTime)
	}
	if req.Featured != nil {
		product.Featured = *req.Featured
	}
	if req.Flavors != nil {
		product.Flavors = modelsCommon.StringArray(*req.Flavors)
	}
	if req.Shapes != nil {
		product.Shapes = modelsCommon.StringArray(*req.Shapes)
	}
	if req.Ingredients != nil {
		product.Ingredients = strings.TrimSpace(*req.Ingredients)
	}
	if req.Allergens != nil {
		product.Allergens = strings.TrimSpace(*req.Allergens)
	}
	if req.ShelfLife != nil {
		product.ShelfLife = strings.TrimSpace(*req.ShelfLife)
	}
	if req.Storage != nil {
		product.Storage = strings.TrimSpace(*req.Storage)
	}
	if req.Status != nil {
		product.Status = strings.TrimSpace(*req.Status)
	}

	if product.CategorySlug == "" && product.Category != "" {
		product.CategorySlug = normalizeSlug(product.Category)
	}
	if product.Slug == "" && product.Name != "" {
		product.Slug = buildProductSlug(product.Name)
	}
}

func buildProductSlug(name string) string {
	slug := normalizeSlug(name)
	if slug == "" {
		return "product-" + strings.ToLower(crypto.GenerateSlug())
	}
	return slug
}

func normalizeSlug(input string) string {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return ""
	}

	re := regexp.MustCompile(`\s+`)
	input = re.ReplaceAllString(input, "-")

	var b strings.Builder
	prevDash := false
	for _, r := range input {
		valid := unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-'
		if !valid {
			continue
		}
		if r == '-' {
			if prevDash {
				continue
			}
			prevDash = true
		} else {
			prevDash = false
		}
		b.WriteRune(r)
	}
	return strings.Trim(b.String(), "-")
}
