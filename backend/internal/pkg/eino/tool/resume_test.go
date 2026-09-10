package tool

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

// reviewGateAgent is a minimal adk.ResumableAgent that wraps an
// InvokableReviewEditTool directly, so the resume path of InvokableRun can be
// exercised in-process through the real Runner checkpoint/restore machinery
// without a chat model. It mirrors the tool-node address handling that a
// ChatModelAgent performs (append a tool address segment, then invoke).
type reviewGateAgent struct {
	toolName string
	tool     *InvokableReviewEditTool
	argsJSON string
}

func (a *reviewGateAgent) Name(context.Context) string        { return "reviewGateAgent" }
func (a *reviewGateAgent) Description(context.Context) string { return "test review gate" }

func (a *reviewGateAgent) invoke(ctx context.Context) (string, error) {
	ctx = adk.AppendAddressSegment(ctx, adk.AddressSegmentTool, a.toolName)
	return a.tool.InvokableRun(ctx, a.argsJSON)
}

func (a *reviewGateAgent) run(ctx context.Context) *adk.AsyncIterator[*adk.AgentEvent] {
	iter, gen := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	go func() {
		defer gen.Close()
		out, err := a.invoke(ctx)
		if err != nil {
			var sig *adk.InterruptSignal
			if errors.As(err, &sig) {
				gen.Send(adk.CompositeInterrupt(ctx, nil, nil, sig))
				return
			}
			gen.Send(&adk.AgentEvent{Err: err})
			return
		}
		gen.Send(&adk.AgentEvent{
			Output: &adk.AgentOutput{
				MessageOutput: &adk.MessageVariant{Message: schema.UserMessage(out)},
			},
		})
	}()
	return iter
}

func (a *reviewGateAgent) Run(ctx context.Context, _ *adk.AgentInput, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return a.run(ctx)
}

func (a *reviewGateAgent) Resume(ctx context.Context, _ *adk.ResumeInfo, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return a.run(ctx)
}

type memCheckPointStore struct {
	mu sync.Mutex
	m  map[string][]byte
}

func (s *memCheckPointStore) Get(_ context.Context, key string) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.m[key]
	return b, ok, nil
}

func (s *memCheckPointStore) Set(_ context.Context, key string, val []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.m == nil {
		s.m = map[string][]byte{}
	}
	s.m[key] = val
	return nil
}

func newReviewGateTool(t *testing.T) *InvokableReviewEditTool {
	t.Helper()
	base, err := utils.InferTool("generate_trade_documents",
		"generate trade documents after human review",
		func(_ context.Context, req struct {
			TotalAmount float64 `json:"total_amount"`
		}) (string, error) {
			return fmt.Sprintf("generated amount=%.0f", req.TotalAmount), nil
		})
	if err != nil {
		t.Fatalf("InferTool: %v", err)
	}
	return &InvokableReviewEditTool{
		InvokableTool:  base,
		RequiresReview: func(string, string) bool { return true },
	}
}

func strPtr(s string) *string { return &s }

// drainMessages consumes an event iterator until close, returning the concatenated
// message contents and the first root-cause interrupt ID (if any).
func drainMessages(t *testing.T, iter *adk.AsyncIterator[*adk.AgentEvent]) (messages, interruptID string) {
	t.Helper()
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			t.Fatalf("unexpected event error: %v", event.Err)
		}
		if event.Output != nil && event.Output.MessageOutput != nil && event.Output.MessageOutput.Message != nil {
			messages += event.Output.MessageOutput.Message.Content
		}
		if interruptID == "" && event.Action != nil && event.Action.Interrupted != nil {
			for _, ic := range event.Action.Interrupted.InterruptContexts {
				if ic.IsRootCause {
					interruptID = ic.ID
				}
			}
		}
	}
	return messages, interruptID
}

// TestInvokableReviewEditTool_ResumeBranches drives the full interrupt → resume
// cycle through the Runner, then resumes with each ReviewEditResult branch and
// asserts the tool behaves accordingly. This is the missing coverage for the
// resume path of InvokableRun.
func TestInvokableReviewEditTool_ResumeBranches(t *testing.T) {
	const checkpointID = "review-cp-1"

	tests := []struct {
		name       string
		resumeData *ReviewEditResult
		want       string
	}{
		{
			name:       "NoNeedToEdit runs with stored arguments",
			resumeData: &ReviewEditResult{NoNeedToEdit: true},
			want:       "generated amount=10000",
		},
		{
			name:       "EditedArgumentsInJSON runs with edited arguments",
			resumeData: &ReviewEditResult{EditedArgumentsInJSON: strPtr(`{"total_amount":20000}`)},
			want:       "generated amount=20000",
		},
		{
			name:       "Disapproved does not run the tool",
			resumeData: &ReviewEditResult{Disapproved: true},
			want:       "disapproved",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gateTool := newReviewGateTool(t)
			agent := &reviewGateAgent{
				toolName: "generate_trade_documents",
				tool:     gateTool,
				argsJSON: `{"total_amount":10000}`,
			}
			runner := adk.NewRunner(context.Background(), adk.RunnerConfig{
				EnableStreaming: false,
				Agent:           agent,
				CheckPointStore: &memCheckPointStore{},
			})

			ctx := context.Background()
			iter := runner.Query(ctx, "generate documents", adk.WithCheckPointID(checkpointID))

			_, interruptID := drainMessages(t, iter)
			if interruptID == "" {
				t.Fatal("expected a review interrupt, got none")
			}

			iter, err := runner.ResumeWithParams(ctx, checkpointID, &adk.ResumeParams{
				Targets: map[string]any{
					interruptID: &ReviewEditInfo{ReviewResult: tc.resumeData},
				},
			})
			if err != nil {
				t.Fatalf("ResumeWithParams: %v", err)
			}

			got, _ := drainMessages(t, iter)
			if !strings.Contains(got, tc.want) {
				t.Fatalf("resume output %q does not contain %q", got, tc.want)
			}
		})
	}
}
