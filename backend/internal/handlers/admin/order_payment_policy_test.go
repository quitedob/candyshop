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

func TestIsPaidOrPartial(t *testing.T) {
	if !isPaidOrPartial("paid") || !isPaidOrPartial("partial") {
		t.Fatal("expected paid/partial to pass")
	}
	if isPaidOrPartial("unpaid") || isPaidOrPartial("refunded") {
		t.Fatal("expected unpaid/refunded to fail")
	}
}

func TestPaymentInsufficientForExecution_NETTermsStillBlocked(t *testing.T) {
	if !paymentInsufficientForExecution("production", "unpaid") {
		t.Fatal("unpaid orders must not enter production even with NET terms")
	}
	if paymentInsufficientForExecution("production", "partial") {
		t.Fatal("partial payment should allow production")
	}
	if paymentInsufficientForExecution("pending", "unpaid") {
		t.Fatal("pending status should not require payment")
	}
}
