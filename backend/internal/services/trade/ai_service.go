package trade

import (
	"context"
	"fmt"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/eino"

	"github.com/cloudwego/eino/adk"
)

// ErrAIServiceDisabled is an alias kept for backward compatibility.
var ErrAIServiceDisabled = eino.ErrDisabled

// AIService provides domain-aware AI capabilities for B2B candy trade.
// Core chat model management and generation are delegated to the embedded eino.Client.
type AIService struct {
	*eino.Client
}

// NewAIService creates a new AI service wrapping the core eino.Client.
func NewAIService(cfg config.AIConfig) (*AIService, error) {
	client, err := eino.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &AIService{Client: client}, nil
}

// NewAIServiceWithAgent creates an AI service pre-wired with the full TradeAgent.
func NewAIServiceWithAgent(cfg config.AIConfig, agent adk.Agent) (*AIService, error) {
	client, err := eino.NewClientWithAgent(cfg, agent)
	if err != nil {
		return nil, err
	}
	return &AIService{Client: client}, nil
}

// AnalyzeInquiry uses AI to analyze customer inquiries with country context.
func (s *AIService) AnalyzeInquiry(ctx context.Context, inquiryText string, targetCountry string) (string, error) {
	countryContext := getCountryContext(targetCountry)

	prompt := fmt.Sprintf(`Analyze the following customer inquiry for a B2B candy manufacturer.

Target Market: %s

Regional Considerations:
- Regulations: %s
- Market Preferences: %s
- Shipping Notes: %s
- Required Certifications: %s

Provide analysis on:
1. Product-matching for this market
2. Quantity estimates and pricing tier
3. Urgency level and recommended response time
4. OEM requirements and certifications needed
5. Potential concerns or questions to clarify
6. Recommended follow-up actions

Customer inquiry:
%s`, targetCountry, countryContext.Regulations, countryContext.MarketPrefs,
		countryContext.ShippingNotes, countryContext.Certifications, inquiryText)

	return s.Generate(ctx, prompt)
}

// GetMarketInsights returns market-specific insights for a country.
func (s *AIService) GetMarketInsights(ctx context.Context, country string) (string, error) {
	prompt := fmt.Sprintf(`Provide B2B candy market insights for %s:
1. Market size and growth trends
2. Popular product categories
3. Key competitors
4. Price sensitivity
5. Seasonal patterns
6. Distribution channels
7. Regulatory landscape`, country)
	return s.Generate(ctx, prompt)
}

// SuggestProducts recommends products based on country and requirements.
func (s *AIService) SuggestProducts(ctx context.Context, country string, requirements string) (string, error) {
	countryContext := getCountryContext(country)
	prompt := fmt.Sprintf(`Suggest CandyPro products for:
Country: %s
Market Context: %s
Customer Requirements: %s

Provide:
1. Top 3 recommended products with reasons
2. Suggested modifications for market fit
3. Certification requirements
4. Estimated pricing tier
5. Potential concerns`, country, countryContext.MarketPrefs, requirements)
	return s.Generate(ctx, prompt)
}

// GetShippingOptions returns shipping information for a country.
func (s *AIService) GetShippingOptions(ctx context.Context, country string) (string, error) {
	countryContext := getCountryContext(country)
	prompt := fmt.Sprintf(`Provide B2B candy shipping options and logistics details for %s:
1. Available shipping methods (sea, air, courier) — do NOT quote specific transit day counts; state that transit varies by method, season, and destination and is confirmed at order time
2. Import duties and taxes (%s)
3. Customs clearance process
4. Temperature control requirements during transit
5. Required shipping documentation`, country, countryContext.ShippingNotes)
	return s.Generate(ctx, prompt)
}

// GetRegulatoryInfo returns regulatory compliance information for a country.
func (s *AIService) GetRegulatoryInfo(ctx context.Context, country string) (string, error) {
	countryContext := getCountryContext(country)

	var prompt string
	if s.ComplianceRetriever() == nil {
		prompt = fmt.Sprintf(`Provide comprehensive B2B candy regulatory compliance information for %s:
1. Food safety authority and registration process
2. Specific regulations (%s)
3. Mandatory certifications (%s)
4. Ingredient restrictions and permitted additives
5. Labeling requirements (language, nutritional facts panel, warnings)
6. Packaging material regulations`, country, countryContext.Regulations, countryContext.Certifications)
	} else {
		prompt = fmt.Sprintf(`Prepare a country compliance brief for B2B candy exports to %s.

Baseline context:
- Regulations: %s
- Certifications: %s

You MUST call the "compliance_lookup" tool first using:
- country: "%s"
- query: "food safety authority, registration, certifications, labeling, additives, packaging"

Then provide:
1. Food authority and registration steps
2. Mandatory certifications
3. Ingredient/additive restrictions
4. Labeling and language requirements
5. Packaging constraints

Requirements:
- Use tool results as primary evidence.
- Cite source filenames from the tool output for each section.
- If no exact tool match exists, clearly state that the local corpus has no exact hit and provide conservative guidance.`, country, countryContext.Regulations, countryContext.Certifications, country)
	}

	return s.Generate(ctx, prompt)
}

