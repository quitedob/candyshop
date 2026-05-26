package shipmenttrack

import (
	"testing"
	"time"
)

func TestResolveTrackingStatus(t *testing.T) {
	eta := time.Now().UTC().Add(48 * time.Hour)
	past := time.Now().UTC().Add(-48 * time.Hour)

	cases := []struct {
		event string
		eta   *time.Time
		want  string
	}{
		{"Package delivered to consignee", nil, "DELIVERED"},
		{"customs hold at port", nil, "DELAYED"},
		{"in transit to destination", &eta, "IN_TRANSIT"},
		{"", &past, "PAST_ETA"},
		{"", nil, "TRACKING_PENDING"},
	}
	for _, tc := range cases {
		if got := ResolveTrackingStatus(tc.event, tc.eta); got != tc.want {
			t.Fatalf("ResolveTrackingStatus(%q) = %q, want %q", tc.event, got, tc.want)
		}
	}
}

func TestMapToDBStatus(t *testing.T) {
	if got := MapToDBStatus("IN_TRANSIT", "DISPATCHED"); got != "IN_TRANSIT" {
		t.Fatalf("expected IN_TRANSIT, got %q", got)
	}
	if got := MapToDBStatus("DELIVERED", "IN_TRANSIT"); got != "DELIVERED" {
		t.Fatalf("expected DELIVERED, got %q", got)
	}
	if got := MapToDBStatus("IN_TRANSIT", "IN_TRANSIT"); got != "" {
		t.Fatalf("expected no change, got %q", got)
	}
}
