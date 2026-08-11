package tool

import (
	"context"
	"testing"

	einoct "github.com/cloudwego/eino/components/tool"
)

// invokeTool runs an Eino tool with the given arguments JSON and returns the raw
// string result, failing the test on error. Eino's BaseTool only exposes Info;
// invocation requires the InvokableTool interface.
func invokeTool(t *testing.T, bt einoct.BaseTool, args string) string {
	t.Helper()
	inv, ok := bt.(einoct.InvokableTool)
	if !ok {
		t.Fatal("tool is not invokable")
	}
	out, err := inv.InvokableRun(context.Background(), args)
	if err != nil {
		t.Fatalf("InvokableRun(%s): %v", args, err)
	}
	return out
}

// innerInvoke unwraps an InvokableReviewEditTool (which enforces a HITL
// review-and-edit interrupt) and runs the underlying invokable tool directly, so
// tests exercise the real handler without driving the review state machine.
func innerInvoke(t *testing.T, bt einoct.BaseTool, args string) string {
	t.Helper()
	if et, ok := bt.(*InvokableReviewEditTool); ok {
		return invokeTool(t, et.InvokableTool, args)
	}
	return invokeTool(t, bt, args)
}
