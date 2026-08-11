package eino

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/eino/prompts/agent"
	"candypro/api/internal/pkg/eino/retry"
	"candypro/api/internal/pkg/eino/tool/rag"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ErrDisabled indicates the AI client is not configured (no API key).
var ErrDisabled = errors.New("ai service is not configured")

// RAGComplianceEnabled 由 cfg.AI.RAGComplianceEnabled 在 NewClient 时赋值；控制合规语料 RAG 是否启用。
var RAGComplianceEnabled = false

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
	translateRetry := cfg.TranslateRetryMax
	if translateRetry <= 0 {
		translateRetry = 2
	}

	rawModel, err := NewDeepSeekChatModel(ctx, cfg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize chat model: %w", err)
	}
	chatModel := retry.New(rawModel, cfg.RetryMaxAttempts, cfg.RetryIntervalSec)

	jsonResponseFormat := openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONObject,
	}
	rawJSONModel, err := NewDeepSeekChatModel(ctx, cfg, &openai.ChatModelConfig{
		ResponseFormat: &jsonResponseFormat,
	})
	var chatModelJSON model.ToolCallingChatModel
	if err != nil {
		log.Printf("Warning: failed to initialize JSON chat model, falling back: %v", err)
		chatModelJSON = chatModel
	} else {
		chatModelJSON = retry.New(rawJSONModel, translateRetry, cfg.RetryIntervalSec)
	}

	c := &Client{
		cfg:           cfg,
		chatModel:     chatModel,
		chatModelJSON: chatModelJSON,
	}

	// 合规 RAG 由配置驱动（AI_RAG_COMPLIANCE_ENABLED=true 时启用，默认关闭）。
	RAGComplianceEnabled = cfg.RAGComplianceEnabled
	if RAGComplianceEnabled {
		if retriever, ragErr := rag.NewFromCorpus(); ragErr == nil {
			c.complianceRetriever = retriever
		} else {
			log.Printf("Warning: compliance RAG enabled but corpus unavailable: %v", ragErr)
		}
	}

	// legacy runner：不挂载 compliance 工具，仅保留纯 Chat 回退。
	runner, err := buildClientRunner(ctx, chatModel)
	if err != nil {
		return nil, fmt.Errorf("failed to build legacy runner: %w", err)
	}
	c.runner = runner

	return c, nil
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

// thinkTagRegex matches <think>...</think> blocks (including multiline) from reasoning models.
var thinkTagRegex = regexp.MustCompile(`(?s)<think>.*?</think>`)

// stripThinkBlocks removes reasoning-model think blocks and trims surrounding whitespace.
func stripThinkBlocks(s string) string {
	return strings.TrimSpace(thinkTagRegex.ReplaceAllString(s, ""))
}

// Generate runs the full TradeAgent (or legacy runner) and returns the final text response.
// Reasoning-model <think> blocks are stripped from the output.
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
	// lastOutput is the trailing MessageOutput event. A ReturnDirectly tool
	// (e.g. submit_quotation_for_human_review) ends the agent with its own
	// user-facing result carried as a Role==Tool message. That content is the
	// answer, not raw tool output, so Generate re-appends it after the stream
	// ends via appendTerminalToolResult; intermediate tool results are filtered
	// by accumulateMessageOutput instead.
	var lastOutput *adk.MessageVariant

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
			lastOutput = event.Output.MessageOutput
			var err error
			finalResponse, err = accumulateMessageOutput(finalResponse, lastOutput)
			if err != nil {
				log.Printf("ADK stream error: %v, RunPath: %v\n", err, event.RunPath)
				return "", err
			}
		}
	}

	finalResponse = appendTerminalToolResult(finalResponse, lastOutput)

	return stripThinkBlocks(finalResponse), nil
}

