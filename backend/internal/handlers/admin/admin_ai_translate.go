package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/sanitize"
	tradeSvc "candypro/api/internal/services/trade"

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
	transResult, err := h.batchTranslateContent(c.Request.Context(), sourceData, req.TargetLocales)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}
	if len(transResult.Fields) == 0 && len(transResult.Warnings) > 0 {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	if product.Translations == nil {
		product.Translations = make(modelsCommon.JSONMap)
	}
	// H24: sanitize the rich-text product fields on write so AI-injected HTML is
	// never merged into Translations unsanitized. Plain-text scalars
	// (name/summary/ingredients/…) are left untouched: they render escaped, and
	// bluemonday would otherwise turn a legitimate '&' into '&amp;'.
	sanitizeHTMLFields(transResult.Fields, "description")
	ensureSourceLocaleInTranslations(product, sourceData)
	mergeTranslations(product.Translations, transResult)

	// AdminAITranslateProduct loads the full row first, so the full-row
	// overwrite is safe and persists the merged Translations map (H11); the
	// zero-skip Update would drop a Translation patch.
	if err := h.services.Product.UpdateProductAll(c.Request.Context(), product); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "update_failed")
		return
	}
	resp := gin.H{"success": true, "translations": product.Translations}
	if len(transResult.Warnings) > 0 {
		resp["warnings"] = transResult.Warnings
	}
	c.JSON(http.StatusOK, resp)
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

	sourceData := postFieldsToTranslate(post)

	transResult, err := h.batchTranslateContent(c.Request.Context(), sourceData, req.TargetLocales)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}
	if len(transResult.Fields) == 0 && len(transResult.Warnings) > 0 {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	if post.Translations == nil {
		post.Translations = make(modelsCommon.JSONMap)
	}
	sanitizeHTMLFields(transResult.Fields, "content", "excerpt")
	ensureSourceLocaleInTranslationsMap(post.Translations, sourceData)
	mergeTranslations(post.Translations, transResult)

	if err := h.services.Content.UpdatePost(c.Request.Context(), post); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "update_failed")
		return
	}
	resp := gin.H{"success": true, "translations": post.Translations}
	if len(transResult.Warnings) > 0 {
		resp["warnings"] = transResult.Warnings
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) translateCase(c *gin.Context, req adminAITranslateRequest) {
	caseStudy, err := h.services.Content.GetCaseByID(c.Request.Context(), req.ContentID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	sourceData := caseFieldsToTranslate(caseStudy)

	transResult, err := h.batchTranslateContent(c.Request.Context(), sourceData, req.TargetLocales)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}
	if len(transResult.Fields) == 0 && len(transResult.Warnings) > 0 {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	if caseStudy.Translations == nil {
		caseStudy.Translations = make(modelsCommon.JSONMap)
	}
	sanitizeHTMLFields(transResult.Fields, "challenge", "solution", "result")
	ensureSourceLocaleInTranslationsMap(caseStudy.Translations, sourceData)
	mergeTranslations(caseStudy.Translations, transResult)

	if err := h.services.Content.UpdateCase(c.Request.Context(), caseStudy); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "update_failed")
		return
	}
	resp := gin.H{"success": true, "translations": caseStudy.Translations}
	if len(transResult.Warnings) > 0 {
		resp["warnings"] = transResult.Warnings
	}
	c.JSON(http.StatusOK, resp)
}

