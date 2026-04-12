package trade

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/eino/prompts"
	"candypro/api/internal/pkg/eino/tool/rag"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

var ErrAIServiceDisabled = errors.New("ai service is not configured")

// AIService provides AI capabilities using Eino framework
type AIService struct {
	cfg                 config.AIConfig
	chatModel           *openai.ChatModel
	chatModelJSON       *openai.ChatModel // dedicated model with json_object response format
	runner              *adk.Runner
	complianceRetriever *rag.ComplianceRetriever
}

// NewAIService creates a new AI service
func NewAIService(cfg config.AIConfig) (*AIService, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, ErrAIServiceDisabled
	}

	ctx := context.Background()

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:  cfg.OpenAIModel,
		APIKey: cfg.OpenAIAPIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat model: %w", err)
	}

	// JSON-mode model: forces response_format = json_object
	jsonResponseFormat := openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONObject,
	}
	chatModelJSON, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:          cfg.OpenAIModel,
		APIKey:         cfg.OpenAIAPIKey,
		ResponseFormat: &jsonResponseFormat,
	})
	if err != nil {
		// Non-fatal: fall back to regular model
		log.Printf("Warning: failed to initialize JSON chat model, falling back: %v", err)
		chatModelJSON = chatModel
	}

	agentTools := make([]tool.BaseTool, 0, 1)
	var complianceRetriever *rag.ComplianceRetriever

	complianceTool, retriever, toolErr := buildComplianceTool()
	if toolErr != nil {
		log.Printf("Warning: compliance_lookup tool disabled: %v", toolErr)
	} else {
		agentTools = append(agentTools, complianceTool)
		complianceRetriever = retriever
	}

	agentConfig := &adk.ChatModelAgentConfig{
		Name:        "CandyProAssistant",
		Description: "CandyPro OEM application assistant specialized in international B2B candy trade.",
		Instruction: buildAssistantInstruction(len(agentTools) > 0),
		Model:       chatModel,
	}

	if len(agentTools) > 0 {
		agentConfig.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: agentTools,
			},
		}
	}

	agent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		return nil, err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           agent,
		EnableStreaming: false,
	})

	return &AIService{
		cfg:                 cfg,
		chatModel:           chatModel,
		chatModelJSON:       chatModelJSON,
		runner:              runner,
		complianceRetriever: complianceRetriever,
	}, nil
}

func buildAssistantInstruction(hasComplianceTool bool) string {
	instruction := prompts.AssistantInstructionBase

	if hasComplianceTool {
		instruction += prompts.AssistantInstructionComplianceAddition
	}
	return instruction
}

func buildComplianceTool() (tool.BaseTool, *rag.ComplianceRetriever, error) {
	corpusDir, err := resolveComplianceCorpusDir()
	if err != nil {
		return nil, nil, err
	}

	retriever, err := rag.NewComplianceRetriever(corpusDir)
	if err != nil {
		return nil, nil, err
	}

	complianceTool, err := rag.NewComplianceTool(retriever)
	if err != nil {
		return nil, nil, err
	}

	return complianceTool, retriever, nil
}

func resolveComplianceCorpusDir() (string, error) {
	candidates := []string{
		filepath.Join("internal", "pkg", "eino", "corpus"),
		filepath.Join(".", "internal", "pkg", "eino", "corpus"),
		filepath.Join("backend", "internal", "pkg", "eino", "corpus"),
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("cannot find compliance corpus directory in known locations")
}

// IsEnabled returns true if AI service is available
func (s *AIService) IsEnabled() bool {
	return s.cfg.IsEnabled() && s.runner != nil
}

// Generate generates a response using the AI model
func (s *AIService) Generate(ctx context.Context, prompt string) (string, error) {
	if s.runner == nil {
		return "", ErrAIServiceDisabled
	}

	iter := s.runner.Query(ctx, prompt)
	var finalResponse string

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if event.Err != nil {
			log.Printf("ADK Agent error: %v, RunPath: %v\n", event.Err, event.RunPath)
			return "", event.Err
		}

		if event.Output != nil && event.Output.MessageOutput != nil {
			if msg := event.Output.MessageOutput.Message; msg != nil {
				if len(msg.Content) > 0 {
					finalResponse += msg.Content
				}
			} else if stream := event.Output.MessageOutput.MessageStream; stream != nil {
				for {
					chunk, err := stream.Recv()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						log.Printf("ADK stream error: %v, RunPath: %v\n", err, event.RunPath)
						return "", err
					}
					if len(chunk.Content) > 0 {
						finalResponse += chunk.Content
					}
				}
			}
		}
	}

	return finalResponse, nil
}

