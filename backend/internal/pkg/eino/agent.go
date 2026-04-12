package eino

import (
	"context"

	"candypro/api/internal/pkg/eino/prompts"
	einotool "candypro/api/internal/pkg/eino/tool"

	commonModel "github.com/cloudwego/eino-examples/adk/common/model"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// NewTradeAgent creates a smart agent that acts as a Trade Coordinator,
// guiding the user through Quotation -> PI -> CI -> etc.
// Utilizing Eino's PlanExecute mechanism.
func NewTradeAgent(ctx context.Context) (adk.Agent, error) {
	// Initialize tools
	piTool, err := einotool.NewGeneratePITool(ctx)
	if err != nil {
		return nil, err
	}

	ciTool, err := einotool.NewGenerateCITool(ctx)
	if err != nil {
		return nil, err
	}

	compTool, err := einotool.NewComplianceCheckTool(ctx)
	if err != nil {
		return nil, err
	}

	scTool, err := einotool.NewGenerateSalesContractTool(ctx)
	if err != nil {
		return nil, err
	}

	plTool, err := einotool.NewGeneratePackingListTool(ctx)
	if err != nil {
		return nil, err
	}

	cooTool, err := einotool.NewGenerateCertificateOfOriginTool(ctx)
	if err != nil {
		return nil, err
	}

	hcTool, err := einotool.NewGenerateHealthCertificateTool(ctx)
	if err != nil {
		return nil, err
	}

	ingTool, err := einotool.NewGenerateIngredientsDeclarationTool(ctx)
	if err != nil {
		return nil, err
	}

	sliTool, err := einotool.NewGenerateSLITool(ctx)
	if err != nil {
		return nil, err
	}

	lcTool, err := einotool.NewValidateLCTool(ctx)
	if err != nil {
		return nil, err
	}

	insTool, err := einotool.NewGenerateInsuranceCertTool(ctx)
	if err != nil {
		return nil, err
	}

	trackTool, err := einotool.NewTrackShipmentTool(ctx)
	if err != nil {
		return nil, err
	}

	llmModel := commonModel.NewChatModel() // Needs API key in env (.env loaded via config)

	// Since we're using a standard ReAct style ChatModelAgent we pass the tools.
	// You can also use PlanExecute if it involves deep long-running planning.
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TradeAssistant",
		Description: "An AI assistant capable of guiding users through candy foreign trade procedures.",
		Instruction: prompts.TradeAgentInstruction,
		Model:       llmModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{
					piTool, ciTool, compTool, scTool, plTool, cooTool,
					hcTool, ingTool, sliTool, lcTool, insTool, trackTool,
				},
			},
			ReturnDirectly: map[string]bool{
				// Can mark tools to return directly if we stream
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return agent, nil
}
