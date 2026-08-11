package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/eino"
	"candypro/api/internal/pkg/eino/retry"
	"candypro/api/internal/pkg/response"
	tradeSvc "candypro/api/internal/services/trade"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

type SSEEvent struct {
	Type         string            `json:"type"`
	AgentName    string            `json:"agent_name,omitempty"`
	Content      string            `json:"content,omitempty"`
	ToolCalls    []schema.ToolCall `json:"tool_calls,omitempty"`
	ActionType   string            `json:"action_type,omitempty"`
	Error        string            `json:"error,omitempty"`
	DocumentType string            `json:"document_type,omitempty"` // UI Anchor
}

// agentRunOptions reads optional temperature query params and returns Eino run
// options. temperature is honored per-request; model override is not supported
// and the frontend no longer sends it — the effective model is read from
// /system/ai/config (SystemAIConfig), which is the single source of truth.
func (h *Handler) agentRunOptions(c *gin.Context) []adk.AgentRunOption {
	var opts []adk.AgentRunOption
	if raw := strings.TrimSpace(c.Query("temperature")); raw != "" {
		if temp, err := strconv.ParseFloat(raw, 32); err == nil && temp >= 0 && temp <= 2 {
			opts = append(opts, adk.WithChatModelOptions([]model.Option{model.WithTemperature(float32(temp))}))
		}
	}
	return opts
}

// SystemAIConfig reports the effective AI configuration to the admin console.
// GET /api/v1/system/ai/config
func (h *Handler) SystemAIConfig(c *gin.Context) {
	modelName := ""
	enabled := false
	if h.cfg != nil {
		modelName = h.cfg.AI.OpenAIModel
		enabled = h.cfg.AI.IsEnabled()
	}
	c.JSON(http.StatusOK, gin.H{"model": modelName, "enabled": enabled})
}

// InitAgent initializes the trade agent with the Graph Tool architecture.
// The TradeService is wired as the document persister so generated docs are saved to DB.
//
// NOTE: This function directly instantiates the ChatModel, which is a bootstrap concern.
// In a future refactor, this should be moved to a service factory (e.g. services.NewTradeAgent)
// to keep the HTTP layer free of model initialization logic.
func (h *Handler) InitAgent() error {
	ctx := context.Background()

	rawModel, modelErr := eino.NewDeepSeekChatModel(ctx, h.cfg.AI, nil)
	if modelErr != nil {
		return fmt.Errorf("init chat model: %w", modelErr)
	}
	chatModel := retry.New(rawModel, h.cfg.AI.RetryMaxAttempts, h.cfg.AI.RetryIntervalSec)

	if h.services != nil && h.services.Trade != nil {
		a, err := eino.NewTradeAgent(ctx, chatModel, h.services.Trade, h.quotationReviewSaver(), h.translateFunc())
		if err != nil {
			return err
		}
		h.tradeAgent = a
		h.agentReady = true
		return nil
	}

	a, err := eino.NewTradeAgent(ctx, chatModel, nil, h.quotationReviewSaver(), h.translateFunc())
	if err != nil {
		return err
	}
	h.tradeAgent = a
	h.agentReady = true
	return nil
}

func (h *Handler) HandleTradeChat(c *gin.Context) {
	if h.tradeAgent == nil {
		response.ErrorResp(c, http.StatusInternalServerError, "trade_ai_not_configured")
		return
	}

	query := c.Query("query")
	if query == "" {
		response.ErrorResp(c, http.StatusBadRequest, "query_param_required")
		return
	}
	// N-06: Limit query length to prevent cost abuse
	const maxQueryLen = 4000
	// Use rune slicing to avoid breaking multi-byte UTF-8 characters
	if len([]rune(query)) > maxQueryLen {
		query = string([]rune(query)[:maxQueryLen])
	}

	// Enrich query with trade + linked order context when tradeId is provided
	tradeID := strings.TrimSpace(c.Query("tradeId"))
	if tradeID != "" {
		query = h.enrichTradeQuery(c, tradeID, query)
	}

	ctx := c.Request.Context()

	// L5: Add server-side timeout to prevent runaway SSE connections
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Create a new runner for each connection with streaming
	runnerCfg := adk.RunnerConfig{
		EnableStreaming: true,
		Agent:          h.tradeAgent,
	}
	if h.checkPointStore != nil {
		runnerCfg.CheckPointStore = h.checkPointStore
	}
	runner := adk.NewRunner(ctx, runnerCfg)

	// Setup SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// Flush headers immediately
	c.Writer.Flush()

	// Query the runner
	iter := runner.Query(ctx, query, h.agentRunOptions(c)...)

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		if err := processAgentEvent(ctx, c.Writer, event); err != nil {
			log.Printf("SSE Process error: %v", err)
			break
		}
	}
}

