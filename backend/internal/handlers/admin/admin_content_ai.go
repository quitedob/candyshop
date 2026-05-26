package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"candypro/api/internal/pkg/eino"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/sanitize"

	"github.com/gin-gonic/gin"
)

// AdminAIGenerateContent uses AI to generate blog/case content with structured JSON output.
// @Summary AI generate content
// @Tags admin-content
// @Accept json
// @Produce json
// @Router /admin/content/ai-generate [post]
func (h *Handler) AdminAIGenerateContent(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	var req struct {
		Topic          string `json:"topic" binding:"required"`
		Type           string `json:"type"`           // "post" or "case", default "post"
		Language       string `json:"language"`       // default "zh"
		Translate      bool   `json:"translate"`      // when true, auto-translate to other locales
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.Type == "" {
		req.Type = "post"
	}
	if req.Language == "" {
		req.Language = "zh"
	}

	content, rawFallback, err := h.aiService.GenerateContent(c.Request.Context(), req.Topic, req.Type, req.Language)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_error")
		return
	}

	if content == nil && rawFallback != "" {
		c.JSON(http.StatusOK, gin.H{
			"content":    sanitize.HTML(rawFallback),
			"topic":      req.Topic,
			"type":       req.Type,
			"parseError": true,
		})
		return
	}

	// Sanitize AI-generated HTML before returning to client
	content.Content = sanitize.HTML(content.Content)
	content.Excerpt = sanitize.HTML(content.Excerpt)

	resp := gin.H{
		"content":  content,
		"topic":    req.Topic,
		"type":     req.Type,
		"language": req.Language,
	}

	if !req.Translate {
		c.JSON(http.StatusOK, resp)
		return
	}

	// 经 translate_content 同款路径自动翻译至其余语言
	sourceData := eino.ContentFieldsForTranslation(content)
	targetLocales := eino.LocalesExcluding(req.Language)
	transResult, _ := h.batchTranslateContent(c.Request.Context(), sourceData, targetLocales)
	// Sanitize translated content fields
	htmlFields := []string{"content", "excerpt", "challenge", "solution", "result"}
	for locale, fields := range transResult.Fields {
		for _, fn := range htmlFields {
			if v, ok := fields[fn]; ok {
				fields[fn] = sanitize.HTML(v)
			}
		}
		transResult.Fields[locale] = fields
	}

	resp["translations"] = transResult.Fields
	if len(transResult.Warnings) > 0 {
		resp["warnings"] = transResult.Warnings
	}

	c.JSON(http.StatusOK, resp)
}

type adminAITranslateFieldsRequest struct {
	ContentType   string            `json:"contentType"` // "post" or "case"
	SourceLocale  string            `json:"sourceLocale"`
	TargetLocales []string          `json:"targetLocales" binding:"required"`
	Fields        map[string]string `json:"fields" binding:"required"`
}

// AdminAITranslateContentFields translates draft fields without persisting content.
// Used after the admin reviews the Chinese draft and clicks publish.
func (h *Handler) AdminAITranslateContentFields(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	var req adminAITranslateFieldsRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if len(req.Fields) == 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}

	sourceData := make(map[string]string, len(req.Fields))
	for k, v := range req.Fields {
		if strings.TrimSpace(v) != "" {
			sourceData[k] = v
		}
	}
	if len(sourceData) == 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}

	transResult, err := h.batchTranslateContent(c.Request.Context(), sourceData, req.TargetLocales)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}
	if len(transResult.Fields) == 0 && len(transResult.Warnings) > 0 {
		response.ErrorResp(c, http.StatusInternalServerError, "translation_failed")
		return
	}

	htmlFields := []string{"content", "excerpt", "challenge", "solution", "result"}
	sanitizeHTMLFields(transResult.Fields, htmlFields...)

	resp := gin.H{"translations": transResult.Fields}
	if len(transResult.Warnings) > 0 {
		resp["warnings"] = transResult.Warnings
	}
	c.JSON(http.StatusOK, resp)
}

