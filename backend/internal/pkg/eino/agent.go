package eino

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"candypro/api/internal/pkg/eino/graph"
	"candypro/api/internal/pkg/eino/prompts/agent"
	einotool "candypro/api/internal/pkg/eino/tool"
	"candypro/api/internal/pkg/eino/tool/rag"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// NewTradeAgent creates a TradeAssistant ChatModelAgent following the official Eino
// "Graph as Agent Tool" architecture.
//
// Architecture:
//
//	Agent (LLM decision center)
//	├── generate_trade_documents (Graph Tool: deterministic doc pipeline)
//	├── translate_content (AI translation to multiple locales)
//	├── check_compliance (rule-based country check)
//	├── compliance_lookup (RAG corpus search)
//	├── validate_lc_documents (L/C document verification)
//	├── track_shipment (shipment tracking)
//	└── submit_quotation_for_human_review (HITL approval)
//
// chatModel is injected by the caller so both agent and AIService share one model config.
// persister may be nil for chat-only deployments.
// translateFn may be nil to skip the translate_content tool.
func NewTradeAgent(ctx context.Context, chatModel model.ToolCallingChatModel, persister einotool.DocumentPersister, translateFn einotool.TranslateFunc) (adk.Agent, error) {
	// ── Graph Tool: deterministic document generation pipeline ──
	// Per Eino official guide: "Encapsulating Graph as Agent's Tool achieves 1+1 > 2"
	docGraph, err := graph.NewDocumentPipelineGraph(ctx, persister)
	if err != nil {
		return nil, fmt.Errorf("document pipeline graph: %w", err)
	}
	docGraphTool, err := graph.NewDocPipelineTool(ctx, docGraph)
	if err != nil {
		return nil, fmt.Errorf("document pipeline tool: %w", err)
	}

	// ── Individual tools for non-document operations ──
	compTool, err := einotool.NewComplianceCheckTool(ctx)
	if err != nil {
		return nil, err
	}

	lcTool, err := einotool.NewValidateLCTool(ctx)
	if err != nil {
		return nil, err
	}

	trackTool, err := einotool.NewTrackShipmentTool(ctx)
	if err != nil {
		return nil, err
	}

	quoteReviewTool, err := einotool.NewQuotationHumanReviewTool(ctx)
	if err != nil {
		return nil, err
	}

	tools := []tool.BaseTool{
		docGraphTool, // Replaces 9 individual doc generation tools
		compTool,
		lcTool,
		trackTool,
		quoteReviewTool,
	}

	// translate_content tool — AI-powered batch translation to multiple locales
	if translateFn != nil {
		translateTool, translateErr := einotool.NewTranslateContentTool(ctx, translateFn)
		if translateErr != nil {
			log.Printf("Warning: translate_content tool not available: %v", translateErr)
		} else {
			tools = append(tools, translateTool)
		}
	}

	// RAG compliance lookup tool is temporarily disabled (Phase 5).
	// Keep buildRagComplianceTool() + rag package + corpus/ intact for Phase 5.
	// Uncomment the block below to re-enable:
	/*
		if ragTool, ragErr := buildRagComplianceTool(); ragErr == nil {
			tools = append(tools, ragTool)
		} else {
			log.Printf("Warning: RAG compliance lookup tool not available: %v", ragErr)
		}
	*/

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TradeAssistant",
		Description: "AI trade coordinator for CandyPro OEM. Generates trade documents, translates content to multiple languages, checks compliance, validates L/C, tracks shipments, and queues quotations for human review.",
		Instruction: agent.TradeAgentInstruction,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
			ReturnDirectly: map[string]bool{
				"submit_quotation_for_human_review": true, // Exit after queuing for review
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Deprecated: use rag.NewFromCorpus() instead.
func buildRagComplianceTool() (tool.BaseTool, error) {
	corpusDir, err := resolveComplianceCorpusDir()
	if err != nil {
		return nil, err
	}

	retriever, err := rag.NewComplianceRetriever(corpusDir)
	if err != nil {
		return nil, err
	}

	return rag.NewComplianceTool(retriever)
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
