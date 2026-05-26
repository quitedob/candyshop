package admin

import (
	"encoding/json"
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/sanitize"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
)

type adminProductUpdateRequest struct {
	Slug           *string              `json:"slug"`
	Name           *string              `json:"name"`
	Summary        *string              `json:"summary"`
	Description    *string              `json:"description"`
	Category       *string              `json:"category"`
	CategorySlug   *string              `json:"categorySlug"`
	Thumbnail      *string              `json:"thumbnail"`
	OgImage        *string              `json:"ogImage"`
	Images         *[]string            `json:"images"`
	OEMAvailable   *bool                `json:"oemAvailable"`
	HalalCertified *bool                `json:"halalCertified"`
	Certifications *[]string            `json:"certifications"`
	MOQ            *int                 `json:"moq"`
	StockQuantity  *int                 `json:"stockQuantity"`
	LeadTime       *string              `json:"leadTime"`
	Featured       *bool                `json:"featured"`
	Flavors        *[]string            `json:"flavors"`
	Shapes         *[]string            `json:"shapes"`
	Ingredients    *string              `json:"ingredients"`
	Allergens      *string              `json:"allergens"`
	ShelfLife      *string              `json:"shelfLife"`
	Storage        *string              `json:"storage"`
	Status         *string              `json:"status"`
	BasePrice      *float64             `json:"basePrice"`
	Translations   *modelsCommon.JSONMap `json:"translations"`

	// Weight & Measurement
	NetWeightPerPiece    *float64 `json:"netWeightPerPiece"`
	NetWeightPerPack     *float64 `json:"netWeightPerPack"`
	GrossWeightPerCarton *float64 `json:"grossWeightPerCarton"`
	PiecesPerPack        *int     `json:"piecesPerPack"`
	PacksPerCarton       *int     `json:"packsPerCarton"`

	// Dimensions
	ProductLengthMM *float64 `json:"productLengthMM"`
	ProductWidthMM  *float64 `json:"productWidthMM"`
	ProductHeightMM *float64 `json:"productHeightMM"`

	// Nutrition
	EnergyKj       *float64 `json:"energyKj"`
	EnergyKcal     *float64 `json:"energyKcal"`
	TotalFatG      *float64 `json:"totalFatG"`
	SaturatedFatG  *float64 `json:"saturatedFatG"`
	CarbohydratesG *float64 `json:"carbohydratesG"`
	SugarsG        *float64 `json:"sugarsG"`
	ProteinG       *float64 `json:"proteinG"`
	SaltG          *float64 `json:"saltG"`
	FiberG         *float64 `json:"fiberG"`

	// Ingredient Compliance
	Additives      *[]string `json:"additives"`
	SweetenerType  *string   `json:"sweetenerType"`
	CocoaSolidsPct *float64  `json:"cocoaSolidsPct"`
	MilkSolidsPct  *float64  `json:"milkSolidsPct"`
	GMOStatus       *string   `json:"gmoStatus"`
	MayContain     *[]string `json:"mayContain"`
	WaterActivity  *float64  `json:"waterActivity"`

	// Trade & Barcode
	GTIN  *string `json:"gtin"`
	HSCode *string `json:"hsCode"`

	// Packaging
	PrimaryPackaging *string `json:"primaryPackaging"`
	InnerPackConfig  *string `json:"innerPackConfig"`
	PalletConfig     *string `json:"palletConfig"`

	// Dietary
	IsVegan      *bool `json:"isVegan"`
	IsGlutenFree *bool `json:"isGlutenFree"`
	IsSugarFree  *bool `json:"isSugarFree"`
	IsKosher     *bool `json:"isKosher"`
	IsOrganic    *bool `json:"isOrganic"`

	// Certification Details
	CertificationDetails *modelsProduct.CertificationDetailArray `json:"certificationDetails"`

	// Sample Specs
	SampleMOQ       *int     `json:"sampleMOQ"`
	SampleLeadTime  *string  `json:"sampleLeadTime"`
	SamplePrice     *float64 `json:"samplePrice"`
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
	if product.MOQ < 1 {
		response.InvalidResp(c, "product_moq_min_1")
		return
	}
	if product.BasePrice <= 0 {
		response.InvalidResp(c, "product_price_required")
		return
	}

	syncProductScalarsFromLocale(&product, i18n.DefaultLocale())

	now := time.Now()
	if product.CreatedAt.IsZero() {
		product.CreatedAt = now
	}
	product.UpdatedAt = now

	if rawID, ok := c.Get("userID"); ok {
		if uid, ok2 := rawID.(string); ok2 {
			product.CreatedBy = &uid
			product.UpdatedBy = &uid
		}
	}

	if err := h.services.Product.CreateProduct(c.Request.Context(), &product); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "product_slug_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "product_create_failed")
		return
	}

	h.logActivityAudit(c, "create", "product", product.ID, "", product.Name)
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
	oldName := product.Name

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
	if req.Status != nil {
		newStatus := strings.TrimSpace(*req.Status)
		if err := modelsProduct.ValidateProductStatusTransition(product.Status, newStatus); err != nil {
			response.InvalidResp(c, "product_status_transition_invalid")
			return
		}
	}

	applyProductPatch(product, req)
	syncProductScalarsFromLocale(product, i18n.DefaultLocale())
	product.UpdatedAt = time.Now()

	if rawID, ok := c.Get("userID"); ok {
		if uid, ok2 := rawID.(string); ok2 {
			product.UpdatedBy = &uid
		}
	}

	if req.MOQ != nil && *req.MOQ < 1 {
		response.InvalidResp(c, "product_moq_min_1")
		return
	}

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

	h.logActivityAudit(c, "update", "product", id, oldName, product.Name)

	out := gin.H{
		"message": "Product updated successfully",
		"product": product,
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
	existing, err := h.services.Product.GetProductByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "product_not_found")
		return
	}

	if err := h.services.Product.DeleteProduct(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "product_delete_failed")
		return
	}

	h.logActivityAudit(c, "delete", "product", id, existing.Name, "")
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
		product.Description = sanitize.HTML(strings.TrimSpace(*req.Description))
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
	if req.OgImage != nil {
		product.OgImage = strings.TrimSpace(*req.OgImage)
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
	if req.BasePrice != nil {
		product.BasePrice = *req.BasePrice
	}
	if req.Translations != nil {
		product.Translations = *req.Translations
	}

	// Weight & Measurement
	if req.NetWeightPerPiece != nil {
		product.NetWeightPerPiece = *req.NetWeightPerPiece
	}
	if req.NetWeightPerPack != nil {
		product.NetWeightPerPack = *req.NetWeightPerPack
	}
	if req.GrossWeightPerCarton != nil {
		product.GrossWeightPerCarton = *req.GrossWeightPerCarton
	}
	if req.PiecesPerPack != nil {
		product.PiecesPerPack = *req.PiecesPerPack
	}
	if req.PacksPerCarton != nil {
		product.PacksPerCarton = *req.PacksPerCarton
	}

	// Dimensions
	if req.ProductLengthMM != nil {
		product.ProductLengthMM = *req.ProductLengthMM
	}
	if req.ProductWidthMM != nil {
		product.ProductWidthMM = *req.ProductWidthMM
	}
	if req.ProductHeightMM != nil {
		product.ProductHeightMM = *req.ProductHeightMM
	}

	// Nutrition
	if req.EnergyKj != nil {
		product.EnergyKj = *req.EnergyKj
	}
	if req.EnergyKcal != nil {
		product.EnergyKcal = *req.EnergyKcal
	}
	if req.TotalFatG != nil {
		product.TotalFatG = *req.TotalFatG
	}
	if req.SaturatedFatG != nil {
		product.SaturatedFatG = *req.SaturatedFatG
	}
	if req.CarbohydratesG != nil {
		product.CarbohydratesG = *req.CarbohydratesG
	}
	if req.SugarsG != nil {
		product.SugarsG = *req.SugarsG
	}
	if req.ProteinG != nil {
		product.ProteinG = *req.ProteinG
	}
	if req.SaltG != nil {
		product.SaltG = *req.SaltG
	}
	if req.FiberG != nil {
		product.FiberG = *req.FiberG
	}

	// Ingredient Compliance
	if req.Additives != nil {
		product.Additives = modelsCommon.StringArray(*req.Additives)
	}
	if req.SweetenerType != nil {
		product.SweetenerType = strings.TrimSpace(*req.SweetenerType)
	}
	if req.CocoaSolidsPct != nil {
		product.CocoaSolidsPct = *req.CocoaSolidsPct
	}
	if req.MilkSolidsPct != nil {
		product.MilkSolidsPct = *req.MilkSolidsPct
	}
	if req.GMOStatus != nil {
		product.GMOStatus = strings.TrimSpace(*req.GMOStatus)
	}
	if req.MayContain != nil {
		product.MayContain = modelsCommon.StringArray(*req.MayContain)
	}
	if req.WaterActivity != nil {
		product.WaterActivity = *req.WaterActivity
	}

	// Trade & Barcode
	if req.GTIN != nil {
		product.GTIN = strings.TrimSpace(*req.GTIN)
	}
	if req.HSCode != nil {
		product.HSCode = strings.TrimSpace(*req.HSCode)
	}

	// Packaging
	if req.PrimaryPackaging != nil {
		product.PrimaryPackaging = strings.TrimSpace(*req.PrimaryPackaging)
	}
	if req.InnerPackConfig != nil {
		product.InnerPackConfig = strings.TrimSpace(*req.InnerPackConfig)
	}
	if req.PalletConfig != nil {
		product.PalletConfig = strings.TrimSpace(*req.PalletConfig)
	}

	// Dietary
	if req.IsVegan != nil {
		product.IsVegan = *req.IsVegan
	}
	if req.IsGlutenFree != nil {
		product.IsGlutenFree = *req.IsGlutenFree
	}
	if req.IsSugarFree != nil {
		product.IsSugarFree = *req.IsSugarFree
	}
	if req.IsKosher != nil {
		product.IsKosher = *req.IsKosher
	}
	if req.IsOrganic != nil {
		product.IsOrganic = *req.IsOrganic
	}

	// Certification Details
	if req.CertificationDetails != nil {
		product.CertificationDetails = *req.CertificationDetails
	}

	// Sample Specs
	if req.SampleMOQ != nil {
		product.SampleMOQ = *req.SampleMOQ
	}
	if req.SampleLeadTime != nil {
		product.SampleLeadTime = strings.TrimSpace(*req.SampleLeadTime)
	}
	if req.SamplePrice != nil {
		product.SamplePrice = *req.SamplePrice
	}

	if product.CategorySlug == "" && product.Category != "" {
		product.CategorySlug = normalizeSlug(product.Category)
	}
	if product.Slug == "" && product.Name != "" {
		product.Slug = buildProductSlug(product.Name)
	}
}

