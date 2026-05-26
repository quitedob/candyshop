// Package eino — DeepAgent B2B Trade Coordinator
//
// Implements the Eino prebuilt DeepAgent pattern for multi-agent B2B trade coordination.
// The main coordinator agent delegates to specialized sub-agents (ProductExpert, PricingExpert,
// LogisticsExpert) and individual tools (document generation, compliance, shipment tracking).
//
// Architecture:
//
//	DeepAgent (Coordinator)
//	├── ProductExpert sub-agent — product matching, catalog queries
//	├── PricingExpert sub-agent — pricing calculations, tiered quotes
//	├── LogisticsExpert sub-agent — shipping estimation, routing
//	├── generate_trade_documents (Graph Tool)
//	├── check_compliance (rule-based)
//	├── validate_lc_documents
//	├── track_shipment
//	└── submit_quotation_for_human_review (HITL)
package eino

import (
	"context"
	"fmt"

	"candypro/api/internal/pkg/eino/graph"
	einotool "candypro/api/internal/pkg/eino/tool"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// NewB2BCoordinatorAgent creates a DeepAgent that coordinates B2B trade operations.
// It uses sub-agents for domain-specific tasks (product, pricing, logistics) and
// individual tools for deterministic operations.
//
// chatModel is shared across all sub-agents and the coordinator.
// persister may be nil for chat-only deployments.
func NewB2BCoordinatorAgent(ctx context.Context, chatModel model.ToolCallingChatModel, persister einotool.DocumentPersister, catalog einotool.ProductCatalogSearcher, quoter einotool.PricingQuoter) (adk.Agent, error) {
	// ── Common tools (shared across coordinator and sub-agents) ──

	// Document generation graph tool
	docGraph, err := graph.NewDocumentPipelineGraph(ctx, persister)
	if err != nil {
		return nil, fmt.Errorf("document pipeline graph: %w", err)
	}
	docGraphTool, err := graph.NewDocPipelineTool(ctx, docGraph)
	if err != nil {
		return nil, fmt.Errorf("document pipeline tool: %w", err)
	}

	// Individual tools
	compTool, err := einotool.NewComplianceCheckTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("compliance check tool: %w", err)
	}

	lcTool, err := einotool.NewValidateLCTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("L/C validation tool: %w", err)
	}

	trackTool, err := einotool.NewTrackShipmentTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("shipment tracking tool: %w", err)
	}

	quoteReviewTool, err := einotool.NewQuotationHumanReviewTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("quotation review tool: %w", err)
	}

	coordinatorTools := []tool.BaseTool{
		docGraphTool,
		compTool,
		lcTool,
		trackTool,
		quoteReviewTool,
	}

	// ── Sub-agents ──
	var productTools []tool.BaseTool
	if catalog != nil {
		if searchTool, err := einotool.NewSearchProductsTool(ctx, catalog); err == nil {
			productTools = append(productTools, searchTool)
		}
		if detailTool, err := einotool.NewGetProductDetailTool(ctx, catalog); err == nil {
			productTools = append(productTools, detailTool)
		}
	}
	productExpert, err := newProductExpertAgent(ctx, chatModel, productTools)
	if err != nil {
		return nil, fmt.Errorf("product expert agent: %w", err)
	}

	pricingExpert, err := newPricingExpertAgent(ctx, chatModel, quoter)
	if err != nil {
		return nil, fmt.Errorf("pricing expert agent: %w", err)
	}

	logisticsExpert, err := newLogisticsExpertAgent(ctx, chatModel, trackTool)
	if err != nil {
		return nil, fmt.Errorf("logistics expert agent: %w", err)
	}

	// ── DeepAgent coordinator ──
	// Uses deep.New to orchestrate sub-agents and tools with task planning
	deepAgent, err := deep.New(ctx, &deep.Config{
		Name:        "B2BTradeCoordinator",
		Description: "B2B trade coordinator for CandyPro OEM. Coordinates product matching, pricing, logistics, document generation, compliance, and L/C validation across specialized sub-agents.",
		ChatModel:   chatModel,
		SubAgents:   []adk.Agent{productExpert, pricingExpert, logisticsExpert},
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: coordinatorTools,
			},
			ReturnDirectly: map[string]bool{
				"submit_quotation_for_human_review": true,
			},
		},
		MaxIteration: 50,
	})
	if err != nil {
		return nil, fmt.Errorf("deep.New: %w", err)
	}

	return deepAgent, nil
}

