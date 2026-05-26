package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ──────────────────────────────────────────────────────────────
// Structured generation: prompts + JSON parsing live here.
// Domain services in services/trade/ delegate here.
// ──────────────────────────────────────────────────────────────

// ProductGenResult is the structured AI output for product generation.
type ProductGenResult struct {
	Name                string   `json:"name"`
	Slug                string   `json:"slug"`
	Summary             string   `json:"summary"`
	Description         string   `json:"description"`
	Category            string   `json:"category"`
	CategorySlug        string   `json:"categorySlug"`
	LeadTime            string   `json:"leadTime"`
	Ingredients         string   `json:"ingredients"`
	Allergens           string   `json:"allergens"`
	ShelfLife           string   `json:"shelfLife"`
	Storage             string   `json:"storage"`
	SweetenerType       string   `json:"sweetenerType"`
	GMOStatus           string   `json:"gmoStatus"`
	PrimaryPackaging    string   `json:"primaryPackaging"`
	InnerPackConfig     string   `json:"innerPackConfig"`
	PalletConfig        string   `json:"palletConfig"`
	SampleLeadTime      string   `json:"sampleLeadTime"`
	GTIN                string   `json:"gtin"`
	HSCode              string   `json:"hsCode"`
	Flavors             []string `json:"flavors"`
	Shapes              []string `json:"shapes"`
	Certifications      []string `json:"certifications"`
	Additives           []string `json:"additives"`
	MayContain          []string `json:"mayContain"`
	MOQ                 int      `json:"moq"`
	BasePrice           float64  `json:"basePrice"`
	StockQuantity       int      `json:"stockQuantity"`
	NetWeightPerPiece   float64  `json:"netWeightPerPiece"`
	NetWeightPerPack    float64  `json:"netWeightPerPack"`
	GrossWeightPerCarton float64 `json:"grossWeightPerCarton"`
	PiecesPerPack       int      `json:"piecesPerPack"`
	PacksPerCarton      int      `json:"packsPerCarton"`
	ProductLengthMM     float64  `json:"productLengthMM"`
	ProductWidthMM      float64  `json:"productWidthMM"`
	ProductHeightMM     float64  `json:"productHeightMM"`
	EnergyKj            float64  `json:"energyKj"`
	EnergyKcal          float64  `json:"energyKcal"`
	TotalFatG           float64  `json:"totalFatG"`
	SaturatedFatG       float64  `json:"saturatedFatG"`
	CarbohydratesG      float64  `json:"carbohydratesG"`
	SugarsG             float64  `json:"sugarsG"`
	ProteinG            float64  `json:"proteinG"`
	SaltG               float64  `json:"saltG"`
	FiberG              float64  `json:"fiberG"`
	CocoaSolidsPct      float64  `json:"cocoaSolidsPct"`
	MilkSolidsPct       float64  `json:"milkSolidsPct"`
	WaterActivity       float64  `json:"waterActivity"`
	SampleMOQ           int      `json:"sampleMOQ"`
	SamplePrice         float64  `json:"samplePrice"`
	OEMAvailable        bool     `json:"oemAvailable"`
	HalalCertified      bool     `json:"halalCertified"`
	Featured            bool     `json:"featured"`
	IsVegan             bool     `json:"isVegan"`
	IsGlutenFree        bool     `json:"isGlutenFree"`
	IsSugarFree         bool     `json:"isSugarFree"`
	IsKosher            bool     `json:"isKosher"`
	IsOrganic           bool     `json:"isOrganic"`
}

// ContentGenResult is the structured AI output for blog/case-study generation.
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
	// case-study specific
	Client    string   `json:"client,omitempty"`
	Industry  string   `json:"industry,omitempty"`
	Location  string   `json:"location,omitempty"`
	Timeline  string   `json:"timeline,omitempty"`
	Challenge string   `json:"challenge,omitempty"`
	Solution  string   `json:"solution,omitempty"`
	Result    string   `json:"result,omitempty"`
	Services  []string `json:"services,omitempty"`
}

