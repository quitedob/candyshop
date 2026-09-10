package system

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"candypro/api/internal/pkg/eino"
	einotool "candypro/api/internal/pkg/eino/tool"
	"candypro/api/internal/pkg/response"

	"github.com/cloudwego/eino/adk"
	"github.com/gin-gonic/gin"
)

// Resume actions accepted by the HITL review-and-edit resume endpoint.
const (
	resumeActionApprove    = "approve"
	resumeActionDisapprove = "disapprove"
	resumeActionEdit       = "edit"
	resumeTimeout          = 5 * time.Minute
)

// resumeAgentRequest is the body of POST /api/v1/system/ai/resume. It carries the
// checkpoint + interrupt identifiers emitted in the "interrupted" SSE event, the
// review decision, and (optionally) edited tool arguments for the "edit" action.
type resumeAgentRequest struct {
	CheckpointID     string `json:"checkpointId" binding:"required"`
	InterruptID      string `json:"interruptId" binding:"required"`
	InterruptAddress string `json:"interrupt_address"`         // optional; source for agent inference (from SSE "interrupted" event)
	Action           string `json:"action" binding:"required"` // approve | disapprove | edit
	EditedArguments  string `json:"editedArguments"`           // JSON, required for edit
	DisapproveReason string `json:"disapproveReason"`          // optional for disapprove
	Agent            string `json:"agent"`                     // optional; explicit override, otherwise inferred from interrupt_address
}

// ResumeAgent resumes an interrupted HITL review-and-edit agent run. It rebuilds
// the same runner (same agent + checkpoint store) and re-enters the interrupted
// tool with the user's decision, streaming the continuation as SSE.
func (h *Handler) ResumeAgent(c *gin.Context) {
	userID, role := authContext(c)
	if userID == "" || !isAdminPortalRole(role) {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}
	if h.checkPointStore == nil {
		response.ErrorResp(c, http.StatusServiceUnavailable, "agent_checkpoint_store_unavailable")
		return
	}

	var req resumeAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if strings.TrimSpace(req.CheckpointID) == "" || strings.TrimSpace(req.InterruptID) == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	if !eino.CheckPointOwnedBy(req.CheckpointID, userID) {
		response.ErrorResp(c, http.StatusForbidden, "forbidden")
		return
	}

	agent, ok := h.resumeAgentFor(req.Agent, req.InterruptAddress)
	if !ok {
		response.ErrorResp(c, http.StatusBadRequest, "unknown_agent")
		return
	}

	reviewResult := buildReviewEditResult(req)
	if reviewResult == nil {
		response.ErrorResp(c, http.StatusBadRequest, "unsupported_resume_action")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), resumeTimeout)
	defer cancel()
	claimed, err := h.checkPointStore.ClaimResume(ctx, req.CheckpointID)
	if err != nil {
		log.Printf("ResumeAgent: checkpoint claim failed: %v", err)
		response.ErrorResp(c, http.StatusServiceUnavailable, "resume_unavailable")
		return
	}
	if !claimed {
		response.ErrorResp(c, http.StatusConflict, "checkpoint_unavailable_or_consumed")
		return
	}
	continuationID := eino.NewOwnedCheckPointID(userID)
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           agent,
		CheckPointStore: h.checkPointStore.WithContinuationID(continuationID),
	})

	iter, err := runner.ResumeWithParams(ctx, req.CheckpointID, &adk.ResumeParams{
		Targets: map[string]any{
			req.InterruptID: &einotool.ReviewEditInfo{ReviewResult: reviewResult},
		},
	})
	if err != nil {
		log.Printf("ResumeAgent: resume failed: %v", err)
		response.ErrorResp(c, http.StatusBadRequest, "resume_failed")
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if err := processAgentEvent(ctx, c.Writer, event, continuationID); err != nil {
			log.Printf("ResumeAgent: SSE process error: %v", err)
			break
		}
	}
}

// buildReviewEditResult maps the client action to the ReviewEditResult fields the
// review_edit.go resume path (InvokableRun) understands. Returns nil for an
// unsupported action so the handler can reject it before building a runner.
func buildReviewEditResult(req resumeAgentRequest) *einotool.ReviewEditResult {
	switch strings.ToLower(strings.TrimSpace(req.Action)) {
	case resumeActionDisapprove:
		result := &einotool.ReviewEditResult{Disapproved: true}
		if reason := strings.TrimSpace(req.DisapproveReason); reason != "" {
			result.DisapproveReason = &reason
		}
		return result
	case resumeActionEdit:
		edited := strings.TrimSpace(req.EditedArguments)
		if edited == "" {
			return nil
		}
		return &einotool.ReviewEditResult{EditedArgumentsInJSON: &edited}
	case resumeActionApprove:
		return &einotool.ReviewEditResult{NoNeedToEdit: true}
	default:
		return nil
	}
}

// resumeAgentFor selects the agent to resume. It prefers an explicit Agent field
// (frontend mode name such as "trade-assistant") and otherwise infers it from the
// interrupt address's leading "agent:<Name>" segment (e.g. "TradeAssistant").
func (h *Handler) resumeAgentFor(agentField, interruptAddress string) (adk.Agent, bool) {
	name := strings.TrimSpace(agentField)
	if name == "" {
		name = agentNameFromAddress(interruptAddress)
	}

	switch {
	case strings.EqualFold(name, "trade-assistant") || strings.EqualFold(name, "TradeAssistant"):
		return h.tradeAgent, h.tradeAgent != nil
	case strings.EqualFold(name, "b2b-coordinator") || strings.EqualFold(name, "B2BTradeCoordinator"):
		return h.b2bCoordinatorAgent, h.b2bCoordinatorAgent != nil
	case strings.EqualFold(name, "order-processing") || strings.EqualFold(name, "plan_execute_replan"):
		return h.orderProcessingAgent, h.orderProcessingAgent != nil
	default:
		return nil, false
	}
}

// agentNameFromAddress extracts the top-level agent name from an interrupt address
// emitted in the SSE "interrupted" event (ic.Address.String()). The address is the
// string form of the Eino execution address, e.g.
// "agent:TradeAssistant;tool:generate_trade_documents", where each segment is
// "type:id[:subID]" joined by ";". The first "agent:" segment is the root agent that
// must be rebuilt to resume the run; its id (e.g. "TradeAssistant") is returned.
func agentNameFromAddress(address string) string {
	for _, seg := range strings.Split(address, ";") {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		typ, rest, ok := strings.Cut(seg, ":")
		if !ok || typ != "agent" {
			continue
		}
		if id, _, _ := strings.Cut(rest, ":"); strings.TrimSpace(id) != "" {
			return strings.TrimSpace(id)
		}
	}
	return ""
}
