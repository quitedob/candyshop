package eino

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/eino/prompts/agent"
	"candypro/api/internal/pkg/eino/retry"
	"candypro/api/internal/pkg/eino/tool/rag"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ErrDisabled indicates the AI client is not configured (no API key).
var ErrDisabled = errors.New("ai service is not configured")

// Client is the core AI client managing chat model lifecycle, agent attachment,
// and generation (plain text + JSON mode). It is shared by all scoped AIService wrappers.
type Client struct {
	cfg                 config.AIConfig
	chatModel           model.ToolCallingChatModel
	chatModelJSON       model.ToolCallingChatModel
	agent               adk.Agent
	runner              *adk.Runner
	complianceRetriever *rag.ComplianceRetriever
	checkPointStore     compose.CheckPointStore
}

// NewClient creates a new AI client with a basic chat model (with retry).
func NewClient(cfg config.AIConfig) (*Client, error) {
	if cfg.OpenAIAPIKey == "" {
		return nil, ErrDisabled
	}

	ctx := context.Background()

	rawModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   cfg.OpenAIModel,
		APIKey:  cfg.OpenAIAPIKey,
		BaseURL: cfg.OpenAIBaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat model: %w", err)
	}
	chatModel := retry.New(rawModel, cfg.RetryMaxAttempts, cfg.RetryIntervalSec)

	jsonResponseFormat := openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONObject,
	}
	rawJSONModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:          cfg.OpenAIModel,
		APIKey:         cfg.OpenAIAPIKey,
		BaseURL:        cfg.OpenAIBaseURL,
		ResponseFormat: &jsonResponseFormat,
	})
	var chatModelJSON model.ToolCallingChatModel
	if err != nil {
		log.Printf("Warning: failed to initialize JSON chat model, falling back: %v", err)
		chatModelJSON = chatModel
	} else {
		chatModelJSON = retry.New(rawJSONModel, cfg.RetryMaxAttempts, cfg.RetryIntervalSec)
	}

	// RAG compliance lookup is temporarily disabled (Phase 5).
	// Keep rag package + corpus/ intact; uncomment below to re-enable.
	/*
		complianceTool, retriever, toolErr := buildComplianceTool()
		if toolErr != nil {
			log.Printf("Warning: compliance_lookup tool disabled: %v", toolErr)
		}
	*/

	runner, err := buildClientRunner(ctx, chatModel, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build legacy runner: %w", err)
	}

	return &Client{
		cfg:           cfg,
		chatModel:     chatModel,
		chatModelJSON: chatModelJSON,
		runner:        runner,
		// complianceRetriever: retriever, // RAG — Phase 5
	}, nil
}

// NewClientWithAgent creates a client pre-wired with the full TradeAgent.
func NewClientWithAgent(cfg config.AIConfig, agent adk.Agent) (*Client, error) {
	c, err := NewClient(cfg)
	if err != nil {
		return nil, err
	}
	c.agent = agent
	return c, nil
}

// AttachAgent sets or replaces the TradeAgent. Call after NewClient if the agent
// is constructed later.
func (c *Client) AttachAgent(agent adk.Agent) {
	c.agent = agent
}

// SetCheckPointStore sets the checkpoint store for agent persistence.
func (c *Client) SetCheckPointStore(store compose.CheckPointStore) {
	c.checkPointStore = store
}

// ComplianceRetriever returns the RAG compliance retriever (nil if disabled).
func (c *Client) ComplianceRetriever() *rag.ComplianceRetriever {
	return c.complianceRetriever
}

// IsEnabled returns true if the AI client was successfully initialized.
func (c *Client) IsEnabled() bool {
	return c.cfg.IsEnabled()
}

// Generate runs the full TradeAgent (or legacy runner) and returns the final text response.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	runner := c.runner
	if c.agent != nil {
		runner = adk.NewRunner(ctx, adk.RunnerConfig{
			Agent:           c.agent,
			EnableStreaming: false,
			CheckPointStore: c.checkPointStore,
		})
	}
	if runner == nil {
		return "", ErrDisabled
	}

	iter := runner.Query(ctx, prompt)
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
func (c *Client) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	if c.chatModelJSON == nil {
		return "", ErrDisabled
	}

	msgs := []*schema.Message{
		{Role: schema.User, Content: prompt},
	}

	resp, err := c.chatModelJSON.Generate(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("json generation failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("empty response from json model")
	}
	return strings.TrimSpace(resp.Content), nil
}

// buildClientRunner creates the legacy single-tool agent runner used when no full TradeAgent is attached.
func buildClientRunner(ctx context.Context, chatModel model.ToolCallingChatModel, complianceTool tool.BaseTool, checkpointStore compose.CheckPointStore) (*adk.Runner, error) {
	agentTools := make([]tool.BaseTool, 0, 1)
	if complianceTool != nil {
		agentTools = append(agentTools, complianceTool)
	}

	agentConfig := &adk.ChatModelAgentConfig{
		Name:        "CandyProAssistant",
		Description: "CandyPro OEM application assistant specialized in international B2B candy trade.",
		Instruction: buildClientInstruction(len(agentTools) > 0),
		Model:       chatModel,
	}

	if len(agentTools) > 0 {
		agentConfig.ToolsConfig = adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: agentTools,
			},
		}
	}

	chatAgent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("buildClientRunner: NewChatModelAgent failed: %w", err)
	}

	return adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           chatAgent,
		EnableStreaming: false,
		CheckPointStore: checkpointStore,
	}), nil
}

func buildClientInstruction(hasComplianceTool bool) string {
	instruction := agent.AssistantInstructionBase
	if hasComplianceTool {
		instruction += agent.AssistantInstructionComplianceAddition
	}
	return instruction
}