// GenerateJSON calls the AI model with response_format=json_object, guaranteeing valid JSON output.
// The prompt MUST contain the word "json" and include an example JSON schema.
func (s *AIService) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	if s.chatModelJSON == nil {
		return "", ErrAIServiceDisabled
	}

	msgs := []*schema.Message{
		{Role: schema.User, Content: prompt},
	}

	resp, err := s.chatModelJSON.Generate(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("json generation failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("empty response from json model")
	}
	return strings.TrimSpace(resp.Content), nil
}

// AnalyzeInquiry uses AI to analyze customer inquiries with country context
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

// CountryContext holds market-specific information
type CountryContext struct {
	Regulations    string
	MarketPrefs    string
	ShippingNotes  string
	Certifications string
}

// getCountryContext returns market-specific context for AI analysis
func getCountryContext(country string) CountryContext {
	normalized := strings.TrimSpace(strings.ToLower(country))

	// Handle aliases - common country name variations
	aliases := map[string]string{
		// Americas
		"us":                       "usa",
		"united states":            "usa",
		"united states of america": "usa",
		"america":                  "usa",
		"can":                      "canada",
		"mx":                       "mexico",
		"br":                       "brazil",
		"ar":                       "argentina",
		// Middle East
		"ksa":                  "saudi arabia",
		"saudi":                "saudi arabia",
		"united arab emirates": "uae",
		"emirates":             "uae",
		"kw":                   "kuwait",
		"qa":                   "qatar",
		"om":                   "oman",
		"bh":                   "bahrain",
		"eg":                   "egypt",
		"tr":                   "turkey",
		"türkiye":              "turkey",
		// Europe
		"prc":            "china",
		"de":             "germany",
		"deutschland":    "germany",
		"fr":             "france",
		"gb":             "uk",
		"united kingdom": "uk",
		"great britain":  "uk",
		"britain":        "uk",
		"nl":             "netherlands",
		"holland":        "netherlands",
		"be":             "belgium",
		"it":             "italy",
		"es":             "spain",
		"pl":             "poland",
		"ru":             "russia",
		// Asia
		"jp":     "japan",
		"nippon": "japan",
		"cn":     "china",
		"kr":     "south korea",
		"korea":  "south korea",
		"tw":     "taiwan",
		"hk":     "hong kong",
		// Southeast Asia
		"id": "indonesia",
		"my": "malaysia",
		"th": "thailand",
		"vn": "vietnam",
		"ph": "philippines",
		"sg": "singapore",
		// South Asia
		"in": "india",
		"pk": "pakistan",
		"bd": "bangladesh",
		// Africa
		"za": "south africa",
		"ng": "nigeria",
		"ke": "kenya",
		"ma": "morocco",
		// Oceania
		"au": "australia",
		"nz": "new zealand",
	}

	if alias, exists := aliases[normalized]; exists {
		normalized = alias
	}

	contexts := map[string]CountryContext{
		// === Americas ===
		"usa": {
			Regulations:    "FDA registration required, FSMA compliance, nutrition labeling (English)",
			MarketPrefs:    "Lower sweetness, natural ingredients preferred, portion-controlled packaging",
			ShippingNotes:  "5-7 days air freight, 20-30 days sea freight, import duty ~6%",
			Certifications: "FDA, HACCP, GMP, OU Kosher (optional)",
		},
		"canada": {
			Regulations:    "CFIA registration, CFIA compliance, bilingual labeling (English/French)",
			MarketPrefs:    "Similar to US preferences, natural/organic trend, maple flavors popular",
			ShippingNotes:  "3-5 days truck from US, NAFTA/CUSMA benefits, import duty ~0-3%",
			Certifications: "CFIA, HACCP, SQF, Organic (optional)",
		},
		"mexico": {
			Regulations:    "COFEPRIS registration, Spanish labeling, NOM standards compliance",
			MarketPrefs:    "Spicy/chili flavors, tamarind combinations, value-oriented packaging",
			ShippingNotes:  "3-5 days truck, USMCA benefits, import duty ~0-10%",
			Certifications: "COFEPRIS, HACCP, Halal (optional)",
		},
		"brazil": {
			Regulations:    "ANVISA registration, Portuguese labeling, Mercosur standards",
			MarketPrefs:    "Tropical fruit flavors, brigadeiro-style, family-size packaging",
			ShippingNotes:  "7-10 days air freight, 25-35 days sea freight, import duty ~14%",
			Certifications: "ANVISA, HACCP, Halal (optional)",
		},
		"argentina": {
			Regulations:    "ANMAT registration, Spanish labeling, Mercosur standards",
			MarketPrefs:    "Dulce de leche flavors, chocolate coatings, premium positioning",
			ShippingNotes:  "10-14 days air freight, import restrictions possible, duty ~12%",
			Certifications: "ANMAT, HACCP",
		},

		// === Middle East ===
		"saudi arabia": {
			Regulations:    "SFDA registration, mandatory Halal certification, Arabic labeling required",
			MarketPrefs:    "Higher sweetness accepted, date/flavor combinations popular, family-size packaging",
			ShippingNotes:  "7-10 days air freight, temperature-sensitive during summer, import duty ~5%",
			Certifications: "Halal (mandatory), SFDA, HACCP",
		},
		"uae": {
			Regulations:    "ESMA registration, Halal required, bilingual labeling (Arabic/English)",
			MarketPrefs:    "Premium positioning, luxury packaging, gift sets popular",
			ShippingNotes:  "5-7 days air freight, free zones available, import duty ~5%",
			Certifications: "Halal (mandatory), ESMA, HACCP",
		},
		"kuwait": {
			Regulations:    "KFDA registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Premium products, date combinations, gift packaging",
			ShippingNotes:  "7-10 days air freight, summer heat considerations, duty ~5%",
			Certifications: "Halal (mandatory), KFDA, HACCP",
		},
		"qatar": {
			Regulations:    "Qatar Ministry registration, mandatory Halal, Arabic/English labeling",
			MarketPrefs:    "Ultra-premium segment, luxury gift sets, high-quality packaging",
			ShippingNotes:  "5-7 days air freight, free zone options, duty ~5%",
			Certifications: "Halal (mandatory), HACCP",
		},
		"oman": {
			Regulations:    "Oman FDA registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Mid-range pricing, family packaging, traditional flavors",
			ShippingNotes:  "7-10 days air freight, GCC standards apply, duty ~5%",
			Certifications: "Halal (mandatory), HACCP",
		},
		"bahrain": {
			Regulations:    "NHRA registration, mandatory Halal, Arabic/English labeling",
			MarketPrefs:    "Premium segment, diverse expat market, gift packaging",
			ShippingNotes:  "5-7 days air freight via Dubai, duty ~5%",
			Certifications: "Halal (mandatory), HACCP",
		},
		"egypt": {
			Regulations:    "EOS registration, mandatory Halal, Arabic labeling",
			MarketPrefs:    "Value-oriented, local flavors, family-size packaging",
			ShippingNotes:  "7-10 days air freight, currency considerations, duty ~5-30%",
			Certifications: "Halal (mandatory), EOS, HACCP",
		},
		"turkey": {
			Regulations:    "Turkish Ministry registration, Halal preferred, Turkish labeling",
			MarketPrefs:    "Moderate sweetness, lokum-style textures, gift packaging",
			ShippingNotes:  "5-7 days truck/air, customs union with EU, duty ~0-10%",
			Certifications: "Halal (preferred), HACCP, TSE",
		},

		// === Europe ===
		"germany": {
			Regulations:    "EU food safety standards, EFSA compliance, German labeling required",
			MarketPrefs:    "Low sugar trend, organic preferred, sustainable packaging important",
			ShippingNotes:  "3-5 days truck from port, strict customs documentation, import duty ~7%",
			Certifications: "IFS, BRC, Organic (optional), Halal (optional)",
		},
		"france": {
			Regulations:    "EU food safety standards, DGCCRF compliance, French labeling required",
			MarketPrefs:    "Premium positioning, artisanal image, high quality chocolate coatings",
			ShippingNotes:  "3-5 days truck from EU port, standard EU customs",
			Certifications: "IFS, BRC, Organic (optional)",
		},
		"uk": {
			Regulations:    "FSA regulations, post-Brexit UK labeling requirements",
			MarketPrefs:    "Premium packaging, traditional flavors, vegetarian/vegan options popular",
			ShippingNotes:  "5-7 days air freight, customs documentation required",
			Certifications: "BRCGS, HACCP",
		},
		"netherlands": {
			Regulations:    "EU food safety standards, NVWA compliance, Dutch/English labeling",
			MarketPrefs:    "Natural ingredients, sustainable packaging, licorice popular",
			ShippingNotes:  "Major port access (Rotterdam), 1-3 days distribution, duty ~7%",
			Certifications: "IFS, BRC, HACCP",
		},
		"belgium": {
			Regulations:    "EU food safety standards, FAVV compliance, Dutch/French labeling",
			MarketPrefs:    "Premium chocolate coatings, praline-style, high quality focus",
			ShippingNotes:  "3-5 days from Rotterdam/Antwerp, EU distribution hub",
			Certifications: "IFS, BRC, HACCP",
		},
		"italy": {
			Regulations:    "EU food safety standards, Ministry of Health compliance, Italian labeling",
			MarketPrefs:    "Premium positioning, hazelnut combinations, artisanal appeal",
			ShippingNotes:  "3-5 days truck from EU port, standard EU customs",
			Certifications: "IFS, BRC, HACCP, Organic (optional)",
		},
		"spain": {
			Regulations:    "EU food safety standards, AESAN compliance, Spanish labeling",
			MarketPrefs:    "Fruit flavors, turron-style, value to mid-range segments",
			ShippingNotes:  "3-5 days truck from EU port, standard EU customs",
			Certifications: "IFS, BRC, HACCP, Halal (optional)",
		},
		"poland": {
			Regulations:    "EU food safety standards, GIS compliance, Polish labeling",
			MarketPrefs:    "Value-oriented, traditional flavors, growing premium segment",
			ShippingNotes:  "3-5 days truck from EU port, lower distribution costs",
			Certifications: "IFS, BRC, HACCP",
		},
		"russia": {
			Regulations:    "EAEU standards, Rospotrebnadzor registration, Russian labeling",
			MarketPrefs:    "Moderate sweetness, chocolate coatings, gift packaging",
			ShippingNotes:  "10-14 days rail/truck, sanctions considerations, duty ~10%",
			Certifications: "EAC, HACCP, Halal (optional)",
		},

		// === Asia ===
		"japan": {
			Regulations:    "MHLW standards, Japanese labeling, strict quality requirements",
			MarketPrefs:    "Refined sweetness, seasonal flavors, gift packaging important",
			ShippingNotes:  "3-5 days air freight, strict import inspection, import duty ~3%",
			Certifications: "JAS, Organic JAS (optional), Halal (optional)",
		},
		"china": {
			Regulations:    "GACC registration, Chinese labeling, GB standards compliance",
			MarketPrefs:    "Regional flavor preferences, gift packaging for holidays, e-commerce ready",
			ShippingNotes:  "5-7 days air freight, CIQ inspection required, import duty ~10%",
			Certifications: "GACC registration, HACCP, ISO22000",
		},
		"south korea": {
			Regulations:    "MFDS registration, Korean labeling, strict quality standards",
			MarketPrefs:    "Low sugar trend, innovative flavors, premium packaging",
			ShippingNotes:  "3-5 days air freight, strict customs, duty ~8%",
			Certifications: "MFDS, HACCP, Halal (optional)",
		},
		"taiwan": {
			Regulations:    "TFDA registration, Chinese labeling, food safety standards",
			MarketPrefs:    "Similar to Japan, bubble tea flavors, gift packaging",
			ShippingNotes:  "3-5 days air freight, standard customs, duty ~5%",
			Certifications: "TFDA, HACCP, Halal (optional)",
		},
		"hong kong": {
			Regulations:    "CFS registration, English/Chinese labeling, food safety standards",
			MarketPrefs:    "Premium positioning, gift packaging, diverse flavors",
			ShippingNotes:  "3-5 days air freight, free port benefits, duty ~0%",
			Certifications: "HACCP, ISO22000, Halal (optional)",
		},

		// === Southeast Asia ===
		"indonesia": {
			Regulations:    "BPOM registration, MUI Halal mandatory, Indonesian labeling",
			MarketPrefs:    "Tropical fruit flavors, moderate sweetness, value packaging",
			ShippingNotes:  "5-7 days air freight, 15-20 days sea, duty ~5-15%",
			Certifications: "Halal MUI (mandatory), BPOM, HACCP",
		},
		"malaysia": {
			Regulations:    "NPRA registration, JAKIM Halal mandatory, Malay/English labeling",
			MarketPrefs:    "Durian flavors, pandan combinations, diverse ethnic market",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-5%",
			Certifications: "Halal JAKIM (mandatory), NPRA, HACCP",
		},
		"thailand": {
			Regulations:    "FDA Thailand registration, Thai labeling, food safety standards",
			MarketPrefs:    "Tropical flavors, mango/mango sticky rice, gift packaging",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-5%",
			Certifications: "Thai FDA, HACCP, Halal (optional)",
		},
		"vietnam": {
			Regulations:    "VFA registration, Vietnamese labeling, food safety standards",
			MarketPrefs:    "Coffee flavors, tropical fruits, growing premium segment",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-10%",
			Certifications: "VFA, HACCP, Halal (optional)",
		},
		"philippines": {
			Regulations:    "FDA Philippines registration, English/Filipino labeling",
			MarketPrefs:    "Ube flavors, tropical fruits, value to mid-range",
			ShippingNotes:  "5-7 days air freight, ASEAN benefits, duty ~0-10%",
			Certifications: "FDA Philippines, HACCP, Halal (optional)",
		},
		"singapore": {
			Regulations:    "SFA registration, English labeling, strict food safety",
			MarketPrefs:    "Premium positioning, innovative flavors, health-conscious",
			ShippingNotes:  "3-5 days air freight, free port, duty ~0%",
			Certifications: "SFA, HACCP, Halal MUIS (optional)",
		},

		// === South Asia ===
		"india": {
			Regulations:    "FSSAI registration, Hindi/English labeling, strict standards",
			MarketPrefs:    "Mango flavors, spice combinations, value packaging, vegetarian",
			ShippingNotes:  "5-7 days air freight, 20-30 days sea, duty ~30%",
			Certifications: "FSSAI, HACCP, Halal (optional), Vegetarian mark",
		},
		"pakistan": {
			Regulations:    "PSQCA registration, mandatory Halal, Urdu labeling",
			MarketPrefs:    "Moderate sweetness, value-oriented, family packaging",
			ShippingNotes:  "5-7 days air freight, import restrictions possible, duty ~10-20%",
			Certifications: "Halal (mandatory), PSQCA, HACCP",
		},
		"bangladesh": {
			Regulations:    "BFSA registration, mandatory Halal, Bengali labeling",
			MarketPrefs:    "Value-oriented, traditional flavors, family packaging",
			ShippingNotes:  "7-10 days air freight, SAFTA benefits, duty ~10-25%",
			Certifications: "Halal (mandatory), BFSA, HACCP",
		},

		// === Africa ===
		"south africa": {
			Regulations:    "DTI registration, English labeling, SABS standards",
			MarketPrefs:    "Value to mid-range, diverse market, gift packaging",
			ShippingNotes:  "7-10 days air freight, AGOA benefits, duty ~5-20%",
			Certifications: "SABS, HACCP, Halal (optional)",
		},
		"nigeria": {
			Regulations:    "NAFDAC registration, English labeling, food safety standards",
			MarketPrefs:    "Value-oriented, growing middle class, family packaging",
			ShippingNotes:  "7-10 days air freight, port congestion possible, duty ~5-20%",
			Certifications: "NAFDAC, HACCP, Halal (optional)",
		},
		"kenya": {
			Regulations:    "KEBS registration, English labeling, EAC standards",
			MarketPrefs:    "Value to mid-range, tropical flavors, growing market",
			ShippingNotes:  "7-10 days air freight, EAC benefits, duty ~0-25%",
			Certifications: "KEBS, HACCP, Halal (optional)",
		},
		"morocco": {
			Regulations:    "ONSSA registration, Arabic/French labeling, food safety standards",
			MarketPrefs:    "Almond/honey combinations, value to mid-range",
			ShippingNotes:  "5-7 days air freight, EU trade agreement, duty ~0-10%",
			Certifications: "ONSSA, HACCP, Halal (optional)",
		},

		// === Oceania ===
		"australia": {
			Regulations:    "FSANZ standards, strict biosecurity, mandatory Australian labeling",
			MarketPrefs:    "Natural colors/flavors, bite-sized portions, high quality ingredients",
			ShippingNotes:  "7-10 days air freight, strict border inspection",
			Certifications: "HACCP, SQF",
		},
		"new zealand": {
			Regulations:    "FSANZ standards, MPI requirements, English labeling",
			MarketPrefs:    "Natural/organic trend, premium positioning, sustainable packaging",
			ShippingNotes:  "7-10 days air freight, strict biosecurity, duty ~0-5%",
			Certifications: "HACCP, MPI approved, Organic (optional)",
		},
	}

	if ctx, ok := contexts[normalized]; ok {
		return ctx
	}

	// Default for unknown countries
	return CountryContext{
		Regulations:    "Check local food safety authority requirements",
		MarketPrefs:    "Research local taste preferences and packaging norms",
		ShippingNotes:  "Verify shipping routes and import regulations",
		Certifications: "HACCP, ISO22000 recommended, check specific requirements",
	}
}

// GetMarketInsights returns market-specific insights for a country
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

// SuggestProducts recommends products based on country and requirements
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

// GetShippingOptions returns shipping specific information for a country
func (s *AIService) GetShippingOptions(ctx context.Context, country string) (string, error) {
	countryContext := getCountryContext(country)
	prompt := fmt.Sprintf(`Provide B2B candy shipping options and logistics details for %s:
1. Standard shipping methods and estimated times
2. Import duties and taxes (%s)
3. Customs clearance process
4. Temperature control requirements during transit
5. Required shipping documentation`, country, countryContext.ShippingNotes)
	return s.Generate(ctx, prompt)
}

// GetRegulatoryInfo returns regulatory compliance information for a country
func (s *AIService) GetRegulatoryInfo(ctx context.Context, country string) (string, error) {
	countryContext := getCountryContext(country)

	var prompt string
	if s.complianceRetriever == nil {
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
