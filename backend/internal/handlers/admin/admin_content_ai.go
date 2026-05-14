package admin

import (
	"encoding/json"
	"net/http"
	"strings"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// aiContentResult is the structured AI response for both post and case content generation.
type aiContentResult struct {
	Title       string   `json:"title"`
	Slug        string   `json:"slug"`
	Content     string   `json:"content"`
	Excerpt     string   `json:"excerpt,omitempty"`
	Category    string   `json:"category,omitempty"`
	ReadTime    int      `json:"readTime,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	AuthorName  string   `json:"authorName,omitempty"`
	AuthorTitle string   `json:"authorTitle,omitempty"`
	AuthorBio   string   `json:"authorBio,omitempty"`
	// case-specific
	Client    string   `json:"client,omitempty"`
	Industry  string   `json:"industry,omitempty"`
	Location  string   `json:"location,omitempty"`
	Timeline  string   `json:"timeline,omitempty"`
	Challenge string   `json:"challenge,omitempty"`
	Solution  string   `json:"solution,omitempty"`
	Result    string   `json:"result,omitempty"`
	Services  []string `json:"services,omitempty"`
}

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

	langInstruction := "Write in English."
	if req.Language == "zh" {
		langInstruction = "用中文撰写。"
	}

	var prompt string
	if req.Type == "case" {
		prompt = buildCasePrompt(req.Topic, langInstruction)
	} else {
		prompt = buildPostPrompt(req.Topic, langInstruction)
	}

	result, err := h.aiService.GenerateJSON(c.Request.Context(), prompt)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_error")
		return
	}

	result = cleanJSON(result)
	var content aiContentResult
	if err := json.Unmarshal([]byte(result), &content); err != nil {
		// Fallback: return raw text so the frontend can still use it
		c.JSON(http.StatusOK, gin.H{
			"content":    result,
			"topic":      req.Topic,
			"type":       req.Type,
			"parseError": true,
		})
		return
	}

	if content.Slug == "" && content.Title != "" {
		content.Slug = slugify(content.Title)
	}
	if req.Type == "post" && content.ReadTime == 0 && content.Content != "" {
		content.ReadTime = max(1, wordCount(stripHTML(content.Content))/200)
	}

	c.JSON(http.StatusOK, gin.H{
		"content": content,
		"topic":   req.Topic,
		"type":    req.Type,
	})
}

func buildPostPrompt(topic, langInstruction string) string {
	return `You are a professional content writer for CandyPro, a candy OEM manufacturer.
Generate a JSON object for a blog post about: ` + topic + `

` + langInstruction + `

Return ONLY a valid JSON object (no markdown fences, no extra text) with exactly these fields:
{
  "title": "Compelling blog post title",
  "slug": "url-friendly-slug-derived-from-title",
  "content": "Full article body in HTML format, 400-800 words, using <h2>, <h3>, <p>, <ul>, <li>, <strong> — no <h1>",
  "excerpt": "Engaging 2-3 sentence excerpt summarizing the article",
  "category": "Relevant category: Candy Manufacturing, OEM Trends, Food Safety, Market Insights, or Packaging",
  "readTime": estimated_minutes_as_integer,
  "tags": ["tag1", "tag2", "tag3", "tag4", "tag5"],
  "authorName": "Suggested author name",
  "authorTitle": "Suggested author title at CandyPro",
  "authorBio": "Short author bio (1-2 sentences)"
}`
}

func buildCasePrompt(topic, langInstruction string) string {
	return `You are a professional content writer for CandyPro, a candy OEM manufacturer.
Generate a JSON object for a case study about: ` + topic + `

` + langInstruction + `

Return ONLY a valid JSON object (no markdown fences, no extra text) with exactly these fields:
{
  "title": "Compelling case study title",
  "slug": "url-friendly-slug-derived-from-title",
  "content": "Full case study body in plain text, 300-600 words, with narrative flow: background, approach, outcome",
  "client": "Client company name",
  "industry": "Client industry (e.g., Food & Beverage, Retail, Confectionery)",
  "location": "Client location (city, country)",
  "timeline": "Project timeline (e.g., '3 months', 'Q1-Q2 2025')",
  "challenge": "The client's challenge or problem (2-3 sentences)",
  "solution": "CandyPro's solution and approach (2-3 sentences)",
  "result": "Measurable results and benefits achieved (2-3 sentences)",
  "services": ["OEM Production", "Custom Formulation", "Packaging Design"]
}`
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func slugify(title string) string {
	s := strings.ToLower(title)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' {
			return r
		}
		return ' '
	}, s)
	return strings.Join(strings.Fields(s), "-")
}

func stripHTML(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '<' || r == '>' {
			return ' '
		}
		return r
	}, s)
	return s
}

func wordCount(s string) int {
	return len(strings.Fields(s))
}
