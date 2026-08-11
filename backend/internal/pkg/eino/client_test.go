package eino

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"candypro/api/internal/config"
	einotool "candypro/api/internal/pkg/eino/tool"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// chdirToBackend moves the test process into the backend directory (the one
// containing internal/pkg/eino/corpus) so rag.resolveCorpusDir() can resolve
// the bundled corpus from the relative candidates it checks. go test runs the
// test binary with CWD = the package source dir, so without this the corpus
// lookup would fail even though the corpus is present in the repo. The
// original working directory is restored via t.Cleanup.
func chdirToBackend(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}
	dir := wd
	for {
		if info, statErr := os.Stat(filepath.Join(dir, "internal", "pkg", "eino", "corpus")); statErr == nil && info.IsDir() {
			if chdirErr := os.Chdir(dir); chdirErr != nil {
				t.Fatalf("os.Chdir(%q) error = %v", dir, chdirErr)
			}
			t.Cleanup(func() { _ = os.Chdir(wd) })
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("could not locate backend dir (internal/pkg/eino/corpus) from %q", wd)
		}
		dir = parent
	}
}

// TestClientRAGComplianceEnabled verifies that NewClient succeeds when RAG
// compliance is enabled (openai.NewChatModel only constructs, no network),
// builds the ComplianceRetriever from the bundled corpus, and flips the
// package-level RAGComplianceEnabled flag to true.
func TestClientRAGComplianceEnabled(t *testing.T) {
	chdirToBackend(t)

	cfg := config.AIConfig{
		OpenAIAPIKey:         "sk-test",
		OpenAIBaseURL:        "http://127.0.0.1:1",
		OpenAIModel:          "gpt-4o",
		RetryMaxAttempts:     1,
		RetryIntervalSec:     0,
		HTTPTimeout:          5 * time.Second,
		TranslateRetryMax:    1,
		RAGComplianceEnabled: true,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if !client.IsEnabled() {
		t.Fatalf("expected client IsEnabled()=true, got false")
	}
	if !RAGComplianceEnabled {
		t.Fatalf("expected package flag RAGComplianceEnabled=true after NewClient, got false")
	}
	if client.ComplianceRetriever() == nil {
		t.Fatalf("expected non-nil ComplianceRetriever with RAGComplianceEnabled=true")
	}
}

// TestClientRAGComplianceDisabled verifies that NewClient leaves the package
// flag false and the retriever nil when RAG compliance is disabled.
func TestClientRAGComplianceDisabled(t *testing.T) {
	cfg := config.AIConfig{
		OpenAIAPIKey:         "sk-test",
		OpenAIBaseURL:        "http://127.0.0.1:1",
		OpenAIModel:          "gpt-4o",
		RetryMaxAttempts:     1,
		RetryIntervalSec:     0,
		HTTPTimeout:          5 * time.Second,
		TranslateRetryMax:    1,
		RAGComplianceEnabled: false,
	}

	client, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	if RAGComplianceEnabled {
		t.Fatalf("expected package flag RAGComplianceEnabled=false after NewClient, got true")
	}
	if client.ComplianceRetriever() != nil {
		t.Fatalf("expected nil ComplianceRetriever with RAGComplianceEnabled=false, got %v", client.ComplianceRetriever())
	}
}

// TestAccumulateMessageOutput_SkipsToolRole is the regression for G29(b):
// raw tool-role content must never be concatenated into the user-facing answer.
func TestAccumulateMessageOutput_SkipsToolRole(t *testing.T) {
	// A plain tool-role message carries raw tool output and must be dropped.
	got, err := accumulateMessageOutput("", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Tool, Content: `{"status":"ok","internal":"secret"}`},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Fatalf("tool-role content leaked into the answer: %q", got)
	}

	// Assistant text is appended as before.
	got, err = accumulateMessageOutput("", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Assistant, Content: "Your shipment is on the way."},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Your shipment is on the way." {
		t.Fatalf("expected assistant text, got %q", got)
	}

	// An assistant message that only carries tool calls has empty content and
	// must not contribute anything.
	got, err = accumulateMessageOutput("prefix ", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Assistant, Content: ""},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "prefix " {
		t.Fatalf("empty assistant message changed output: %q", got)
	}
}

// TestAccumulateMessageOutput_StreamFiltersToolChunks covers the streaming
// branch: tool-result chunks are filtered, assistant chunks are concatenated.
func TestAccumulateMessageOutput_StreamFiltersToolChunks(t *testing.T) {
	stream := schema.StreamReaderFromArray([]*schema.Message{
		{Role: schema.Tool, Content: `{"raw":"result-1"}`},
		{Role: schema.Assistant, Content: "Your shipment is "},
		{Role: schema.Tool, Content: `{"raw":"result-2"}`},
		{Role: schema.Assistant, Content: "arriving today."},
	})
	got, err := accumulateMessageOutput("", &adk.MessageVariant{MessageStream: stream})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Your shipment is arriving today." {
		t.Fatalf("stream concatenation leaked tool chunks: %q", got)
	}
}

