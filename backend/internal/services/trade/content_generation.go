package trade

import (
	"context"
	"encoding/json"
	"strings"
)

// ContentGenResult is the structured AI response for blog post or case study content generation.
type ContentGenResult struct {
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

// GenerateContent uses AI to generate blog post or case study content with structured JSON output.
// contentType must be "post" or "case". language must be "en" or "zh".
func (s *AIService) GenerateContent(ctx context.Context, topic, contentType, language string) (*ContentGenResult, string, error) {
	langInstruction := "Write in English."
	if language == "zh" {
		langInstruction = "用中文撰写。"
	}

	var prompt string
	if contentType == "case" {
		prompt = buildCasePrompt(topic, langInstruction)
	} else {
		prompt = buildPostPrompt(topic, langInstruction)
	}

	raw, err := s.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, "", err
	}

	raw = cleanJSONBlock(raw)
	var result ContentGenResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		// Return raw text as fallback
		return nil, raw, nil
	}

	if result.Slug == "" && result.Title != "" {
		result.Slug = slugifyContent(result.Title)
	}
	if contentType == "post" && result.ReadTime == 0 && result.Content != "" {
		result.ReadTime = max(1, wordCountContent(stripHTMLContent(result.Content))/200)
	}

	return &result, "", nil
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

func cleanJSONBlock(s string) string {
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

func slugifyContent(title string) string {
	s := strings.ToLower(title)
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' {
			return r
		}
		return ' '
	}, s)
	return strings.Join(strings.Fields(s), "-")
}

func stripHTMLContent(s string) string {
	s = strings.Map(func(r rune) rune {
		if r == '<' || r == '>' {
			return ' '
		}
		return r
	}, s)
	return s
}

func wordCountContent(s string) int {
	return len(strings.Fields(s))
}