// HandleB2BCoordinatorChat handles SSE streaming for the multi-agent B2B coordinator (DeepAgent).
func (h *Handler) HandleB2BCoordinatorChat(c *gin.Context) {
	if h.b2bCoordinatorAgent == nil {
		response.ErrorResp(c, http.StatusInternalServerError, "b2b_coordinator_not_configured")
		return
	}

	query := c.Query("query")
	if query == "" {
		response.ErrorResp(c, http.StatusBadRequest, "query_param_required")
		return
	}
	const maxQueryLen = 4000
	if len([]rune(query)) > maxQueryLen {
		query = string([]rune(query)[:maxQueryLen])
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	runnerCfg := adk.RunnerConfig{
		EnableStreaming: true,
		Agent:          h.b2bCoordinatorAgent,
	}
	if h.checkPointStore != nil {
		runnerCfg.CheckPointStore = h.checkPointStore
	}
	runner := adk.NewRunner(ctx, runnerCfg)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	iter := runner.Query(ctx, query, h.agentRunOptions(c)...)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err := processAgentEvent(ctx, c.Writer, event); err != nil {
			log.Printf("B2B Coordinator SSE error: %v", err)
			break
		}
	}
}

// HandleOrderProcessingChat handles SSE streaming for the Plan-Execute-Replan order processing agent.
func (h *Handler) HandleOrderProcessingChat(c *gin.Context) {
	if h.orderProcessingAgent == nil {
		response.ErrorResp(c, http.StatusInternalServerError, "order_processing_not_configured")
		return
	}

	query := c.Query("query")
	if query == "" {
		response.ErrorResp(c, http.StatusBadRequest, "query_param_required")
		return
	}
	const maxQueryLen = 4000
	if len([]rune(query)) > maxQueryLen {
		query = string([]rune(query)[:maxQueryLen])
	}

	if orderID := strings.TrimSpace(c.Query("orderId")); orderID != "" {
		query = h.enrichOrderQuery(c, orderID, query)
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	runnerCfg := adk.RunnerConfig{
		EnableStreaming: true,
		Agent:          h.orderProcessingAgent,
	}
	if h.checkPointStore != nil {
		runnerCfg.CheckPointStore = h.checkPointStore
	}
	runner := adk.NewRunner(ctx, runnerCfg)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	iter := runner.Query(ctx, query, h.agentRunOptions(c)...)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err := processAgentEvent(ctx, c.Writer, event); err != nil {
			log.Printf("Order Processing SSE error: %v", err)
			break
		}
	}
}

