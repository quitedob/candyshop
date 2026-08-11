// Package eino — Plan-Execute-Replan Order Processing Pipeline
//
// Implements the Eino prebuilt planexecute pattern for structured B2B order processing.
// The pipeline follows a Plan-Execute-Replan cycle to handle complex order workflows
// with exception handling and human-in-the-loop approval gates.
//
// Pipeline flow:
//
//	Planner:  Analyze inquiry/order → Generate execution plan
//	Executor: Execute first step (validate compliance, validate L/C documents, track shipment)
//	Replanner: Evaluate result → Complete (with response) or Revise plan
//	↳ Loop until completion or max iterations
package eino

import (
	"context"
	"fmt"

	einotool "candypro/api/internal/pkg/eino/tool"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/planexecute"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/components/tool"
)

// NewOrderProcessingAgent creates a Plan-Execute-Replan agent for structured order processing.
// It follows a plan → execute → evaluate/replan cycle to handle complex B2B order workflows.
//
// The pipeline handles:
// 1. Planning: Analyzes the order request and creates a step-by-step execution plan
// 2. Execution: Executes each step (compliance check, L/C validation, shipment tracking)
// 3. Replanning: Evaluates results and either completes or revises the plan
//
// chatModel is used by all three components (planner, executor, replanner).
func NewOrderProcessingAgent(ctx context.Context, chatModel model.ToolCallingChatModel) (adk.Agent, error) {
	// ── Common tools for the executor ──
	compTool, err := einotool.NewComplianceCheckTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("compliance check tool: %w", err)
	}

	lcTool, err := einotool.NewValidateLCTool(ctx)
	if err != nil {
		return nil, fmt.Errorf("L/C validation tool: %w", err)
	}

	trackTool, err := einotool.NewTrackShipmentTool(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("shipment tracking tool: %w", err)
	}

	executorTools := adk.ToolsConfig{
		ToolsNodeConfig: compose.ToolsNodeConfig{
			Tools: []tool.BaseTool{
				compTool,
				lcTool,
				trackTool,
			},
		},
	}

	// ── Planner ──
	// Generates a step-by-step plan from the order request
	planner, err := planexecute.NewPlanner(ctx, &planexecute.PlannerConfig{
		ToolCallingChatModel: chatModel,
	})
	if err != nil {
		return nil, fmt.Errorf("new planner: %w", err)
	}

	// ── Executor ──
	// Executes the current step using available tools
	executor, err := planexecute.NewExecutor(ctx, &planexecute.ExecutorConfig{
		Model:         chatModel,
		ToolsConfig:   executorTools,
		MaxIterations: 20,
	})
	if err != nil {
		return nil, fmt.Errorf("new executor: %w", err)
	}

	// ── Replanner ──
	// Evaluates execution results and either responds or revises the plan
	replanner, err := planexecute.NewReplanner(ctx, &planexecute.ReplannerConfig{
		ChatModel: chatModel,
	})
	if err != nil {
		return nil, fmt.Errorf("new replanner: %w", err)
	}

	// ── Assemble P-E-R agent ──
	agent, err := planexecute.New(ctx, &planexecute.Config{
		Planner:       planner,
		Executor:      executor,
		Replanner:     replanner,
		MaxIterations: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("planexecute.New: %w", err)
	}

	return agent, nil
}