// TestAccumulateMessageOutput_NilVariant is a no-op guard.
func TestAccumulateMessageOutput_NilVariant(t *testing.T) {
	got, err := accumulateMessageOutput("keep", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "keep" {
		t.Fatalf("nil variant changed output: %q", got)
	}
}

// TestAppendTerminalToolResult is the regression for the G29(b) refutation: a
// ReturnDirectly tool (e.g. submit_quotation_for_human_review) terminates the
// agent run with a Role==Tool message that carries the user-facing result.
// accumulateMessageOutput must still drop intermediate tool-role content, but
// the trailing result must survive so Generate does not come back empty.
func TestAppendTerminalToolResult(t *testing.T) {
	// Pure ReturnDirectly termination: no assistant text, so the terminal tool
	// result becomes the whole answer.
	got := appendTerminalToolResult("", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Tool, Content: `{"status":"pending_human_review","summary":"Queued quotation review"}`},
	})
	if !strings.Contains(got, "pending_human_review") || !strings.Contains(got, "Queued quotation review") {
		t.Fatalf("appendTerminalToolResult dropped the pure terminal result, got %q", got)
	}

	// Mixed turn: assistant text precedes the ReturnDirectly terminal result,
	// both must be present in the answer.
	got = appendTerminalToolResult("Submitting your quotation... ", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Tool, Content: "Queued for review"},
	})
	if got != "Submitting your quotation... Queued for review" {
		t.Fatalf("appendTerminalToolResult mishandled the mixed turn, got %q", got)
	}

	// Cosmetic separator (arguer edge case): when the accumulated assistant text
	// does NOT end in whitespace, the appended terminal result must be separated
	// so two sentences never merge into "...hint 5.0%)Queued quotation review".
	got = appendTerminalToolResult("Final discount hint 5.0%", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Tool, Content: "Queued for review"},
	})
	if got != "Final discount hint 5.0%\n\nQueued for review" {
		t.Fatalf("appendTerminalToolResult did not separate the merged text, got %q", got)
	}

	// Assistant text that already ends in a newline must not gain a double
	// separator.
	got = appendTerminalToolResult("Done.\n", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Tool, Content: "Queued for review"},
	})
	if got != "Done.\nQueued for review" {
		t.Fatalf("appendTerminalToolResult added a redundant separator, got %q", got)
	}

	// A trailing Tool-role STREAM variant (Message nil) is not the ReturnDirectly
	// terminal (non-streaming Generate emits plain messages) and must be ignored.
	got = appendTerminalToolResult("prefix", &adk.MessageVariant{
		MessageStream: schema.StreamReaderFromArray([]*schema.Message{
			{Role: schema.Tool, Content: "raw stream result"},
		}),
	})
	if got != "prefix" {
		t.Fatalf("appendTerminalToolResult surfaced a stream tool variant, got %q", got)
	}

	// Trailing assistant event (the normal termination) must not be re-appended:
	// its content is already accumulated by accumulateMessageOutput.
	got = appendTerminalToolResult("Your shipment is on the way.", &adk.MessageVariant{
		Message: &schema.Message{Role: schema.Assistant, Content: " More text"},
	})
	if got != "Your shipment is on the way." {
		t.Fatalf("appendTerminalToolResult re-appended a non-tool trailing event, got %q", got)
	}

	// Empty tool content and nil variant are no-ops.
	if got := appendTerminalToolResult("x", &adk.MessageVariant{Message: &schema.Message{Role: schema.Tool, Content: ""}}); got != "x" {
		t.Fatalf("empty tool content changed output: %q", got)
	}
	if got := appendTerminalToolResult("x", nil); got != "x" {
		t.Fatalf("nil trailing variant changed output: %q", got)
	}
}

// fakeReturnDirectlyModel is a minimal ToolCallingChatModel that answers every
// prompt with a single tool_call to submit_quotation_for_human_review, driving
// a ChatModelAgent to terminate through the ReturnDirectly path.
type fakeReturnDirectlyModel struct {
	toolCalls []schema.ToolCall
}

func (m *fakeReturnDirectlyModel) Generate(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	return &schema.Message{Role: schema.Assistant, ToolCalls: m.toolCalls}, nil
}

func (m *fakeReturnDirectlyModel) Stream(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, errors.New("fakeReturnDirectlyModel does not support streaming")
}

func (m *fakeReturnDirectlyModel) WithTools(_ []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return m, nil
}

// TestGenerate_ReturnDirectlyTerminalResultPreserved is the end-to-end G29(b)
// regression: the non-streaming Client.Generate path (used by the system
// Chatbot and GenerateQuotationText with the full agent attached) must not come
// back empty when the agent ends on the ReturnDirectly tool
// submit_quotation_for_human_review. The terminal Role==Tool message carries
// the user-facing review-queued result; this test pins that it survives.
func TestGenerate_ReturnDirectlyTerminalResultPreserved(t *testing.T) {
	ctx := context.Background()

	// nil saver: the tool reports pending_human_review without persisting.
	quoteTool, err := einotool.NewQuotationHumanReviewTool(ctx, nil)
	if err != nil {
		t.Fatalf("NewQuotationHumanReviewTool() error = %v", err)
	}

	fake := &fakeReturnDirectlyModel{toolCalls: []schema.ToolCall{{
		ID:   "call_1",
		Type: "function",
		Function: schema.FunctionCall{
			Name:      "submit_quotation_for_human_review",
			Arguments: `{"customer_ref":"REF1","currency":"USD","total_amount":123.45}`,
		},
	}}}

	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "TestReturnDirectlyAgent",
		Description: "test agent for G29(b) regression",
		Instruction: "You queue quotations for human review.",
		Model:       fake,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{quoteTool},
			},
			ReturnDirectly: map[string]bool{
				"submit_quotation_for_human_review": true,
			},
		},
	})
	if err != nil {
		t.Fatalf("NewChatModelAgent() error = %v", err)
	}

	c := &Client{agent: a}
	reply, err := c.Generate(ctx, "Please quote for REF1")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if !strings.Contains(reply, "pending_human_review") || !strings.Contains(reply, "Queued quotation review") {
		t.Fatalf("Generate() dropped the ReturnDirectly terminal result, got reply %q", reply)
	}
}
