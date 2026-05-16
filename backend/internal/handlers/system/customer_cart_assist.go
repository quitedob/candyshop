package system

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"
	tradeSvc "candypro/api/internal/services/trade"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type cartRecommendRequest struct {
	Prompt        string  `json:"prompt" binding:"required"`
	TargetCountry string  `json:"targetCountry"`
	Quantity      int     `json:"quantity"`
	Budget        float64 `json:"budget"`
	Currency      string  `json:"currency"`
}

// CustomerAIRecommendForCart returns AI product recommendations as JSON without creating an order.
// POST /user/ai/recommend
func (h *Handler) CustomerAIRecommendForCart(c *gin.Context) {
	if h.aiUnavailable(c) {
		return
	}
	if h.services == nil || h.services.Search == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req cartRecommendRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	targetCountry := strings.TrimSpace(req.TargetCountry)
	if targetCountry == "" {
		targetCountry = "global"
	}
	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}
	quantity := req.Quantity
	if quantity < 1 {
		quantity = 1
	}

	// Search for candidate products
	searchRes, err := h.services.Search.Search(c.Request.Context(), req.Prompt, "products", 10)
	if err != nil || len(searchRes.Products) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"recommendations": []any{},
			"summary":         "No matching products found. Try different keywords.",
			"warnings":        []string{},
		})
		return
	}

	candidateJSON, _ := json.Marshal(buildCandidatePayload(searchRes.Products))

	draft, err := h.aiService.CartRecommend(c.Request.Context(), tradeSvc.CartRecommendParams{
		Prompt:        req.Prompt,
		TargetCountry: targetCountry,
		Quantity:      quantity,
		Budget:        req.Budget,
		Currency:      currency,
		CandidateJSON: string(candidateJSON),
	})
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_recommendation_failed")
		return
	}

	// Build enriched result with full product info
	candidateByID := make(map[string]modelsProduct.Product, len(searchRes.Products))
	for _, p := range searchRes.Products {
		candidateByID[p.ID] = p
	}

	results := make([]gin.H, 0, len(draft.RecommendedProducts))
	for _, rec := range draft.RecommendedProducts {
		pid := strings.TrimSpace(rec.ProductID)
		if pid == "" {
			continue
		}
		product, ok := candidateByID[pid]
		if !ok {
			continue
		}
		qty := rec.Quantity
		if qty < 1 {
			qty = quantity
		}
		if product.MOQ > 0 && qty < product.MOQ {
			qty = product.MOQ
		}
		results = append(results, gin.H{
			"id":             product.ID,
			"name":           product.Name,
			"slug":           product.Slug,
			"thumbnail":      product.Thumbnail,
			"category":       product.Category,
			"moq":            product.MOQ,
			"leadTime":       product.LeadTime,
			"halalCertified": product.HalalCertified,
			"oemAvailable":   product.OEMAvailable,
			"quantity":       qty,
			"unitPrice":      rec.UnitPrice,
			"specifications": strings.TrimSpace(rec.Specifications),
			"reason":         strings.TrimSpace(rec.Reason),
		})
	}

	// Fallback: if AI returned nothing, use top search results
	if len(results) == 0 && len(searchRes.Products) > 0 {
		for _, p := range searchRes.Products[:min(3, len(searchRes.Products))] {
			qty := quantity
			if p.MOQ > 0 && qty < p.MOQ {
				qty = p.MOQ
			}
			results = append(results, gin.H{
				"id":             p.ID,
				"name":           p.Name,
				"slug":           p.Slug,
				"thumbnail":      p.Thumbnail,
				"category":       p.Category,
				"moq":            p.MOQ,
				"leadTime":       p.LeadTime,
				"halalCertified": p.HalalCertified,
				"oemAvailable":   p.OEMAvailable,
				"quantity":       qty,
				"unitPrice":      0,
				"specifications": "",
				"reason":         "Best match based on your search query.",
			})
		}
		draft.Summary = "Showing top search results. AI recommendation was not available."
	}

	c.JSON(http.StatusOK, gin.H{
		"recommendations": results,
		"summary":         strings.TrimSpace(draft.Summary),
		"warnings":        dedupeNonEmpty(draft.Warnings),
	})
}
