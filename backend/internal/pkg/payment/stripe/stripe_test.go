package stripe

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// rewriteTransport rewrites every request to the fake API server, regardless of
// the Host in the request (the adapter always targets api.stripe.com).
type rewriteTransport struct {
	base *url.URL
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = t.base.Scheme
	req2.URL.Host = t.base.Host
	return http.DefaultTransport.RoundTrip(req2)
}

// TestStripeAmountScaling verifies the currency-aware amount conversion used for
// Authorization (G16): two-decimal currencies multiply by 100, zero-decimal
// currencies (JPY, KRW, VND, ...) pass the raw amount through.
func TestStripeAmountScaling(t *testing.T) {
	cases := []struct {
		amount   float64
		currency string
		want     string
	}{
		{100.00, "USD", "10000"},
		{99.99, "EUR", "9999"},
		{1234, "JPY", "1234"},
		{500, "KRW", "500"},
		{2500, "VND", "2500"},
	}
	for _, c := range cases {
		if got := stripeAmount(c.amount, c.currency); got != c.want {
			t.Fatalf("stripeAmount(%v,%s)=%s want %s", c.amount, c.currency, got, c.want)
		}
	}
}

// newStripeTestServer returns a fake Stripe API serving a PaymentIntent with the
// given currency and records every POST body (capture/refund) into *bodies.
func newStripeTestServer(t *testing.T, currency string) (*httptest.Server, *[]string) {
	t.Helper()
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/payment_intents/") {
			// intent lookup: Capture/Refund read currency from here
			_, _ = io.WriteString(w, `{"id":"pi_123","currency":"`+currency+`","amount":1234}`)
			return
		}
		if r.Method == http.MethodPost {
			b, _ := io.ReadAll(r.Body)
			bodies = append(bodies, string(b))
			_, _ = io.WriteString(w, `{"id":"pi_123","status":"succeeded"}`)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv, &bodies
}

func newStripeAdapter(t *testing.T, srv *httptest.Server) *Adapter {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}
	a := New("sk_test", "")
	a.httpClient = &http.Client{Transport: &rewriteTransport{base: u}}
	return a
}

// TestCaptureZeroDecimalCurrency is the core G16 regression test: a partial
// capture in a zero-decimal currency (JPY) must send the RAW amount (1234), not
// amount*100 (123400), to Stripe. On the pre-fix path the capture sent
// amount_to_capture=123400, which Stripe rejects (or over-captures).
func TestCaptureZeroDecimalCurrency(t *testing.T) {
	srv, bodies := newStripeTestServer(t, "jpy")
	a := newStripeAdapter(t, srv)

	if _, err := a.Capture(context.Background(), "pi_123", 1234); err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(*bodies) != 1 {
		t.Fatalf("expected 1 capture request, got %d", len(*bodies))
	}
	body := (*bodies)[0]
	if !strings.Contains(body, "amount_to_capture=1234") {
		t.Fatalf("zero-decimal capture must send raw amount 1234, body=%q", body)
	}
	if strings.Contains(body, "123400") {
		t.Fatalf("zero-decimal capture must NOT multiply by 100, body=%q", body)
	}
}

// TestCaptureTwoDecimalCurrency guards the non-zero-decimal path: a partial
// capture in USD must still send cents (100.00 -> 10000), i.e. no regression for
// the common case.
func TestCaptureTwoDecimalCurrency(t *testing.T) {
	srv, bodies := newStripeTestServer(t, "usd")
	a := newStripeAdapter(t, srv)

	if _, err := a.Capture(context.Background(), "pi_123", 100.00); err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(*bodies) != 1 {
		t.Fatalf("expected 1 capture request, got %d", len(*bodies))
	}
	if !strings.Contains((*bodies)[0], "amount_to_capture=10000") {
		t.Fatalf("two-decimal capture must send cents 10000, body=%q", (*bodies)[0])
	}
}

// TestRefundZeroDecimalCurrency is the G16 refund regression test: a partial
// refund in a zero-decimal currency must send the raw amount (500), not 50000.
func TestRefundZeroDecimalCurrency(t *testing.T) {
	srv, bodies := newStripeTestServer(t, "jpy")
	a := newStripeAdapter(t, srv)

	if _, err := a.Refund(context.Background(), "pi_123", 500); err != nil {
		t.Fatalf("refund: %v", err)
	}
	if len(*bodies) != 1 {
		t.Fatalf("expected 1 refund request, got %d", len(*bodies))
	}
	body := (*bodies)[0]
	if !strings.Contains(body, "amount=500") {
		t.Fatalf("zero-decimal refund must send raw amount 500, body=%q", body)
	}
	if strings.Contains(body, "50000") {
		t.Fatalf("zero-decimal refund must NOT multiply by 100, body=%q", body)
	}
}

// TestRefundTwoDecimalCurrency guards the non-zero-decimal refund path.
func TestRefundTwoDecimalCurrency(t *testing.T) {
	srv, bodies := newStripeTestServer(t, "usd")
	a := newStripeAdapter(t, srv)

	if _, err := a.Refund(context.Background(), "pi_123", 50.00); err != nil {
		t.Fatalf("refund: %v", err)
	}
	if len(*bodies) != 1 {
		t.Fatalf("expected 1 refund request, got %d", len(*bodies))
	}
	if !strings.Contains((*bodies)[0], "amount=5000") {
		t.Fatalf("two-decimal refund must send cents 5000, body=%q", (*bodies)[0])
	}
}

// TestEventID verifies the webhook event id is parsed and exposed so consumers
// can dedupe replays (G16). Previously the id was dropped by the parser.
func TestEventID(t *testing.T) {
	payload := []byte(`{"id":"evt_1J2A3B4C5D6E","type":"payment_intent.succeeded","data":{"object":{"id":"pi_1","currency":"usd"}}}`)
	if got := EventID(payload); got != "evt_1J2A3B4C5D6E" {
		t.Fatalf("EventID = %q, want evt_1J2A3B4C5D6E", got)
	}
	if got := EventID([]byte(`{"type":"payment_intent.payment_failed"}`)); got != "" {
		t.Fatalf("EventID = %q, want empty when no id present", got)
	}
	if got := EventID([]byte(`not-json`)); got != "" {
		t.Fatalf("EventID = %q, want empty for unparseable payload", got)
	}
}