// syncProductScalarsFromLocale 将指定语言的翻译字段同步到产品标量列（供搜索/兼容旧逻辑）。
func syncProductScalarsFromLocale(product *modelsProduct.Product, locale string) {
	if product.Translations == nil {
		return
	}
	fields, ok := product.Translations[locale]
	if !ok {
		return
	}
	if v := strings.TrimSpace(fields["name"]); v != "" {
		product.Name = v
	}
	if v := strings.TrimSpace(fields["summary"]); v != "" {
		product.Summary = v
	}
	if v := strings.TrimSpace(fields["description"]); v != "" {
		product.Description = sanitize.HTML(v)
	}
	if v := strings.TrimSpace(fields["category"]); v != "" {
		product.Category = v
	}
	if v := strings.TrimSpace(fields["ingredients"]); v != "" {
		product.Ingredients = v
	}
	if v := strings.TrimSpace(fields["allergens"]); v != "" {
		product.Allergens = v
	}
	if v := strings.TrimSpace(fields["storage"]); v != "" {
		product.Storage = v
	}
	if v := strings.TrimSpace(fields["shelfLife"]); v != "" {
		product.ShelfLife = v
	}
	if v := strings.TrimSpace(fields["leadTime"]); v != "" {
		product.LeadTime = v
	}
	if v := strings.TrimSpace(fields["flavors"]); v != "" {
		var arr []string
		if json.Unmarshal([]byte(v), &arr) == nil && len(arr) > 0 {
			product.Flavors = modelsCommon.StringArray(arr)
		}
	}
	if v := strings.TrimSpace(fields["shapes"]); v != "" {
		var arr []string
		if json.Unmarshal([]byte(v), &arr) == nil && len(arr) > 0 {
			product.Shapes = modelsCommon.StringArray(arr)
		}
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
