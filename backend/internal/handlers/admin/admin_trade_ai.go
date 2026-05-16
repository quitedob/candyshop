package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	tradeModels "candypro/api/internal/models/trade"
	"candypro/api/internal/pkg/response"
	tradeSvc "candypro/api/internal/services/trade"

	"github.com/cloudwego/eino/adk"
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

	prompt := tradeSvc.BuildTradeDocGenerationPrompt(docType, trade, req.Prompt, req.Context)
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
	"PROFORMA_INVOICE":       tradeModels.DocTypeProformaInvoice,
	"COMMERCIAL_INVOICE":     tradeModels.DocTypeCommercialInvoice,
	"SALES_CONTRACT":         tradeModels.DocTypeSalesContract,
	"PACKING_LIST":           tradeModels.DocTypePackingList,
	"ORIGIN_CERTIFICATE":     tradeModels.DocTypeOriginCertificate,
	"HEALTH_CERTIFICATE":     tradeModels.DocTypeHealthCertificate,
	"BILL_OF_LADING":         tradeModels.DocTypeBillOfLading,
	"INGREDIENTS_DECLARATION": "INGREDIENTS_DECLARATION",
	"SHIPPER_LETTER_OF_INSTRUCTION": "SHIPPER_LETTER_OF_INSTRUCTION",
	"INSURANCE_CERTIFICATE":  "INSURANCE_CERTIFICATE",
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
	if len(query) > maxQueryLen {
		query = query[:maxQueryLen]
	}
	query = fmt.Sprintf("[Trade ID: %d] %s", tradeID, query)

	agent := h.tradeAgent
	if agent == nil {
		// Fallback to non-streaming
		if h.aiService == nil || !h.aiService.IsEnabled() {
			response.ErrorResp(c, http.StatusServiceUnavailable, "ai_not_configured")
			return
		}
		reply, genErr := h.aiService.Generate(c.Request.Context(), query)
		if genErr != nil {
			response.ErrorResp(c, http.StatusInternalServerError, "ai_chat_failed")
			return
		}
		c.JSON(http.StatusOK, gin.H{"reply": reply, "tradeId": tradeID})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
	defer cancel()

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           agent,
	})

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Flush()

	iter := runner.Query(ctx, query)
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err := processAdminAgentEvent(c.Writer, event); err != nil {
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
}

func processAdminAgentEvent(w gin.ResponseWriter, event *adk.AgentEvent) error {
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
				if len(msg.ToolCalls) > 0 {
					switch msg.ToolCalls[0].Function.Name {
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
					}
				}
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

	if event.Action != nil && event.Action.Exit {
		sendAdminSSEEvent(w, adminSSEEvent{Type: "action", ActionType: "exit", Content: "Done"})
	}

	return nil
}

func sendAdminSSEEvent(w gin.ResponseWriter, event adminSSEEvent) {
	data, _ := json.Marshal(event)
	fmt.Fprintf(w, "data: %s\n\n", data)
	w.Flush()
}
