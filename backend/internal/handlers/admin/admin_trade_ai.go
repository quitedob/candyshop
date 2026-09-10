package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	tradeModels "candypro/api/internal/models/trade"
	"candypro/api/internal/pkg/eino"
	einotool "candypro/api/internal/pkg/eino/tool"
	"candypro/api/internal/pkg/response"
	tradeSvc "candypro/api/internal/services/trade"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// AdminAIGenerateTradeDocument uses the TradeAgent to generate a trade document (PI, CI, SC, etc.)
// and persist it to the database. POST /admin/trades/:id/documents/ai-generate
func (h *Handler) AdminAIGenerateTradeDocument(c *gin.Context) {
	if h.aiService == nil || !h.aiService.IsEnabled() {
		response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
		return
	}

	tradeID, err := parseUintParam(c, "id")
	if err != nil {
		response.InvalidResp(c, "invalid_transaction_id")
		return
	}

	var req struct {
		DocType string `json:"docType" binding:"required"`
		Prompt  string `json:"prompt"`
		Context string `json:"context"` // additional context from admin
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	docType := strings.ToUpper(strings.TrimSpace(req.DocType))
	if !isSupportedDocType(docType) {
		response.InvalidResp(c, "unsupported_doc_type")
		return
	}

	// Fetch the trade transaction for context injection
	trade, fetchErr := h.services.Trade.GetTransaction(c.Request.Context(), tradeID)
	if fetchErr != nil {
		response.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return
	}

	order := h.loadOrderForTrade(c.Request.Context(), trade)
	orderCtx := tradeSvc.BuildOrderContextJSON(order)
	prompt := tradeSvc.BuildTradeDocGenerationPrompt(docType, trade, req.Prompt, req.Context, orderCtx)
	reply, genErr := h.aiService.Generate(c.Request.Context(), prompt)
	if genErr != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "ai_generation_failed")
		return
	}

	// Re-fetch documents to include what the agent just created
	docs, _ := h.services.Trade.ListDocuments(c.Request.Context(), tradeID)

	c.JSON(http.StatusOK, gin.H{
		"message":       "AI generation complete",
		"aiResponse":    strings.TrimSpace(reply),
		"docType":       docType,
		"tradeId":       tradeID,
		"documents":     docs,
		"generatedHint": "Check documents below — AI may have created new drafts via tools.",
	})
}

var supportedDocTypes = map[string]string{
	"PROFORMA_INVOICE":              tradeModels.DocTypeProformaInvoice,
	"COMMERCIAL_INVOICE":            tradeModels.DocTypeCommercialInvoice,
	"SALES_CONTRACT":                tradeModels.DocTypeSalesContract,
	"PACKING_LIST":                  tradeModels.DocTypePackingList,
	"ORIGIN_CERTIFICATE":            tradeModels.DocTypeOriginCertificate,
	"HEALTH_CERTIFICATE":            tradeModels.DocTypeHealthCertificate,
	"BILL_OF_LADING":                tradeModels.DocTypeBillOfLading,
	"INGREDIENTS_DECLARATION":       "INGREDIENTS_DECLARATION",
	"SHIPPER_LETTER_OF_INSTRUCTION": "SHIPPER_LETTER_OF_INSTRUCTION",
	"INSURANCE_CERTIFICATE":         "INSURANCE_CERTIFICATE",
}

func isSupportedDocType(docType string) bool {
	_, ok := supportedDocTypes[docType]
	return ok
}

// AdminAITradeChat handles admin-side SSE trade chat with the full TradeAgent.
// GET /admin/trades/:id/ai-chat
func (h *Handler) AdminAITradeChat(c *gin.Context) {
	tradeID, err := parseUintParam(c, "id")
	if err != nil {
		response.InvalidResp(c, "invalid_transaction_id")
		return
	}

	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter required"})
		return
	}
	const maxQueryLen = 4000
	runes := []rune(query)
	if len(runes) > maxQueryLen {
		query = string(runes[:maxQueryLen])
	}

	trade, fetchErr := h.services.Trade.GetTransaction(c.Request.Context(), tradeID)
	if fetchErr != nil {
		response.ErrorResp(c, http.StatusNotFound, "trade_not_found")
		return
	}
	order := h.loadOrderForTrade(c.Request.Context(), trade)
	if block := tradeSvc.BuildTradeOrderContextBlock(trade, order); block != "" {
		query = block + "\n\nUser question:\n" + query
	} else {
		query = fmt.Sprintf("[Trade ID: %d] %s", tradeID, query)
	}

	agent := h.tradeAgent
	if agent == nil {
		// Fallback to non-streaming
		log.Printf("Error: trade AI agent unavailable, /admin/trades/:id/ai-chat falling back to non-streaming mode")
		if h.aiService == nil || !h.aiService.IsEnabled() {
			response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
			return
		}
		reply, genErr := h.aiService.Generate(c.Request.Context(), query)
		if genErr != nil {
			log.Printf("Error: trade AI non-streaming fallback generation failed: %v", genErr)
			response.ErrorResp(c, http.StatusInternalServerError, "ai_chat_failed")
			return
		}
		c.JSON(http.StatusOK, gin.H{"reply": reply, "tradeId": tradeID})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	runnerCfg := adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           agent,
	}
	if h.checkPointStore != nil {
		runnerCfg.CheckPointStore = h.checkPointStore
	}
	runner := adk.NewRunner(ctx, runnerCfg)

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	var opts []adk.AgentRunOption
	if raw := strings.TrimSpace(c.Query("temperature")); raw != "" {
		if temp, err := strconv.ParseFloat(raw, 32); err == nil && temp >= 0 && temp <= 2 {
			opts = append(opts, adk.WithChatModelOptions([]model.Option{model.WithTemperature(float32(temp))}))
		}
	}
	// A fresh owned run prevents concurrent conversations overwriting checkpoints.
	checkpointID := eino.NewOwnedCheckPointID(c.GetString("userID"))
	opts = append(opts, adk.WithCheckPointID(checkpointID))
	iter := runner.Query(ctx, query, opts...)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err := processAdminAgentEvent(c.Writer, event, checkpointID); err != nil {
			log.Printf("Admin SSE error: %v", err)
			break
		}
	}
}

