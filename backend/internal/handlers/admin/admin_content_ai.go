package admin

import (
	"net/http"

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
		Topic    string `json:"topic" binding:"required"`
		Type     string `json:"type"`     // "post" or "case", default "post"
		Language string `json:"language"` // "en" or "zh", default "en"
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.Type == "" {
		req.Type = "post"
	}
	if req.Language == "" {
		req.Language = "en"
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

	// Auto-translate text fields to all non-source locales
	sourceData := map[string]string{
		"title":   content.Title,
		"excerpt": content.Excerpt,
		"content": content.Content,
	}
	targetLocales := make([]string, 0, 8)
	for _, l := range []string{"en", "zh", "ko", "ar", "ja", "th", "vi", "id", "ms"} {
		if l != req.Language {
			targetLocales = append(targetLocales, l)
		}
	}
	transResult, _ := h.aiService.BatchTranslateFields(c.Request.Context(), sourceData, targetLocales)
	// Sanitize translated content fields
	for locale, fields := range transResult.Fields {
		if v, ok := fields["content"]; ok {
			fields["content"] = sanitize.HTML(v)
		}
		if v, ok := fields["excerpt"]; ok {
			fields["excerpt"] = sanitize.HTML(v)
		}
		transResult.Fields[locale] = fields
	}

	resp := gin.H{
		"content":      content,
		"translations": transResult.Fields,
		"topic":        req.Topic,
		"type":         req.Type,
	}
	if len(transResult.Warnings) > 0 {
		resp["warnings"] = transResult.Warnings
	}

	c.JSON(http.StatusOK, resp)
}