// appendTerminalToolResult appends the content of a trailing Role==Tool
// MessageOutput to final. A ReturnDirectly tool (e.g.
// submit_quotation_for_human_review) terminates the agent run with its
// user-facing result carried as a Role==Tool message; dropping it — as the
// tool-role filter in accumulateMessageOutput would — leaves the answer empty.
//
// The trailing-event predicate is source-verified against Eino v0.7.36 for the
// ChatModelAgent react loop that every attached agent is:
//   - a non-ReturnDirectly tool result is never the trailing event, because
//     compose/tool_node.go ToolsNode.Invoke returns it as a ToolMessage which
//     is sent immediately and the react loop (adk/react.go checkReturnDirect)
//     always routes back to the model, which then emits a further assistant
//     event;
//   - a failing tool never produces a Tool message at all — ToolsNode.Invoke
//     returns an error, the graph run fails, and Generate returns early on
//     event.Err before this helper runs.
// So in the current wiring the ONLY way a Tool message is the trailing event is
// the ReturnDirectly terminal, which carries the user-facing tool result.
//
// A separator is inserted between accumulated assistant text and the appended
// terminal result so the two never merge without whitespace.
func appendTerminalToolResult(final string, last *adk.MessageVariant) string {
	if last == nil || last.Message == nil {
		return final
	}
	if last.Message.Role != schema.Tool || last.Message.Content == "" {
		return final
	}
	if final != "" && !strings.HasSuffix(final, " ") && !strings.HasSuffix(final, "\n") {
		final += "\n\n"
	}
	return final + last.Message.Content
}

// accumulateMessageOutput appends user-facing text from one agent event's
// MessageOutput to final. Intermediate tool-role messages (raw tool results)
// and tool-result stream chunks are skipped so internal tool output never leaks
// into the user-facing answer; only assistant text is accumulated. The terminal
// result of a ReturnDirectly tool is preserved separately by Generate via
// appendTerminalToolResult, because that message is the agent's final output
// rather than an intermediate tool result.
func accumulateMessageOutput(final string, mv *adk.MessageVariant) (string, error) {
	if mv == nil {
		return final, nil
	}
	if msg := mv.Message; msg != nil {
		if msg.Role != schema.Tool && len(msg.Content) > 0 {
			return final + msg.Content, nil
		}
		return final, nil
	}
	if stream := mv.MessageStream; stream != nil {
		for {
			chunk, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				return final, err
			}
			if chunk.Role != schema.Tool && len(chunk.Content) > 0 {
				final += chunk.Content
			}
		}
	}
	return final, nil
}

// HasAgent reports whether the full TradeAgent is attached.
func (c *Client) HasAgent() bool {
	return c.agent != nil
}

// GenerateDirect 直连 ChatModel，不经过 TradeAgent，适用于工具内部等需避免 Agent 递归的场景。
func (c *Client) GenerateDirect(ctx context.Context, prompt string) (string, error) {
	if c.chatModel == nil {
		return "", ErrDisabled
	}
	msgs := []*schema.Message{
		{Role: schema.User, Content: prompt},
	}
	resp, err := c.chatModel.Generate(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("direct generation failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("empty response from chat model")
	}
	return stripThinkBlocks(strings.TrimSpace(resp.Content)), nil
}

// GenerateJSON calls the AI model with response_format=json_object, guaranteeing valid JSON output.
// The prompt MUST contain the word "json" and include an example JSON schema.
func (c *Client) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	if c.chatModelJSON == nil {
		return "", ErrDisabled
	}

	msgs := []*schema.Message{
		{
			Role: schema.System,
			Content: "You are a JSON API assistant. Always respond with valid JSON only. " +
				"No markdown fences, no explanations, no extra text.",
		},
		{Role: schema.User, Content: prompt},
	}

	resp, err := c.chatModelJSON.Generate(ctx, msgs)
	if err != nil {
		return "", fmt.Errorf("json generation failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("empty response from json model")
	}
	return stripThinkBlocks(strings.TrimSpace(resp.Content)), nil
}

// buildClientRunner creates the legacy tool-less agent runner used when no full TradeAgent is attached.
func buildClientRunner(ctx context.Context, chatModel model.ToolCallingChatModel) (*adk.Runner, error) {
	agentConfig := &adk.ChatModelAgentConfig{
		Name:        "CandyProAssistant",
		Description: "CandyPro OEM application assistant specialized in international B2B candy trade.",
		Instruction: buildClientInstruction(),
		Model:       chatModel,
	}

	chatAgent, err := adk.NewChatModelAgent(ctx, agentConfig)
	if err != nil {
		return nil, fmt.Errorf("buildClientRunner: NewChatModelAgent failed: %w", err)
	}

	return adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           chatAgent,
		EnableStreaming: false,
	}), nil
}

func buildClientInstruction() string {
	return agent.AssistantInstructionBase
}