type adminSSEEvent struct {
	Type         string            `json:"type"`
	AgentName    string            `json:"agent_name,omitempty"`
	Content      string            `json:"content,omitempty"`
	ToolCalls    []schema.ToolCall `json:"tool_calls,omitempty"`
	ActionType   string            `json:"action_type,omitempty"`
	Error        string            `json:"error,omitempty"`
	DocumentType string            `json:"document_type,omitempty"`

	// HITL review-and-edit resume fields (same shape as the system SSE handler).
	CheckpointID     string          `json:"checkpoint_id,omitempty"`
	InterruptID      string          `json:"interrupt_id,omitempty"`
	InterruptAddress string          `json:"interrupt_address,omitempty"`
	Review           json.RawMessage `json:"review,omitempty"`
}

func processAdminAgentEvent(w gin.ResponseWriter, event *adk.AgentEvent, checkpointID string) error {
	if event.Err != nil {
		sendAdminSSEEvent(w, adminSSEEvent{
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
			sseEvent := adminSSEEvent{
				Type:      eventType,
				AgentName: event.AgentName,
				Content:   msg.Content,
			}
			if len(msg.ToolCalls) > 0 {
				sseEvent.ToolCalls = msg.ToolCalls
				sseEvent.DocumentType = mapAdminToolToDocType(msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)
			}
			sendAdminSSEEvent(w, sseEvent)
		}

		if stream := msgOutput.MessageStream; stream != nil {
			for {
				chunk, err := stream.Recv()
				if err == io.EOF {
					break
				}
				if err != nil {
					sendAdminSSEEvent(w, adminSSEEvent{Type: "error", Error: err.Error()})
					return err
				}
				if chunk.Content != "" {
					evtType := "stream_chunk"
					if chunk.Role == schema.Tool {
						evtType = "tool_result_chunk"
					}
					sendAdminSSEEvent(w, adminSSEEvent{Type: evtType, Content: chunk.Content})
				}
			}
		}
	}

	if event.Action != nil {
		if event.Action.Interrupted != nil {
			for _, ic := range event.Action.Interrupted.InterruptContexts {
				sseEvent := adminSSEEvent{
					Type:       "action",
					ActionType: "interrupted",
					Content:    fmt.Sprintf("%v", ic.Info),
				}
				if info, ok := ic.Info.(*einotool.ReviewEditInfo); ok {
					sseEvent.CheckpointID = checkpointID
					sseEvent.InterruptID = ic.ID
					sseEvent.InterruptAddress = ic.Address.String()
					if reviewJSON, mErr := json.Marshal(struct {
						ToolName        string `json:"tool_name"`
						ArgumentsInJSON string `json:"arguments_in_json"`
					}{
						ToolName:        info.ToolName,
						ArgumentsInJSON: info.ArgumentsInJSON,
					}); mErr == nil {
						sseEvent.Review = reviewJSON
					}
				}
				sendAdminSSEEvent(w, sseEvent)
			}
		}
		if event.Action.Exit {
			sendAdminSSEEvent(w, adminSSEEvent{Type: "action", ActionType: "exit", Content: "Done"})
		}
	}

	return nil
}

func mapAdminToolToDocType(toolName, argsJSON string) string {
	switch toolName {
	case "generate_proforma_invoice":
		return tradeModels.DocTypeProformaInvoice
	case "generate_trade_documents":
		// A batch may request arbitrary doc types; resolve the anchor from the
		// tool-call args instead of assuming PROFORMA_INVOICE (aligns with the
		// system SSE handler).
		if docType := eino.FirstDocTypeFromArgs(argsJSON); docType != "" {
			return docType
		}
		return eino.FallbackTradeDocumentType
	case "generate_commercial_invoice":
		return tradeModels.DocTypeCommercialInvoice
	case "generate_packing_list":
		return tradeModels.DocTypePackingList
	case "generate_certificate_of_origin":
		return tradeModels.DocTypeOriginCertificate
	case "generate_sales_contract":
		return tradeModels.DocTypeSalesContract
	case "generate_health_certificate_request":
		return tradeModels.DocTypeHealthCertificate
	case "generate_bill_of_lading":
		return tradeModels.DocTypeBillOfLading
	case "generate_ingredients_declaration":
		return "INGREDIENTS_DECLARATION"
	case "generate_shipper_letter_of_instruction":
		return "SHIPPER_LETTER_OF_INSTRUCTION"
	case "generate_insurance_certificate_request":
		return "INSURANCE_CERTIFICATE"
	default:
		return ""
	}
}

func sendAdminSSEEvent(w gin.ResponseWriter, event adminSSEEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
}

// loadOrderForTrade 按贸易关联 orderId 加载订单行项目等业务数据。
func (h *Handler) loadOrderForTrade(ctx context.Context, trade *tradeModels.TradeTransaction) *modelsOrder.Order {
	if trade == nil || trade.OrderID == nil || h.services == nil || h.services.Order == nil {
		return nil
	}
	orderID := strings.TrimSpace(*trade.OrderID)
	if orderID == "" {
		return nil
	}
	order, err := h.services.Order.GetOrder(ctx, orderID)
	if err != nil {
		return nil
	}
	return order
}
