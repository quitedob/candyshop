package system

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	modelsAuth "candypro/api/internal/models/auth"
	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func extractPrompt(c *gin.Context) string {
	var req struct {
		Prompt  string `json:"prompt"`
		Message string `json:"message"`
		Content string `json:"content"`
		Query   string `json:"query"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		return ""
	}

	const maxPromptLen = 4000 // ~1000 tokens, prevents cost abuse

	truncate := func(s string) string {
		if len(s) > maxPromptLen {
			return s[:maxPromptLen]
		}
		return s
	}

	switch {
	case strings.TrimSpace(req.Prompt) != "":
		return truncate(strings.TrimSpace(req.Prompt))
	case strings.TrimSpace(req.Message) != "":
		return truncate(strings.TrimSpace(req.Message))
	case strings.TrimSpace(req.Content) != "":
		return truncate(strings.TrimSpace(req.Content))
	default:
		return truncate(strings.TrimSpace(req.Query))
	}
}

func (h *Handler) aiUnavailable(c *gin.Context) bool {
	if h.aiService != nil && h.aiService.IsEnabled() {
		return false
	}
	response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
	return true
}

func authContext(c *gin.Context) (string, string) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("userRole")

	userIDStr, _ := userID.(string)
	roleStr, _ := role.(string)
	return strings.TrimSpace(userIDStr), strings.TrimSpace(roleStr)
}

// Chatbot replies to public AI chatbot request.
func (h *Handler) Chatbot(c *gin.Context) {
	if h.aiUnavailable(c) {
		return
	}

	prompt := extractPrompt(c)
	if prompt == "" {
		response.ErrorResp(c, http.StatusBadRequest, "ai_prompt_required")
		return
	}

	reply, err := h.aiService.Generate(c.Request.Context(), prompt)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_chatbot_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply": reply,
		// 不回传用户原文，降低隐私泄露与提示词复述风险；前端如需展示可仅用本地 state
		"generatedAt": time.Now(),
		"notice":      "AI-generated content is for general information only and is not legal, regulatory, or financial advice.",
	})
}

// AnalyzeInquiry analyzes inquiry text with destination context.
func (h *Handler) AnalyzeInquiry(c *gin.Context) {
	if h.aiUnavailable(c) {
		return
	}

	var req struct {
		InquiryText   string `json:"inquiryText"`
		TargetCountry string `json:"targetCountry"`
		Prompt        string `json:"prompt"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	inquiryText := strings.TrimSpace(req.InquiryText)
	if inquiryText == "" {
		inquiryText = strings.TrimSpace(req.Prompt)
	}
	if inquiryText == "" {
		response.ErrorResp(c, http.StatusBadRequest, "inquiry_text_required")
		return
	}

	targetCountry := strings.TrimSpace(req.TargetCountry)
	if targetCountry == "" {
		targetCountry = "global"
	}

	analysis, err := h.aiService.AnalyzeInquiry(c.Request.Context(), inquiryText, targetCountry)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_analyze_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"analysis":      analysis,
		"targetCountry": targetCountry,
		"generatedAt":   time.Now(),
	})
}

// GenerateQuotation generates structured quotation draft text.
func (h *Handler) GenerateQuotation(c *gin.Context) {
	if h.aiUnavailable(c) {
		return
	}

	var req struct {
		Prompt               string   `json:"prompt"`
		InquiryID            string   `json:"inquiryId"`
		Currency             string   `json:"currency"`
		CustomerRequirements string   `json:"customerRequirements"`
		TargetCountry        string   `json:"targetCountry"`
		ProductID            string   `json:"productId"`
		ProductIDs           []string `json:"productIds"`
		MarketCode           string   `json:"marketCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	if strings.TrimSpace(req.Prompt) == "" && strings.TrimSpace(req.CustomerRequirements) == "" {
		response.ErrorResp(c, http.StatusBadRequest, "prompt_or_requirements_required")
		return
	}

	// Fetch product cost stacks for AI drafting reference (non-binding)
	marketCode := strings.TrimSpace(req.MarketCode)
	if marketCode == "" {
		marketCode = strings.TrimSpace(req.TargetCountry)
	}
	if marketCode == "" {
		marketCode = "GLOBAL"
	}
	var costInject []gin.H
	var costStackInjection string
	if h.services != nil && h.services.Product != nil {
		ids := append([]string{}, req.ProductIDs...)
		if pid := strings.TrimSpace(req.ProductID); pid != "" {
			ids = append([]string{pid}, ids...)
		}
		seen := map[string]struct{}{}
		for _, raw := range ids {
			pid := strings.TrimSpace(raw)
			if pid == "" {
				continue
			}
			if _, dup := seen[pid]; dup {
				continue
			}
			seen[pid] = struct{}{}
			costJSON, costErr := h.services.Product.BuildPricingCostContextJSON(c.Request.Context(), pid, marketCode)
			if costErr != nil || costJSON == "" || costJSON == "{}" {
				continue
			}
			costInject = append(costInject, gin.H{"productId": pid, "costStackJSON": costJSON})
			costStackInjection += "\n\n---\nCost stack reference for product " + pid +
				" (read-only internal data, not a binding quote; drafting guidance only):\n" + costJSON
		}
	}

	quotation, err := h.aiService.GenerateQuotationText(
		c.Request.Context(),
		strings.TrimSpace(req.Prompt),
		currency,
		strings.TrimSpace(req.InquiryID),
		strings.TrimSpace(req.TargetCountry),
		strings.TrimSpace(req.CustomerRequirements),
		costStackInjection,
	)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_quotation_failed")
		return
	}

	pricingAssist := gin.H{
		"disclaimerZh":     "成本栈与模型输出仅供内部起草参考，不构成对客户的价格承诺，须人工复核。",
		"disclaimerEn":     "Cost stack and model output are non-binding drafting aids; human review is required before any customer commitment.",
		"injectedProducts": len(costInject),
	}
	if len(costInject) > 0 {
		pricingAssist["summaries"] = costInject
	}
	c.JSON(http.StatusOK, gin.H{
		"quotation":     quotation,
		"currency":      currency,
		"generatedAt":   time.Now(),
		"pricingAssist": pricingAssist,
	})
}

// Translate uses AI to translate content into target language.
func (h *Handler) Translate(c *gin.Context) {
	if h.aiUnavailable(c) {
		return
	}

	var req struct {
		Text       string `json:"text" binding:"required"`
		TargetLang string `json:"targetLang" binding:"required"`
		SourceLang string `json:"sourceLang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	translated, err := h.aiService.TranslateSingleText(
		c.Request.Context(),
		strings.TrimSpace(req.Text),
		strings.TrimSpace(req.SourceLang),
		strings.TrimSpace(req.TargetLang),
	)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_translate_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"text":       req.Text,
		"targetLang": req.TargetLang,
		"translated": strings.TrimSpace(translated),
	})
}

