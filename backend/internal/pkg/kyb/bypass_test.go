package kyb

import "testing"

func TestPendingOrderAllowed_SampleTier(t *testing.T) {
	cfg := TierConfig{
		BypassMaxOrderUSD:       100,
		BypassSampleMaxOrderUSD: 500,
		SampleProductIDs:        []string{"sample-a"},
	}
	if !PendingOrderAllowed(cfg, 400, []string{"sample-a"}) {
		t.Fatalf("sample-only order should allow under sample cap")
	}
	if PendingOrderAllowed(cfg, 600, []string{"sample-a"}) {
		t.Fatalf("sample-only order over both caps should deny")
	}
	if !PendingOrderAllowed(cfg, 80, []string{"other"}) {
		t.Fatalf("non-sample should allow under general cap")
	}
	if PendingOrderAllowed(cfg, 150, []string{"other"}) {
		t.Fatalf("non-sample over general cap should deny")
	}
}
