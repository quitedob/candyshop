package admin

import (
	"net/http"

	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type adminAITranslateRequest struct {
	ProductID     string   `json:"productId"`
	ContentID     string   `json:"contentId"`
	ContentType   string   `json:"contentType"` // "post" or "case"
	TargetLocales []string `json:"targetLocales" binding:"required"`
}

// AdminAITranslateProduct uses AI to translate product text fields into target locales.
func (h *Handler) AdminAITranslateProduct(c *gin.Context) {
	if h.aiService == nil || h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req adminAITranslateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.ProductID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	product, err := h.services.Product.GetProductByID(c.Request.Context(), req.ProductID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	sourceData := productFieldsToTranslate(product)
	translations, err := h.aiService.BatchTranslateFields(c.Request.Context(), sourceData, req.TargetLocales)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	if product.Translations == nil {
		product.Translations = make(modelsCommon.JSONMap)
	}
	mergeTranslations(product.Translations, translations)

	if err := h.services.Product.UpdateProduct(c.Request.Context(), product); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "translations": product.Translations})
}

// AdminAITranslateContent uses AI to translate blog post or case study fields.
func (h *Handler) AdminAITranslateContent(c *gin.Context) {
	if h.aiService == nil || h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req adminAITranslateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.ContentID == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	switch req.ContentType {
	case "post", "blog_post", "blogpost":
		h.translatePost(c, req)
	case "case", "case_study", "casestudy":
		h.translateCase(c, req)
	default:
		response.InvalidResp(c, "invalid_request")
	}
}

func (h *Handler) translatePost(c *gin.Context, req adminAITranslateRequest) {
	post, err := h.services.Content.GetPostByID(c.Request.Context(), req.ContentID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	sourceData := map[string]string{
		"title":   post.Title,
		"excerpt": post.Excerpt,
		"content": post.Content,
	}

	translations, err := h.aiService.BatchTranslateFields(c.Request.Context(), sourceData, req.TargetLocales)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	if post.Translations == nil {
		post.Translations = make(modelsCommon.JSONMap)
	}
	mergeTranslations(post.Translations, translations)

	if err := h.services.Content.UpdatePost(c.Request.Context(), post); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "translations": post.Translations})
}

func (h *Handler) translateCase(c *gin.Context, req adminAITranslateRequest) {
	caseStudy, err := h.services.Content.GetCaseByID(c.Request.Context(), req.ContentID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	sourceData := map[string]string{
		"title":     caseStudy.Title,
		"challenge": caseStudy.Challenge,
		"solution":  caseStudy.Solution,
		"result":    caseStudy.Result,
	}

	translations, err := h.aiService.BatchTranslateFields(c.Request.Context(), sourceData, req.TargetLocales)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	if caseStudy.Translations == nil {
		caseStudy.Translations = make(modelsCommon.JSONMap)
	}
	mergeTranslations(caseStudy.Translations, translations)

	if err := h.services.Content.UpdateCase(c.Request.Context(), caseStudy); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "translations": caseStudy.Translations})
}

func productFieldsToTranslate(p *modelsProduct.Product) map[string]string {
	return map[string]string{
		"name":        p.Name,
		"summary":     p.Summary,
		"description": p.Description,
		"ingredients": p.Ingredients,
		"allergens":   p.Allergens,
		"shelfLife":   p.ShelfLife,
		"storage":     p.Storage,
		"leadTime":    p.LeadTime,
	}
}

func mergeTranslations(target modelsCommon.JSONMap, translations map[string]map[string]string) {
	for locale, fields := range translations {
		if target[locale] == nil {
			target[locale] = make(map[string]string)
		}
		for fieldName, translatedText := range fields {
			target[locale][fieldName] = translatedText
		}
	}
}