// GenerateProduct produces a fully-populated ProductGenResult from a free-text description.
func (c *Client) GenerateProduct(ctx context.Context, description, language string) (*ProductGenResult, string, error) {
	prompt := buildProductGenPrompt(description, language)
	raw, err := c.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, "", fmt.Errorf("product generation: %w", err)
	}

	raw = cleanJSONBlock(raw)
	var result ProductGenResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, raw, nil
	}

	if result.Slug == "" && result.Name != "" {
		result.Slug = slugify(result.Name)
	}
	if result.CategorySlug == "" && result.Category != "" {
		result.CategorySlug = slugify(result.Category)
	}

	return &result, "", nil
}

// ContentReviseResult is the structured AI output for partial draft revision.
type ContentReviseResult struct {
	Updates map[string]json.RawMessage `json:"updates"`
	Summary string                     `json:"summary"`
}

// InlineEditResult is the AI output for Cursor-style selection replacement.
type InlineEditResult struct {
	Replacement string `json:"replacement"`
	Format        string `json:"format"` // html or plain
}

// InlineEditContent rewrites only the selected excerpt per user instruction.
func (c *Client) InlineEditContent(ctx context.Context, language, instruction, selectedText, selectedHTML, contextBefore, contextAfter, fieldType string) (*InlineEditResult, string, error) {
	prompt := buildInlineEditPrompt(language, instruction, selectedText, selectedHTML, contextBefore, contextAfter, fieldType)
	raw, err := c.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, "", fmt.Errorf("inline edit: %w", err)
	}

	raw = cleanJSONBlock(raw)
	var result InlineEditResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, raw, nil
	}
	if result.Format == "" {
		if fieldType == "html" {
			result.Format = "html"
		} else {
			result.Format = "plain"
		}
	}
	return &result, "", nil
}

// ReviseContent applies user feedback to an existing draft, changing only relevant fields/sections.
func (c *Client) ReviseContent(ctx context.Context, contentType, language, instruction, selectedText string, focusFields []string, current map[string]string) (*ContentReviseResult, string, error) {
	prompt := buildContentRevisePrompt(contentType, language, instruction, selectedText, focusFields, current)
	raw, err := c.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, "", fmt.Errorf("content revision: %w", err)
	}

	raw = cleanJSONBlock(raw)
	var result ContentReviseResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, raw, nil
	}
	if result.Updates == nil {
		result.Updates = map[string]json.RawMessage{}
	}
	return &result, "", nil
}

// GenerateContent produces a structured ContentGenResult from a topic string.
func (c *Client) GenerateContent(ctx context.Context, topic, contentType, language string) (*ContentGenResult, string, error) {
	langInstruction := localeInstruction(language)

	var prompt string
	if contentType == "case" {
		prompt = buildCaseGenPrompt(topic, langInstruction)
	} else {
		prompt = buildPostGenPrompt(topic, langInstruction)
	}

	raw, err := c.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, "", fmt.Errorf("content generation: %w", err)
	}

	raw = cleanJSONBlock(raw)
	var result ContentGenResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, raw, nil
	}

	if result.Slug == "" && result.Title != "" {
		result.Slug = slugify(result.Title)
	}
	if contentType == "post" && result.ReadTime == 0 && result.Content != "" {
		result.ReadTime = max(1, wordCount(stripHTML(result.Content))/200)
	}
	if contentType == "post" {
		result.Category = normalizeBlogCategory(result.Category)
	}

	return &result, "", nil
}

// NormalizeBlogCategory maps free-text or AI labels to public blog category slugs.
func NormalizeBlogCategory(raw string) string {
	return normalizeBlogCategory(raw)
}