type adminAIReviseRequest struct {
	Type         string            `json:"type"`
	Language     string            `json:"language"`
	Instruction  string            `json:"instruction" binding:"required"`
	SelectedText string            `json:"selectedText"`
	FocusFields  []string          `json:"focusFields"`
	Current      map[string]string `json:"current" binding:"required"`
}

// AdminAIReviseContentDraft revises specific parts of a draft without full regeneration.
func (h *Handler) AdminAIReviseContentDraft(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	var req adminAIReviseRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if strings.TrimSpace(req.Instruction) == "" || len(req.Current) == 0 {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if req.Type == "" {
		req.Type = "post"
	}
	if req.Language == "" {
		req.Language = "zh"
	}

	result, rawFallback, err := h.aiService.ReviseContent(
		c.Request.Context(),
		req.Type,
		req.Language,
		req.Instruction,
		req.SelectedText,
		req.FocusFields,
		req.Current,
	)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_error")
		return
	}
	if result == nil && rawFallback != "" {
		c.JSON(http.StatusOK, gin.H{"parseError": true, "raw": rawFallback})
		return
	}

	updates := normalizeReviseUpdates(req.Type, result.Updates)
	c.JSON(http.StatusOK, gin.H{
		"updates": updates,
		"summary": result.Summary,
	})
}

func normalizeReviseUpdates(contentType string, raw map[string]json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	htmlFields := map[string]bool{
		"content": true, "excerpt": true, "challenge": true, "solution": true, "result": true,
	}
	out := make(map[string]any, len(raw))
	for key, val := range raw {
		if len(val) == 0 || string(val) == "null" {
			continue
		}
		switch key {
		case "tags", "services":
			var items []string
			if err := json.Unmarshal(val, &items); err == nil {
				out[key] = items
			}
		case "readTime":
			var n int
			if err := json.Unmarshal(val, &n); err == nil {
				out[key] = n
			}
		default:
			var s string
			if err := json.Unmarshal(val, &s); err == nil && strings.TrimSpace(s) != "" {
				if htmlFields[key] {
					s = sanitize.HTML(s)
				}
				if key == "category" && contentType == "post" {
					s = eino.NormalizeBlogCategory(s)
				}
				out[key] = s
			}
		}
	}
	return out
}

type adminAIInlineEditRequest struct {
	Instruction   string `json:"instruction" binding:"required"`
	SelectedText  string `json:"selectedText" binding:"required"`
	SelectedHTML  string `json:"selectedHtml"`
	ContextBefore string `json:"contextBefore"`
	ContextAfter  string `json:"contextAfter"`
	FieldType     string `json:"fieldType"` // html, plain, title
	Language      string `json:"language"`
}

// AdminAIInlineEditContent performs Cursor-style inline selection replacement.
func (h *Handler) AdminAIInlineEditContent(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	var req adminAIInlineEditRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if strings.TrimSpace(req.Instruction) == "" || strings.TrimSpace(req.SelectedText) == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if req.FieldType == "" {
		req.FieldType = "plain"
	}
	if req.Language == "" {
		req.Language = "zh"
	}

	result, rawFallback, err := h.aiService.InlineEditContent(
		c.Request.Context(),
		req.Language,
		req.Instruction,
		req.SelectedText,
		req.SelectedHTML,
		req.ContextBefore,
		req.ContextAfter,
		req.FieldType,
	)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_error")
		return
	}
	if result == nil && rawFallback != "" {
		c.JSON(http.StatusOK, gin.H{"parseError": true, "raw": rawFallback})
		return
	}

	replacement := result.Replacement
	if result.Format == "html" || req.FieldType == "html" {
		replacement = sanitize.HTML(replacement)
	}

	c.JSON(http.StatusOK, gin.H{
		"replacement": replacement,
		"format":      result.Format,
	})
}
