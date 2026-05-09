package admin

import "testing"

func TestRequiresFullPrepaymentCountry(t *testing.T) {
	// Test with nil policy service (fallback hardcoded logic)
	h := &Handler{}
	cases := []struct {
		country string
		want    bool
	}{
		{country: "India", want: true},
		{country: "in", want: true},
		{country: "Pakistan", want: true},
		{country: "pk", want: true},
		{country: "USA", want: false},
	}

	for _, tc := range cases {
		got := h.requiresFullPrepaymentCountry(tc.country)
		if got != tc.want {
			t.Fatalf("requiresFullPrepaymentCountry(%q)=%v want=%v", tc.country, got, tc.want)
		}
	}
}

func TestRequiresPaidBeforeExecution(t *testing.T) {
	if !requiresPaidBeforeExecution("confirmed") {
		t.Fatalf("expected confirmed to require paid status")
	}
	if !requiresPaidBeforeExecution("production") {
		t.Fatalf("expected production to require paid status")
	}
	if requiresPaidBeforeExecution("pending") {
		t.Fatalf("expected pending to not require paid status")
	}
	if requiresPaidBeforeExecution("cancelled") {
		t.Fatalf("expected cancelled to not require paid status")
	}
}
