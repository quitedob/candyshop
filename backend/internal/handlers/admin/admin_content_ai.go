package admin

import (
	"net/http"

	"candypro/api/internal/pkg/response"

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
			"content":    rawFallback,
			"topic":      req.Topic,
			"type":       req.Type,
			"parseError": true,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"topic":   req.Topic,
		"type":    req.Type,
	})
}