func productFieldsToTranslate(p *modelsProduct.Product) map[string]string {
	sourceLocale := i18n.DefaultLocale()
	out := map[string]string{}

	scalars := map[string]string{
		"name":        p.Name,
		"summary":     p.Summary,
		"description": p.Description,
		"category":    p.Category,
		"ingredients": p.Ingredients,
		"allergens":   p.Allergens,
		"shelfLife":   p.ShelfLife,
		"storage":     p.Storage,
		"leadTime":    p.LeadTime,
	}
	if p.Translations != nil {
		if tr, ok := p.Translations[sourceLocale]; ok {
			for k, v := range tr {
				if strings.TrimSpace(v) != "" {
					out[k] = v
				}
			}
		}
	}
	for k, v := range scalars {
		if _, exists := out[k]; !exists && strings.TrimSpace(v) != "" {
			out[k] = v
		}
	}
	if _, ok := out["flavors"]; !ok && len(p.Flavors) > 0 {
		if b, err := json.Marshal(p.Flavors); err == nil {
			out["flavors"] = string(b)
		}
	}
	if _, ok := out["shapes"]; !ok && len(p.Shapes) > 0 {
		if b, err := json.Marshal(p.Shapes); err == nil {
			out["shapes"] = string(b)
		}
	}
	return out
}

func postFieldsToTranslate(post *modelsProduct.BlogPost) map[string]string {
	sourceLocale := i18n.DefaultLocale()
	out := map[string]string{}
	scalars := map[string]string{
		"title": post.Title, "excerpt": post.Excerpt, "content": post.Content,
		"authorName": post.AuthorName, "authorTitle": post.AuthorTitle, "authorBio": post.AuthorBio,
	}
	if post.Author.Name != "" && scalars["authorName"] == "" {
		scalars["authorName"] = post.Author.Name
	}
	if post.Author.Title != "" && scalars["authorTitle"] == "" {
		scalars["authorTitle"] = post.Author.Title
	}
	if post.Author.Bio != "" && scalars["authorBio"] == "" {
		scalars["authorBio"] = post.Author.Bio
	}
	mergeTranslationSource(out, scalars, post.Translations, sourceLocale)
	return out
}

func caseFieldsToTranslate(cs *modelsProduct.CaseStudy) map[string]string {
	sourceLocale := i18n.DefaultLocale()
	out := map[string]string{}
	scalars := map[string]string{
		"title": cs.Title, "challenge": cs.Challenge, "solution": cs.Solution, "result": cs.Result,
	}
	mergeTranslationSource(out, scalars, cs.Translations, sourceLocale)
	return out
}

func mergeTranslationSource(out, scalars map[string]string, translations modelsCommon.JSONMap, locale string) {
	if translations != nil {
		if tr, ok := translations[locale]; ok {
			for k, v := range tr {
				if strings.TrimSpace(v) != "" {
					out[k] = v
				}
			}
		}
	}
	for k, v := range scalars {
		if _, exists := out[k]; !exists && strings.TrimSpace(v) != "" {
			out[k] = v
		}
	}
}

func mergeTranslations(target modelsCommon.JSONMap, result *tradeSvc.TranslationResult) {
	for locale, fields := range result.Fields {
		if target[locale] == nil {
			target[locale] = make(map[string]string)
		}
		for fieldName, translatedText := range fields {
			target[locale][fieldName] = translatedText
		}
	}
}

// ensureSourceLocaleInTranslations 将源语言（zh）字段写入 translations，避免 AI 翻译后 zh 丢失。
func ensureSourceLocaleInTranslations(p *modelsProduct.Product, sourceData map[string]string) {
	ensureSourceLocaleInTranslationsMap(p.Translations, sourceData)
}

func ensureSourceLocaleInTranslationsMap(target modelsCommon.JSONMap, sourceData map[string]string) {
	if target == nil || len(sourceData) == 0 {
		return
	}
	locale := i18n.DefaultLocale()
	if target[locale] == nil {
		target[locale] = make(map[string]string)
	}
	for k, v := range sourceData {
		if strings.TrimSpace(v) == "" {
			continue
		}
		target[locale][k] = v
	}
}

// sanitizeHTMLFields sanitizes specified HTML fields in all locale translation maps.
func sanitizeHTMLFields(fields map[string]map[string]string, htmlFieldNames ...string) {
	for _, localeFields := range fields {
		for _, fieldName := range htmlFieldNames {
			if v, ok := localeFields[fieldName]; ok {
				localeFields[fieldName] = sanitize.HTML(v)
			}
		}
	}
}
