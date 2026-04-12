package system

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	modelsProduct "candypro/api/internal/models/product"
	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/roles"

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
	c.JSON(http.StatusServiceUnavailable, modelsProduct.ErrorResponse{
		Error:   "service_unavailable",
		Message: "AI service is not configured",
	})
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
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "prompt is required",
		})
		return
	}

	reply, err := h.aiService.Generate(c.Request.Context(), prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate chatbot response",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reply":       reply,
		"input":       prompt,
		"generatedAt": time.Now(),
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
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	inquiryText := strings.TrimSpace(req.InquiryText)
	if inquiryText == "" {
		inquiryText = strings.TrimSpace(req.Prompt)
	}
	if inquiryText == "" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "inquiryText is required",
		})
		return
	}

	targetCountry := strings.TrimSpace(req.TargetCountry)
	if targetCountry == "" {
		targetCountry = "global"
	}

	analysis, err := h.aiService.AnalyzeInquiry(c.Request.Context(), inquiryText, targetCountry)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to analyze inquiry",
		})
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
		Prompt               string `json:"prompt"`
		InquiryID            string `json:"inquiryId"`
		Currency             string `json:"currency"`
		CustomerRequirements string `json:"customerRequirements"`
		TargetCountry        string `json:"targetCountry"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		if strings.TrimSpace(req.CustomerRequirements) == "" {
			c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
				Error:   "invalid_request",
				Message: "prompt or customerRequirements is required",
			})
			return
		}
		prompt = fmt.Sprintf(
			"Generate a B2B candy OEM quotation draft in %s for inquiry %s. Target country: %s. Requirements: %s. Include pricing assumptions, MOQ, lead time, payment terms, validity and exclusions.",
			currency,
			strings.TrimSpace(req.InquiryID),
			strings.TrimSpace(req.TargetCountry),
			strings.TrimSpace(req.CustomerRequirements),
		)
	}

	quotation, err := h.aiService.Generate(c.Request.Context(), prompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to generate quotation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"quotation":   quotation,
		"currency":    currency,
		"generatedAt": time.Now(),
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
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	sourceLang := strings.TrimSpace(req.SourceLang)
	if sourceLang == "" {
		sourceLang = "auto"
	}

	translationPrompt := fmt.Sprintf(
		"Translate the following text from %s to %s. Return only the translated content without explanations:\n\n%s",
		sourceLang,
		strings.TrimSpace(req.TargetLang),
		strings.TrimSpace(req.Text),
	)

	translated, err := h.aiService.Generate(c.Request.Context(), translationPrompt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to translate content",
		})
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
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "prompt is required",
		})
		return
	}
	if h.services == nil || h.services.Search == nil {
		c.JSON(http.StatusServiceUnavailable, modelsProduct.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Search service is unavailable",
		})
		return
	}

	searchRes, err := h.services.Search.Search(c.Request.Context(), prompt, "products", 6)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to search products",
		})
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
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "query is required",
		})
		return
	}
	if h.services == nil || h.services.Search == nil {
		c.JSON(http.StatusServiceUnavailable, modelsProduct.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Search service is unavailable",
		})
		return
	}

	results, err := h.services.Search.Search(c.Request.Context(), prompt, "all", 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to perform AI search",
		})
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
		c.JSON(http.StatusServiceUnavailable, modelsProduct.ErrorResponse{
			Error:   "service_unavailable",
			Message: "Trade service is unavailable",
		})
		return
	}

	idStr := strings.TrimSpace(c.Param("id"))
	tradeID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "invalid conversation id",
		})
		return
	}

	transaction, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(tradeID))
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Trade transaction not found",
		})
		return
	}

	currentUserID, currentRole := authContext(c)
	if currentRole != roles.Admin && currentRole != roles.SuperAdmin && transaction.UserID != currentUserID {
		c.JSON(http.StatusForbidden, modelsProduct.ErrorResponse{
			Error:   "forbidden",
			Message: "You do not have access to this conversation",
		})
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