func processAgentEvent(ctx context.Context, w gin.ResponseWriter, event *adk.AgentEvent) error {
	if event.Err != nil {
		sendSSEEvent(w, SSEEvent{
			Type:      "error",
			AgentName: event.AgentName,
			Error:     event.Err.Error(),
		})
		return event.Err
	}

	if event.Output != nil && event.Output.MessageOutput != nil {
		msgOutput := event.Output.MessageOutput
		if msg := msgOutput.Message; msg != nil {
			eventType := "message"
			if msg.Role == schema.Tool {
				eventType = "tool_result"
			}
			sseEvent := SSEEvent{
				Type:      eventType,
				AgentName: event.AgentName,
				Content:   msg.Content,
			}
			if len(msg.ToolCalls) > 0 {
				sseEvent.ToolCalls = msg.ToolCalls
				// Inject the document_type anchor based on tool name
				toolName := msg.ToolCalls[0].Function.Name
				switch toolName {
				case "generate_proforma_invoice":
					sseEvent.DocumentType = "PROFORMA_INVOICE"
				case "generate_commercial_invoice":
					sseEvent.DocumentType = "COMMERCIAL_INVOICE"
				case "generate_packing_list":
					sseEvent.DocumentType = "PACKING_LIST"
				case "generate_certificate_of_origin":
					sseEvent.DocumentType = "ORIGIN_CERTIFICATE"
				case "generate_sales_contract":
					sseEvent.DocumentType = "SALES_CONTRACT"
				case "generate_health_certificate_request":
					sseEvent.DocumentType = "HEALTH_CERTIFICATE"
				case "generate_ingredients_declaration":
					sseEvent.DocumentType = "INGREDIENTS_DECLARATION"
				case "generate_shipper_letter_of_instruction":
					sseEvent.DocumentType = "SHIPMENT_INSTRUCTION"
				case "validate_lc_documents":
					sseEvent.DocumentType = "LC_VALIDATION"
				case "generate_insurance_certificate_request":
					sseEvent.DocumentType = "INSURANCE_CERTIFICATE"
				case "track_shipment":
					sseEvent.DocumentType = "SHIPMENT_TRACKING"
				case "generate_trade_documents":
					if docType := eino.FirstDocTypeFromArgs(msg.ToolCalls[0].Function.Arguments); docType != "" {
						sseEvent.DocumentType = docType
					} else {
						sseEvent.DocumentType = eino.FallbackTradeDocumentType
					}
				}
			}
			sendSSEEvent(w, sseEvent)
		}

		if stream := msgOutput.MessageStream; stream != nil {
			for {
				chunk, err := stream.Recv()
				if errors.Is(err, io.EOF) {
					break
				}
				if err != nil {
					sendSSEEvent(w, SSEEvent{Type: "error", Error: err.Error()})
					return err
				}
				if chunk.Content != "" {
					evtType := "stream_chunk"
					if chunk.Role == schema.Tool {
						evtType = "tool_result_chunk"
					}
					sendSSEEvent(w, SSEEvent{Type: evtType, Content: chunk.Content})
				}
			}
		}
	}

	if event.Action != nil {
		if event.Action.Interrupted != nil {
			for _, ic := range event.Action.Interrupted.InterruptContexts {
				content := fmt.Sprintf("%v", ic.Info)
				sendSSEEvent(w, SSEEvent{
					Type:       "action",
					ActionType: "interrupted",
					Content:    content,
				})
			}
		}
		if event.Action.Exit {
			sendSSEEvent(w, SSEEvent{Type: "action", ActionType: "exit", Content: "Done"})
		}
	}

	return nil
}

func sendSSEEvent(w gin.ResponseWriter, event SSEEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	// Write standard SSE format
	_, err = fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
	return err
}

// enrichTradeQuery 注入贸易头信息及关联订单行项目，供 TradeAgent 使用。
// Ownership is enforced exactly like resolveOrderForUser (IDOR #13): admins may
// enrich any trade, a customer only their own. A non-owner gets no trade context,
// so another customer's trade/order pricing cannot leak into an AI response.
func (h *Handler) enrichTradeQuery(c *gin.Context, tradeIDStr, query string) string {
	if h.services == nil || h.services.Trade == nil {
		return fmt.Sprintf("[Trade ID: %s] %s", tradeIDStr, query)
	}
	id, err := strconv.ParseUint(tradeIDStr, 10, 64)
	if err != nil {
		return fmt.Sprintf("[Trade ID: %s] %s", tradeIDStr, query)
	}
	trade, err := h.services.Trade.GetTransaction(c.Request.Context(), uint(id))
	if err != nil {
		return fmt.Sprintf("[Trade ID: %s] %s", tradeIDStr, query)
	}
	userID, role := authContext(c)
	if !isAdminPortalRole(role) && strings.TrimSpace(trade.UserID) != userID {
		return fmt.Sprintf("[Trade ID: %s] %s", tradeIDStr, query)
	}
	var order *modelsOrder.Order
	if h.services.Order != nil && trade.OrderID != nil && strings.TrimSpace(*trade.OrderID) != "" {
		order, _ = h.services.Order.GetOrder(c.Request.Context(), strings.TrimSpace(*trade.OrderID))
	}
	if block := tradeSvc.BuildTradeOrderContextBlock(trade, order); block != "" {
		return block + "\n\nUser question:\n" + query
	}
	return fmt.Sprintf("[Trade ID: %s] %s", tradeIDStr, query)
}
