package system

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	modelsAuth "candypro/api/internal/models/auth"
	"candypro/api/internal/pkg/eino"
	einotool "candypro/api/internal/pkg/eino/tool"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// resumeHandlerTestAgent drives the real Eino interrupt/restore machinery and
// review tool without contacting a model provider or writing trade documents.
type resumeHandlerTestAgent struct {
	reviewTool *einotool.InvokableReviewEditTool
}

func (agent *resumeHandlerTestAgent) Name(context.Context) string { return "TradeAssistant" }
func (agent *resumeHandlerTestAgent) Description(context.Context) string {
	return "isolated resume regression"
}
func (agent *resumeHandlerTestAgent) Run(ctx context.Context, _ *adk.AgentInput, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return agent.invoke(ctx)
}
func (agent *resumeHandlerTestAgent) Resume(ctx context.Context, _ *adk.ResumeInfo, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return agent.invoke(ctx)
}
func (agent *resumeHandlerTestAgent) invoke(ctx context.Context) *adk.AsyncIterator[*adk.AgentEvent] {
	iterator, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	go func() {
		defer generator.Close()
		toolContext := adk.AppendAddressSegment(ctx, adk.AddressSegmentTool, "generate_trade_documents")
		output, err := agent.reviewTool.InvokableRun(toolContext, `{"total_amount":10000}`)
		if err != nil {
			var interrupt *adk.InterruptSignal
			if errors.As(err, &interrupt) {
				generator.Send(adk.CompositeInterrupt(ctx, nil, nil, interrupt))
				return
			}
			generator.Send(&adk.AgentEvent{Err: err})
			return
		}
		generator.Send(&adk.AgentEvent{Output: &adk.AgentOutput{MessageOutput: &adk.MessageVariant{Message: schema.ToolMessage(output, "test-document-tool")}}})
	}()
	return iterator
}

func TestResumeAgent_AuthorizationAndDuplicateEffects(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	connection, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	connection.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = connection.Close() })
	store := eino.NewPostgresCheckPointStore(database)
	if err := eino.AutoMigrateErr(); err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	release := make(chan struct{})
	var documentWrites atomic.Int32
	baseTool, err := utils.InferTool("generate_trade_documents", "isolated document effect", func(context.Context, struct {
		TotalAmount float64 `json:"total_amount"`
	}) (string, error) {
		documentWrites.Add(1)
		close(entered)
		<-release
		return "document created", nil
	})
	if err != nil {
		t.Fatal(err)
	}
	agent := &resumeHandlerTestAgent{reviewTool: &einotool.InvokableReviewEditTool{InvokableTool: baseTool, RequiresReview: func(string, string) bool { return true }}}
	handler := &Handler{tradeAgent: agent, checkPointStore: store}
	checkpointID := eino.NewOwnedCheckPointID("admin-owner")
	runner := adk.NewRunner(context.Background(), adk.RunnerConfig{Agent: agent, CheckPointStore: store})
	iterator := runner.Query(context.Background(), "generate documents", adk.WithCheckPointID(checkpointID))
	interruptID := ""
	for {
		event, available := iterator.Next()
		if !available {
			break
		}
		if event.Err != nil {
			t.Fatal(event.Err)
		}
		if event.Action != nil && event.Action.Interrupted != nil {
			for _, interrupt := range event.Action.Interrupted.InterruptContexts {
				if interrupt.IsRootCause {
					interruptID = interrupt.ID
				}
			}
		}
	}
	if interruptID == "" {
		t.Fatal("expected a pending review")
	}
	requestBody, err := json.Marshal(resumeAgentRequest{CheckpointID: checkpointID, InterruptID: interruptID, Action: resumeActionApprove, Agent: "trade-assistant"})
	if err != nil {
		t.Fatal(err)
	}
	invoke := func(userID, role string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		requestContext, _ := gin.CreateTestContext(response)
		requestContext.Set("userID", userID)
		requestContext.Set("userRole", role)
		requestContext.Request = httptest.NewRequest(http.MethodPost, "/api/v1/system/ai/resume", strings.NewReader(string(requestBody)))
		requestContext.Request.Header.Set("Content-Type", "application/json")
		handler.ResumeAgent(requestContext)
		return response
	}
	for _, caller := range []struct{ userID, role string }{{"customer", modelsAuth.User}, {"admin-other", modelsAuth.Admin}, {"", ""}} {
		if response := invoke(caller.userID, caller.role); response.Code != http.StatusForbidden {
			t.Fatalf("unauthorized caller %q response=%d", caller.userID, response.Code)
		}
	}
	// A non-targeted interrupt must survive under a new owned key while the
	// consumed key stays closed. Exercise the runner's actual checkpoint write.
	requestBody, err = json.Marshal(resumeAgentRequest{CheckpointID: checkpointID, InterruptID: "different-interrupt", Action: resumeActionApprove, Agent: "trade-assistant"})
	if err != nil {
		t.Fatal(err)
	}
	interruptedResponse := invoke("admin-owner", modelsAuth.Admin)
	var continuation SSEEvent
	for _, line := range strings.Split(interruptedResponse.Body.String(), "\n") {
		if payload, found := strings.CutPrefix(line, "data: "); found {
			var event SSEEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				t.Fatal(err)
			}
			if event.ActionType == "interrupted" {
				continuation = event
			}
		}
	}
	if continuation.CheckpointID == checkpointID || !eino.CheckPointOwnedBy(continuation.CheckpointID, "admin-owner") || continuation.InterruptID == "" {
		t.Fatalf("missing rotated continuation: %s", interruptedResponse.Body.String())
	}
	if response := invoke("admin-owner", modelsAuth.Admin); response.Code != http.StatusConflict {
		t.Fatalf("consumed parent checkpoint response=%d", response.Code)
	}
	requestBody, err = json.Marshal(resumeAgentRequest{CheckpointID: continuation.CheckpointID, InterruptID: continuation.InterruptID, Action: resumeActionApprove, Agent: "trade-assistant"})
	if err != nil {
		t.Fatal(err)
	}
	completed := make(chan *httptest.ResponseRecorder, 1)
	go func() { completed <- invoke("admin-owner", modelsAuth.Admin) }()
	<-entered
	duplicate := invoke("admin-owner", modelsAuth.Admin)
	close(release)
	first := <-completed
	if duplicate.Code != http.StatusConflict {
		t.Fatalf("concurrent resume status=%d", duplicate.Code)
	}
	if first.Code != http.StatusOK || !strings.Contains(first.Body.String(), "document created") {
		t.Fatalf("authorized resume=%d %s", first.Code, first.Body.String())
	}
	if replay := invoke("admin-owner", modelsAuth.Admin); replay.Code != http.StatusConflict {
		t.Fatalf("completed checkpoint replay=%d", replay.Code)
	}
	if documentWrites.Load() != 1 {
		t.Fatalf("document effects=%d", documentWrites.Load())
	}
}