// RecommendProducts returns product matches from search and AI fit summary.
func (h *Handler) RecommendProducts(c *gin.Context) {
	prompt := extractPrompt(c)
	if prompt == "" {
		response.ErrorResp(c, http.StatusBadRequest, "ai_prompt_required")
		return
	}
	if h.services == nil || h.services.Search == nil {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_search_unavailable")
		return
	}

	searchRes, err := h.services.Search.Search(c.Request.Context(), prompt, "products", 6)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_search_failed")
		return
	}

	recommendations := make([]gin.H, 0, len(searchRes.Products))
	for _, p := range searchRes.Products {
		recommendations = append(recommendations, gin.H{
			"id":       p.ID,
			"slug":     p.Slug,
			"name":     p.Name,
			"summary":  p.Summary,
			"moq":      p.MOQ,
			"leadTime": p.LeadTime,
		})
	}

	aiSummary := ""
	if h.aiService != nil && h.aiService.IsEnabled() {
		aiSummary, _ = h.aiService.SuggestProducts(c.Request.Context(), "global", prompt)
	}

	c.JSON(http.StatusOK, gin.H{
		"query":           prompt,
		"recommendations": recommendations,
		"aiSummary":       strings.TrimSpace(aiSummary),
	})
}

// AISearch searches products/posts/cases.
func (h *Handler) AISearch(c *gin.Context) {
	prompt := extractPrompt(c)
	if prompt == "" {
		response.ErrorResp(c, http.StatusBadRequest, "query_required")
		return
	}
	if h.services == nil || h.services.Search == nil {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_search_unavailable")
		return
	}

	results, err := h.services.Search.Search(c.Request.Context(), prompt, "all", 10)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_search_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   prompt,
		"results": results,
	})
}

// GetConversation returns trade-centric conversation context.
func (h *Handler) GetConversation(c *gin.Context) {
	if h.services == nil || h.services.Trade == nil {
		response.ErrorResp(c, http.StatusServiceUnavailable, "service_not_configured")
		return
	}

	idStr := strings.TrimSpace(c.Param("id"))
	tradeID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_conversation_id")
		return
	}

	transaction, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(tradeID))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return
	}

	currentUserID, currentRole := authContext(c)
	if currentRole != modelsAuth.Admin && currentRole != modelsAuth.SuperAdmin && transaction.UserID != currentUserID {
		response.ErrorResp(c, http.StatusForbidden, "trade_conversation_no_access")
		return
	}

	messages := []gin.H{
		{
			"role":      "system",
			"content":   fmt.Sprintf("Trade %d is in %s status with currency %s and terms %s.", transaction.ID, transaction.Status, transaction.Currency, transaction.Terms),
			"timestamp": transaction.UpdatedAt,
		},
	}

	for _, doc := range transaction.Documents {
		messages = append(messages, gin.H{
			"role":      "assistant",
			"content":   fmt.Sprintf("Document %s (%s) is %s.", doc.DocNumber, doc.Type, doc.Status),
			"timestamp": doc.UpdatedAt,
		})
	}

	if len(transaction.Documents) == 0 {
		messages = append(messages, gin.H{
			"role":      "assistant",
			"content":   "No trade documents generated yet.",
			"timestamp": transaction.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id": idStr,
		"transaction": gin.H{
			"id":        transaction.ID,
			"reference": transaction.Reference,
			"status":    transaction.Status,
			"currency":  transaction.Currency,
			"terms":     transaction.Terms,
			"userId":    transaction.UserID,
			"docCount":  len(transaction.Documents),
		},
		"messages": messages,
		"summary": gin.H{
			"documentTypes": collectDocumentTypes(transaction.Documents),
			"lastUpdatedAt": transaction.UpdatedAt,
		},
	})
}

func collectDocumentTypes(docs []modelsTrade.TradeDocument) []string {
	seen := make(map[string]struct{}, len(docs))
	types := make([]string, 0, len(docs))
	for _, doc := range docs {
		if _, ok := seen[doc.Type]; ok {
			continue
		}
		seen[doc.Type] = struct{}{}
		types = append(types, doc.Type)
	}
	return types
}