// ProductFieldsForTranslation returns the text fields of a product that should be translated.
func ProductFieldsForTranslation(p *ProductGenResult) map[string]string {
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

// ContentFieldsForTranslation returns the text fields of generated content for translation.
func ContentFieldsForTranslation(c *ContentGenResult) map[string]string {
	fields := map[string]string{
		"title":   c.Title,
		"excerpt": c.Excerpt,
		"content": c.Content,
	}
	if c.AuthorName != "" {
		fields["authorName"] = c.AuthorName
	}
	if c.AuthorTitle != "" {
		fields["authorTitle"] = c.AuthorTitle
	}
	if c.AuthorBio != "" {
		fields["authorBio"] = c.AuthorBio
	}
	if c.Challenge != "" {
		fields["challenge"] = c.Challenge
	}
	if c.Solution != "" {
		fields["solution"] = c.Solution
	}
	if c.Result != "" {
		fields["result"] = c.Result
	}
	return fields
}

// SupportedLocales is the ordered list of all supported translation locales.
var SupportedLocales = []string{"en", "zh", "ko", "ar", "ja", "th", "vi", "id", "ms"}

var localeNames = map[string]string{
	"en": "English",
	"zh": "Chinese (Simplified)",
	"ko": "Korean",
	"ar": "Arabic",
	"ja": "Japanese",
	"th": "Thai",
	"vi": "Vietnamese",
	"id": "Indonesian",
	"ms": "Malay",
}

func localeInstruction(language string) string {
	name, ok := localeNames[language]
	if !ok {
		name = localeNames["en"]
	}
	return "Write in " + name + "."
}

// LocalesExcluding returns SupportedLocales without the given locale.
func LocalesExcluding(locale string) []string {
	var out []string
	for _, l := range SupportedLocales {
		if l != locale {
			out = append(out, l)
		}
	}
	return out
}

// ── prompt builders ──────────────────────────────────────────

func buildProductGenPrompt(description, language string) string {
	langInstruction := localeInstruction(language)
	return `You are a product data specialist for CandyPro, a candy OEM manufacturer.
Given a natural-language description, generate a COMPLETE JSON object with ALL product fields filled with realistic, industry-appropriate values.

Description: ` + description + `

` + langInstruction + `

Return ONLY a valid JSON object (no markdown fences, no extra text) with exactly these fields:
{
  "name": "Product name",
  "slug": "url-friendly-slug",
  "summary": "2-3 sentence product summary highlighting key selling points",
  "description": "Detailed product description, 100-200 words",
  "category": "Product category (e.g., Hard Candy, Gummies, Chocolate, Lollipops, Toffee, Marshmallow, Jelly, Biscuit, Snack)",
  "categorySlug": "url-friendly-category-slug",
  "leadTime": "Production lead time — set by admin per product or stated in order confirmation (do not invent week/month estimates)",
  "ingredients": "Full ingredients list as a single string",
  "allergens": "Known allergens (e.g., 'Contains milk and soy.')",
  "shelfLife": "Shelf life (e.g., '12 months')",
  "storage": "Storage instructions (e.g., 'Store in a cool, dry place below 25°C')",
  "sweetenerType": "Sweetener (e.g., 'Sugar', 'Maltitol', 'Stevia', 'None')",
  "gmoStatus": "One of: 'Non-GMO', 'GMO', 'GMO-Free Certified', or empty string",
  "primaryPackaging": "Primary packaging: flow-wrap, foil, box, bag, jar, blister, tin",
  "innerPackConfig": "Inner pack config (e.g., '12 units per display box')",
  "palletConfig": "Pallet config (e.g., '48 cases/layer × 5 layers')",
  "sampleLeadTime": "Sample lead time — set by admin or stated in order confirmation",
  "gtin": "A realistic 13-digit EAN-13",
  "hsCode": "HS code for candy (e.g., '1704.90', '1806.32')",
  "flavors": ["Flavor1", "Flavor2"],
  "shapes": ["Shape1"],
  "certifications": ["Certification"],
  "additives": ["E-number or additive"],
  "mayContain": ["Cross-contaminant allergen"],
  "moq": minimum_order_quantity_units,
  "basePrice": unit_price_USD,
  "stockQuantity": typical_stock,
  "netWeightPerPiece": grams_per_piece,
  "netWeightPerPack": grams_per_pack,
  "grossWeightPerCarton": kg_per_carton,
  "piecesPerPack": pieces_per_pack,
  "packsPerCarton": packs_per_carton,
  "productLengthMM": length_mm,
  "productWidthMM": width_mm,
  "productHeightMM": height_mm,
  "energyKj": kJ_per_100g,
  "energyKcal": kcal_per_100g,
  "totalFatG": g_per_100g,
  "saturatedFatG": g_per_100g,
  "carbohydratesG": g_per_100g,
  "sugarsG": g_per_100g,
  "proteinG": g_per_100g,
  "saltG": g_per_100g,
  "fiberG": g_per_100g,
  "cocoaSolidsPct": 0_to_100_for_chocolate,
  "milkSolidsPct": 0_to_100_for_milk_chocolate,
  "waterActivity": 0_to_1,
  "sampleMOQ": sample_min_order,
  "samplePrice": sample_price_USD,
  "oemAvailable": true_or_false,
  "halalCertified": true_or_false,
  "featured": true_or_false,
  "isVegan": true_or_false,
  "isGlutenFree": true_or_false,
  "isSugarFree": true_or_false,
  "isKosher": true_or_false,
  "isOrganic": true_or_false
}

Make nutrition/weight/dimension values realistic for the candy type. Always fill ALL fields — use 0 for non-applicable numerics, empty arrays for unused arrays.`
}

func buildPostGenPrompt(topic, langInstruction string) string {
	return `You are a professional content writer for CandyPro, a candy OEM manufacturer.
The administrator's brief below may be written in ANY language — read and understand it completely, then write the article in the language specified below.

Brief: ` + topic + `

` + langInstruction + `

Return ONLY a valid JSON object (no markdown fences, no extra text) with exactly these fields:
{
  "title": "Compelling blog post title",
  "slug": "url-friendly-slug-derived-from-title",
  "content": "Full article body in HTML format, 400-800 words, using <h2>, <h3>, <p>, <ul>, <li>, <strong> — no <h1>",
  "excerpt": "Engaging 2-3 sentence excerpt summarizing the article",
  "category": "Exactly one of: compliance | product_knowledge | packaging | market_insights",
  "readTime": estimated_minutes_as_integer,
  "tags": ["tag1", "tag2", "tag3", "tag4", "tag5"],
  "authorName": "Suggested author name",
  "authorTitle": "Suggested author title at CandyPro",
  "authorBio": "Short author bio (1-2 sentences)"
}`
}

func buildCaseGenPrompt(topic, langInstruction string) string {
	return `You are a professional content writer for CandyPro, a candy OEM manufacturer.
The administrator's brief below may be written in ANY language — read and understand it completely, then write the case study in the language specified below.

Brief: ` + topic + `

` + langInstruction + `

Return ONLY a valid JSON object (no markdown fences, no extra text) with exactly these fields:
{
  "title": "Compelling case study title",
  "slug": "url-friendly-slug-derived-from-title",
  "content": "Full case study body in plain text, 300-600 words, with narrative flow: background, approach, outcome",
  "client": "Client company name",
  "industry": "Client industry (e.g., Food & Beverage, Retail, Confectionery)",
  "location": "Client location (city, country)",
  "timeline": "Project timeline — provided by sales team in quotation or order confirmation only (do not invent week/month estimates)",
  "challenge": "The client's challenge or problem (2-3 sentences)",
  "solution": "CandyPro's solution and approach (2-3 sentences)",
  "result": "Measurable results and benefits achieved (2-3 sentences)",
  "services": ["OEM Production", "Custom Formulation", "Packaging Design"]
}`
}

func buildInlineEditPrompt(language, instruction, selectedText, selectedHTML, contextBefore, contextAfter, fieldType string) string {
	langInstruction := localeInstruction(language)
	formatHint := "Return plain text only — no HTML tags, no markdown fences."
	formatVal := "plain"
	switch fieldType {
	case "html":
		formatHint = "Return an HTML fragment only (e.g. <p>, <ul>, <strong>). No <html> wrapper, no markdown fences."
		formatVal = "html"
	case "title":
		formatHint = "Return a single-line title string — no HTML, no quotes wrapper."
	}

	selectedBlock := selectedText
	if fieldType == "html" && strings.TrimSpace(selectedHTML) != "" {
		selectedBlock = selectedHTML
	}

	return `You are a Cursor-style inline editing assistant for CandyPro CMS content.
Rewrite ONLY the selected excerpt according to the instruction. Do NOT return the full document.

` + langInstruction + `

Context before selection:
` + contextBefore + `

SELECTED EXCERPT (replace this only):
` + selectedBlock + `

Context after selection:
` + contextAfter + `

Instruction (any language):
` + instruction + `

Rules:
1. Return ONLY the replacement for the selected excerpt — same format and comparable length unless instruction asks otherwise.
2. ` + formatHint + `
3. Preserve factual accuracy and terminology from surrounding context.
4. Do not include the unchanged context before/after in your output.

Return ONLY valid JSON:
{
  "replacement": "<replacement text only>",
  "format": "` + formatVal + `"
}`
}

func buildContentRevisePrompt(contentType, language, instruction, selectedText string, focusFields []string, current map[string]string) string {
	langInstruction := localeInstruction(language)
	currentJSON, _ := json.Marshal(current)

	focusHint := "Infer which fields to change from the administrator's instruction. Do NOT rewrite unchanged fields."
	if len(focusFields) > 0 {
		focusHint = "Only modify these fields unless the instruction clearly requires another: " + strings.Join(focusFields, ", ")
	}

	selectedBlock := ""
	if strings.TrimSpace(selectedText) != "" {
		selectedBlock = `
The administrator selected this excerpt from the draft — apply the revision primarily here:
"` + strings.TrimSpace(selectedText) + `"
When "content" is updated, return the COMPLETE content field with only the targeted section changed; keep all other paragraphs identical.
`
	}

	var allowedFields string
	if contentType == "case" {
		allowedFields = `Allowed keys in "updates": title, slug, content, client, industry, location, timeline, challenge, solution, result, services (array of strings).`
	} else {
		allowedFields = `Allowed keys in "updates": title, slug, excerpt, content, category (one of compliance|product_knowledge|packaging|market_insights), readTime (integer), tags (array of strings), authorName, authorTitle, authorBio.`
	}

	return `You are an editorial assistant for CandyPro CMS. The administrator is reviewing an AI-generated draft and wants targeted fixes — like Cursor inline edit — NOT a full rewrite.

` + langInstruction + `

Current draft (JSON):
` + string(currentJSON) + `

Revision instruction (may be in any language):
` + instruction + `
` + selectedBlock + `
` + focusHint + `
` + allowedFields + `

Rules:
1. Return ONLY changed fields inside "updates". Omit fields that stay the same.
2. Preserve tone, structure, and factual consistency with the rest of the draft.
3. If title changes materially, also update "slug".
4. If content changes, update "readTime" when appropriate (post only).
5. "summary" is one short sentence in Chinese explaining what you changed.

Return ONLY valid JSON (no markdown fences):
{
  "updates": { },
  "summary": "..."
}`
}

// ── shared helpers ───────────────────────────────────────────

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
	return strings.Map(func(r rune) rune {
		if r == '<' || r == '>' {
			return ' '
		}
		return r
	}, s)
}

func wordCount(s string) int {
	return len(strings.Fields(s))
}

// normalizeBlogCategory maps AI or free-text labels to public blog category slugs.
func normalizeBlogCategory(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "-", "_")
	switch s {
	case "compliance", "food_safety", "regulatory", "法规", "合规":
		return "compliance"
	case "product_knowledge", "product", "candy_manufacturing", "manufacturing", "产品知识":
		return "product_knowledge"
	case "packaging", "包装":
		return "packaging"
	case "market_insights", "market", "oem_trends", "trends", "insights", "市场洞察":
		return "market_insights"
	default:
		if s == "" {
			return "market_insights"
		}
		for _, slug := range []string{"compliance", "product_knowledge", "packaging", "market_insights"} {
			if strings.Contains(s, slug) {
				return slug
			}
		}
		return "market_insights"
	}
}