// newProductExpertAgent creates a sub-agent focused on product matching and catalog queries.
func newProductExpertAgent(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	cfg := &adk.ChatModelAgentConfig{
		Name:        "ProductExpert",
		Description: "Product matching and catalog specialist. Searches the product catalog, matches customer requirements to available products, and recommends suitable options based on specifications, certifications, and market requirements.",
		Instruction: `You are a Product Expert at CandyPro OEM, a professional B2B candy manufacturer.

Your role:
1. Match customer requirements to products in the catalog
2. Recommend suitable product options based on specifications, certifications, and target market
3. Provide product details: MOQ, lead time, certifications (Halal, HACCP, ISO), packaging options
4. Identify customization opportunities (OEM, flavors, shapes, packaging)

Use search_products and get_product_detail tools when you need live catalog data.

Be concise and specific. Always reference product names and specifications.`,
		Model: chatModel,
	}
	if len(tools) > 0 {
		cfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		}
	}
	return adk.NewChatModelAgent(ctx, cfg)
}

// newPricingExpertAgent creates a sub-agent focused on pricing calculations.
func newPricingExpertAgent(ctx context.Context, chatModel model.ToolCallingChatModel, quoter einotool.PricingQuoter) (adk.Agent, error) {
	var pricingTools []tool.BaseTool
	if quoter != nil {
		if tiersTool, err := einotool.NewGetPriceTiersTool(ctx, quoter); err == nil {
			pricingTools = append(pricingTools, tiersTool)
		}
		if quoteTool, err := einotool.NewComputeQuoteDraftTool(ctx, quoter); err == nil {
			pricingTools = append(pricingTools, quoteTool)
		}
	}
	cfg := &adk.ChatModelAgentConfig{
		Name:        "PricingExpert",
		Description: "Pricing and quotation specialist. Calculates pricing based on quantities, tiers, market costs, and currency conversion. Provides quotation drafts for human review.",
		Instruction: `You are a Pricing Expert at CandyPro OEM, a professional B2B candy manufacturer.

Your role:
1. Calculate pricing based on quantities and pricing tiers
2. Factor in target market costs and currency conversion
3. Prepare quotation drafts with clear pricing assumptions
4. Flag orders that require human approval (>$10,000)

Use get_price_tiers and compute_quote_draft tools when you need live pricing data.

When the coordinator asks you to calculate pricing:
- Consider quantity-based tier pricing
- Account for target country costs
- Include MOQ adjustments
- Specify currency (default USD)
- Flag amounts above $10,000 for human review
- Provide clear breakdown: unit price, subtotal, estimated shipping, total

Be precise and structured. Always state assumptions.`,
		Model: chatModel,
	}
	if len(pricingTools) > 0 {
		cfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: pricingTools},
		}
	}
	return adk.NewChatModelAgent(ctx, cfg)
}

// newLogisticsExpertAgent creates a sub-agent focused on logistics coordination.
func newLogisticsExpertAgent(ctx context.Context, chatModel model.ToolCallingChatModel, trackTool tool.BaseTool) (adk.Agent, error) {
	cfg := &adk.ChatModelAgentConfig{
		Name:        "LogisticsExpert",
		Description: "Logistics and shipping specialist. Advises on shipping routes, Incoterms, documentation, and tracking. Does not quote guaranteed transit day counts.",
		Instruction: `You are a Logistics Expert at CandyPro OEM, a professional B2B candy manufacturer.

Your role:
1. Advise on shipping routes and methods based on origin (Shanghai, China) and destination — never state specific guaranteed transit day/week counts; say transit varies by method, season, and destination and is confirmed at order time
2. Advise on Incoterms (FOB, CIF, EXW, DDP) and their implications
3. Coordinate required shipping documentation
4. Provide tracking information when available using track_shipment tool
5. Consider temperature control for candy products

Be practical about requirements without promising fixed delivery windows.`,
		Model: chatModel,
	}
	if trackTool != nil {
		cfg.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: []tool.BaseTool{trackTool}},
		}
	}
	return adk.NewChatModelAgent(ctx, cfg)
}
