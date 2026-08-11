package tool

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cloudwego/eino/components/tool/utils"
)

// TestInvokableReviewEditTool_ReadOnlyCompletesWithoutReview is the regression
// for G29(a): the HITL wrapper must not interrupt read-only informational tools
// (track_shipment, translate_content, validate_lc_documents). Because the SSE
// flow has no resume path, an unconditional interrupt would mean those tools can
// never complete. Invoking through the wrapper must return the real result.
func TestInvokableReviewEditTool_ReadOnlyCompletesWithoutReview(t *testing.T) {
	t.Setenv("TRACK_SHIPMENT_17TRACK_API_KEY", "")
	bt, err := NewTrackShipmentTool(context.Background(), nil)
	if err != nil {
		t.Fatalf("NewTrackShipmentTool: %v", err)
	}
	et, ok := bt.(*InvokableReviewEditTool)
	if !ok {
		t.Fatalf("expected *InvokableReviewEditTool, got %T", bt)
	}

	out, err := et.InvokableRun(context.Background(), `{"bl_number":"MSKU1234567","carrier_code":"msku"}`)
	if err != nil {
		t.Fatalf("track_shipment must complete through the wrapper: %v", err)
	}
	var resp TrackShipmentResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("expected a tracking JSON result, got %q", out)
	}
	if resp.CurrentLocation == "" {
		t.Fatalf("expected a populated tracking result, got %q", out)
	}
}

// TestInvokableReviewEditTool_TranslateCompletesWithoutReview ensures the same
// pass-through for translate_content, which is another read-only tool that the
// default review policy exempts.
func TestInvokableReviewEditTool_TranslateCompletesWithoutReview(t *testing.T) {
	tl, err := NewTranslateContentTool(context.Background(), func(_ context.Context, sourceData map[string]string, targetLocales []string) (map[string]map[string]string, error) {
		out := map[string]map[string]string{}
		for _, loc := range targetLocales {
			out[loc] = map[string]string{}
			for k, v := range sourceData {
				out[loc][k] = v + "@" + loc
			}
		}
		return out, nil
	})
	if err != nil {
		t.Fatalf("NewTranslateContentTool: %v", err)
	}
	et, ok := tl.(*InvokableReviewEditTool)
	if !ok {
		t.Fatalf("expected *InvokableReviewEditTool, got %T", tl)
	}

	out, err := et.InvokableRun(context.Background(), `{"fields":{"name":"Candy"},"target_locales":["ar","ja"]}`)
	if err != nil {
		t.Fatalf("translate_content must complete through the wrapper: %v", err)
	}
	var resp TranslateContentResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("expected a translation JSON result, got %q", out)
	}
	if resp.LocalesDone != 2 {
		t.Fatalf("expected 2 locales, got %d", resp.LocalesDone)
	}
}

// TestInvokableReviewEditTool_ReviewRequiredStillInterrupts ensures a tool that
// is NOT exempt from the review gate still interrupts, preserving the HITL flow
// for future state-changing tools.
func TestInvokableReviewEditTool_ReviewRequiredStillInterrupts(t *testing.T) {
	base, err := utils.InferTool("update_draft_document",
		"update a draft document after human approval",
		func(ctx context.Context, req struct {
			DraftID string `json:"draft_id"`
		}) (string, error) {
			return "updated " + req.DraftID, nil
		})
	if err != nil {
		t.Fatalf("InferTool: %v", err)
	}
	et := &InvokableReviewEditTool{InvokableTool: base}

	out, err := et.InvokableRun(context.Background(), `{"draft_id":"d1"}`)
	if err == nil {
		t.Fatalf("expected an interrupt for a review-required tool, got out=%q", out)
	}
	if out == "updated d1" {
		t.Fatalf("review-required tool must not execute before review, got %q", out)
	}
}

// TestInvokableReviewEditTool_CustomPolicyForcesReview verifies the configurable
// RequiresReview policy can force a normally-exempt tool through the gate.
func TestInvokableReviewEditTool_CustomPolicyForcesReview(t *testing.T) {
	t.Setenv("TRACK_SHIPMENT_17TRACK_API_KEY", "")
	bt, err := NewTrackShipmentTool(context.Background(), nil)
	if err != nil {
		t.Fatalf("NewTrackShipmentTool: %v", err)
	}
	et, ok := bt.(*InvokableReviewEditTool)
	if !ok {
		t.Fatalf("expected *InvokableReviewEditTool, got %T", bt)
	}
	et.RequiresReview = func(toolName, _ string) bool { return true }

	if _, err := et.InvokableRun(context.Background(), `{"bl_number":"MSKU1234567","carrier_code":"msku"}`); err == nil {
		t.Fatal("expected an interrupt when a custom policy forces review")
	}
}
