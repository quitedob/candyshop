package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// OrderAssistParams holds the business parameters for AI order assistance.
type OrderAssistParams struct {
	Prompt                 string
	TargetCountry          string
	Quantity               int
	Budget                 float64
	Currency               string
	AdditionalRequirements string
	ShippingCountry        string
	CandidateJSON          string
}

// OrderAssistDraft is the AI-structured result for order assistance.
type OrderAssistDraft struct {
	RecommendedProducts  []OrderAssistItem `json:"recommendedProducts"`
	ComplianceChecklist  []string          `json:"complianceChecklist"`
	RequiredCertificates []string          `json:"requiredCertificates"`
	MissingInformation   []string          `json:"missingInformation"`
	Warnings             []string          `json:"warnings"`
	Summary              string            `json:"summary"`
}

// OrderAssistItem is a single product recommendation from AI.
type OrderAssistItem struct {
	ProductID      string  `json:"productId"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	Specifications string  `json:"specifications"`
	Reason         string  `json:"reason"`
}

// OrderAssistAnalyze calls AI to select products and produce a structured order draft.
func (s *AIService) OrderAssistAnalyze(ctx context.Context, params OrderAssistParams) (*OrderAssistDraft, error) {
	prompt := buildOrderAssistPrompt(params)

	raw, err := s.GenerateJSON(ctx, prompt)
	if err != nil {
		// Fallback to regular Generate
		raw, err = s.Generate(ctx, prompt)
		if err != nil {
			return nil, err
		}
	}

	return parseOrderAssistDraft(raw)
}

func buildOrderAssistPrompt(params OrderAssistParams) string {
	return fmt.Sprintf(`You are a cross-border B2B confectionery order assistant.

Select products from the provided candidate list and produce STRICT JSON only.
Do not output markdown. Do not include explanations outside JSON.

User request:
- prompt: %s
- targetCountry: %s
- requestedQuantity: %d
- budget: %.2f
- currency: %s
- additionalRequirements: %s
- shippingAddressCountry: %s

Candidate products JSON:
%s

Output schema:
{
  "recommendedProducts": [
    {
      "productId": "string",
      "quantity": 0,
      "unitPrice": 0,
      "specifications": "string",
      "reason": "string"
    }
  ],
  "complianceChecklist": ["string"],
  "requiredCertificates": ["string"],
  "missingInformation": ["string"],
  "warnings": ["string"],
  "summary": "string"
}

Rules:
1. Use only productId values from candidates.
2. quantity must be integer >= 1.
3. Keep output practical for import/export compliance in target country.
4. If budget likely insufficient, include a warning.`,
		strings.TrimSpace(params.Prompt),
		strings.TrimSpace(params.TargetCountry),
		params.Quantity,
		params.Budget,
		strings.TrimSpace(params.Currency),
		strings.TrimSpace(params.AdditionalRequirements),
		strings.TrimSpace(params.ShippingCountry),
		params.CandidateJSON,
	)
}

func parseOrderAssistDraft(raw string) (*OrderAssistDraft, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return &OrderAssistDraft{}, nil
	}

	var draft OrderAssistDraft
	if err := json.Unmarshal([]byte(text), &draft); err == nil {
		return &draft, nil
	}

	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return &OrderAssistDraft{}, nil
	}

	jsonPart := strings.TrimSpace(text[start : end+1])
	if err := json.Unmarshal([]byte(jsonPart), &draft); err != nil {
		return &OrderAssistDraft{}, nil
	}
	return &draft, nil
}

// CartRecommendParams holds the business parameters for AI cart recommendations.
type CartRecommendParams struct {
	Prompt        string
	TargetCountry string
	Quantity      int
	Budget        float64
	Currency      string
	CandidateJSON string
}

// CartRecommendDraft is the AI-structured result for cart recommendations.
type CartRecommendDraft struct {
	RecommendedProducts []CartRecommendItem `json:"recommendedProducts"`
	Summary             string              `json:"summary"`
	Warnings            []string            `json:"warnings"`
}

// CartRecommendItem is a single product recommendation from AI.
type CartRecommendItem struct {
	ProductID      string  `json:"productId"`
	Quantity       int     `json:"quantity"`
	UnitPrice      float64 `json:"unitPrice"`
	Specifications string  `json:"specifications"`
	Reason         string  `json:"reason"`
}

// CartRecommend calls AI to recommend products for the shopping cart.
func (s *AIService) CartRecommend(ctx context.Context, params CartRecommendParams) (*CartRecommendDraft, error) {
	prompt := fmt.Sprintf(`You are a B2B candy product selection assistant.
Select the best matching products from the candidates and return STRICT JSON only. No markdown, no explanation outside JSON.

User request:
- prompt: %s
- targetCountry: %s
- requestedQuantity: %d
- budget: %.2f %s

Candidate products:
%s

Output JSON schema:
{
  "recommendedProducts": [
    {
      "productId": "string",
      "quantity": 1,
      "unitPrice": 0,
      "specifications": "string",
      "reason": "one sentence why this product fits"
    }
  ],
  "summary": "brief overall recommendation summary",
  "warnings": ["optional warnings"]
}

Rules:
- Only use productId values from the candidates list.
- quantity must be >= 1 and respect MOQ.
- unitPrice is your best estimate; 0 if unknown.
- Select 1-5 most relevant products.`,
		strings.TrimSpace(params.Prompt),
		params.TargetCountry,
		params.Quantity,
		params.Budget,
		params.Currency,
		params.CandidateJSON,
	)

	raw, err := s.GenerateJSON(ctx, prompt)
	if err != nil {
		raw, err = s.Generate(ctx, prompt)
		if err != nil {
			return nil, err
		}
	}

	var draft CartRecommendDraft
	text := strings.TrimSpace(raw)
	if jsonErr := json.Unmarshal([]byte(text), &draft); jsonErr != nil {
		start := strings.Index(text, "{")
		end := strings.LastIndex(text, "}")
		if start >= 0 && end > start {
			if ue := json.Unmarshal([]byte(text[start:end+1]), &draft); ue != nil {
				log.Printf("order_assist: fallback JSON unmarshal failed: %v", ue)
			}
		}
	}

	return &draft, nil
}
