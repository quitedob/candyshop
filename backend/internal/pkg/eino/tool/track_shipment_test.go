package tool

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestTrackShipmentTool_StubWithoutTracker(t *testing.T) {
	t.Setenv("TRACK_SHIPMENT_17TRACK_API_KEY", "")
	tl, err := NewTrackShipmentTool(context.Background(), nil)
	if err != nil {
		t.Fatalf("NewTrackShipmentTool: %v", err)
	}
	out := innerInvoke(t, tl, `{"bl_number":"MSKU1234567","carrier_code":"msku"}`)
	var resp TrackShipmentResponse
	if err := json.Unmarshal([]byte(out), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !strings.Contains(resp.LastEvent, "No carrier event data available") {
		t.Fatalf("expected stub last event, got %q", resp.LastEvent)
	}
	if resp.CurrentLocation != "UNAVAILABLE" {
		t.Fatalf("expected UNAVAILABLE location, got %q", resp.CurrentLocation)
	}
}

func TestTrackShipmentTool_InvalidInput(t *testing.T) {
	t.Setenv("TRACK_SHIPMENT_17TRACK_API_KEY", "")
	tl, err := NewTrackShipmentTool(context.Background(), nil)
	if err != nil {
		t.Fatalf("NewTrackShipmentTool: %v", err)
	}
	et, ok := tl.(*InvokableReviewEditTool)
	if !ok {
		t.Fatal("expected *InvokableReviewEditTool")
	}
	if _, err := et.InvokableTool.InvokableRun(context.Background(), `{"bl_number":"","carrier_code":""}`); err == nil {
		t.Fatal("expected error for missing bl_number/carrier_code")
	}
}
