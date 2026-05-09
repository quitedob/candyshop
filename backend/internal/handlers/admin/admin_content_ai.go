package admin

import (
	"fmt"
	"net/http"

	"candypro/api/internal/utils"

	"github.com/gin-gonic/gin"
)

// AdminAIGenerateContent uses AI to generate blog/case content from a topic.
// @Summary AI generate content
// @Tags admin-content
// @Accept json
// @Produce json
// @Router /admin/content/ai-generate [post]
func (h *Handler) AdminAIGenerateContent(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		utils.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	var req struct {
		Topic    string `json:"topic" binding:"required"`
		Type     string `json:"type"`     // "post" or "case", default "post"
		Language string `json:"language"` // "en" or "zh", default "en"
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.Type == "" {
		req.Type = "post"
	}
	if req.Language == "" {
		req.Language = "en"
	}

	langInstruction := "Write in English."
	if req.Language == "zh" {
		langInstruction = "用中文撰写。"
	}

	prompt := fmt.Sprintf(`You are a professional content writer for CandyPro, a candy OEM manufacturer.
Generate a blog article in Markdown format about: %s

%s

Requirements:
- Use Markdown with proper headings (##, ###), bullet points, bold text
- Include a compelling title (as # heading)
- Write an engaging excerpt (2-3 sentences)
- Article body should be 400-800 words
- Include relevant sections with subheadings
- Focus on candy manufacturing, OEM, food industry topics
- Professional B2B tone
- Suggest 3-5 relevant tags at the end as a comma-separated list after "Tags: "

Format the output exactly as:
# [Title]

[Excerpt paragraph]

## [Section 1]
...

## [Section 2]
...

Tags: tag1, tag2, tag3`, req.Topic, langInstruction)

	result, err := h.aiService.Generate(c.Request.Context(), prompt)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "ai_error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"content": result,
		"topic":   req.Topic,
		"type":    req.Type,
	})
}
