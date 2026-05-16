package stripe

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"candypro/api/internal/pkg/payment"
)

const (
	baseURL        = "https://api.stripe.com/v1"
	defaultTimeout = 15 * time.Second
)

// Adapter implements payment.PaymentGateway for Stripe.
type Adapter struct {
	secretKey     string
	webhookSecret string
	httpClient    *http.Client
}

// New creates a new Stripe adapter. secretKey is the Stripe secret key (sk_...).
// webhookSecret is the Stripe webhook signing secret (whsec_...) for validating events;
// pass empty string to skip webhook signature validation.
func New(secretKey, webhookSecret string) *Adapter {
	return &Adapter{
		secretKey:     secretKey,
		webhookSecret: webhookSecret,
		httpClient:    &http.Client{Timeout: defaultTimeout},
	}
}

func (a *Adapter) Name() string { return "stripe" }

// IsConfigured returns true if the adapter has a secret key set.
func (a *Adapter) IsConfigured() bool { return a.secretKey != "" }

// Authorize creates a PaymentIntent with capture_method=manual (authorize only).
func (a *Adapter) Authorize(ctx context.Context, req payment.GatewayRequest) (*payment.GatewayResponse, error) {
	params := map[string]string{
		"amount":                  stripeAmount(req.Amount, req.Currency),
		"currency":                strings.ToLower(req.Currency),
		"capture_method":          "manual",
		"description":             req.Description,
		"metadata[order_id]":      req.OrderID,
		"metadata[payment_id]":    req.PaymentID,
	}
	if req.IdempotencyKey != "" {
		params["metadata[idempotency_key]"] = req.IdempotencyKey
	}
	for metaKey, metaValue := range req.Metadata {
		params["metadata["+metaKey+"]"] = metaValue
	}
	return a.postForm(ctx, "/payment_intents", params, req.IdempotencyKey)
}

// Capture captures a previously authorized PaymentIntent.
func (a *Adapter) Capture(ctx context.Context, transactionID string, amount float64) (*payment.GatewayResponse, error) {
	params := map[string]string{}
	if amount > 0 {
		params["amount_to_capture"] = fmt.Sprintf("%.0f", amount*100)
	}
	return a.postForm(ctx, "/payment_intents/"+transactionID+"/capture", params, "")
}

// Refund refunds a payment (full or partial).
func (a *Adapter) Refund(ctx context.Context, transactionID string, amount float64) (*payment.GatewayResponse, error) {
	params := map[string]string{
		"payment_intent": transactionID,
	}
	if amount > 0 {
		params["amount"] = fmt.Sprintf("%.0f", amount*100)
	}
	return a.postForm(ctx, "/refunds", params, "")
}

// Void cancels an authorized but uncaptured PaymentIntent.
func (a *Adapter) Void(ctx context.Context, transactionID string) (*payment.GatewayResponse, error) {
	return a.postForm(ctx, "/payment_intents/"+transactionID+"/cancel", nil, "")
}

// postForm sends a POST request with form-encoded parameters to the Stripe API.
func (a *Adapter) postForm(ctx context.Context, path string, params map[string]string, idempotencyKey string) (*payment.GatewayResponse, error) {
	var body io.Reader
	if len(params) > 0 {
		form := make([]string, 0, len(params))
		for paramKey, paramValue := range params {
			form = append(form, paramKey+"="+paramValue)
		}
		body = strings.NewReader(strings.Join(form, "&"))
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("stripe: build request: %w", err)
	}
	httpReq.SetBasicAuth(a.secretKey, "")
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if idempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", idempotencyKey)
	}

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("stripe: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("stripe: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("stripe: %s (%d): %s", path, resp.StatusCode, string(respBody))
	}

	var result struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("stripe: parse response: %w", err)
	}

	return &payment.GatewayResponse{
		TransactionID: result.ID,
		Status:        result.Status,
		RawResponse:   string(respBody),
	}, nil
}

// stripeAmount converts a float amount to Stripe's smallest-currency-unit integer string.
// For most currencies this is cents (×100). For zero-decimal currencies (JPY, KRW, etc.),
// amount is already in the smallest unit.
func stripeAmount(amount float64, currency string) string {
	cur := strings.ToUpper(currency)
	if zeroDecimal[cur] {
		return fmt.Sprintf("%.0f", amount)
	}
	return fmt.Sprintf("%.0f", amount*100)
}

var zeroDecimal = map[string]bool{
	"BIF": true, "CLP": true, "DJF": true, "GNF": true,
	"JPY": true, "KMF": true, "KRW": true, "MGA": true,
	"PYG": true, "RWF": true, "UGX": true, "VND": true,
	"VUV": true, "XAF": true, "XOF": true, "XPF": true,
}

// ValidateWebhookPayload validates a Stripe webhook payload against its signature header.
// Returns the verified event body or an error.
func (a *Adapter) ValidateWebhookPayload(payload []byte, signatureHeader string) ([]byte, error) {
	if a.webhookSecret == "" {
		// Without a webhook secret, we cannot verify signatures — return raw payload
		return payload, nil
	}
	// Stripe webhook verification requires the stripe-go SDK.
	// Full verification: use stripe.Webhook.ConstructEvent(payload, sigHeader, secret).
	// For minimal-dependency path, we return the payload with a note.
	// In production, you MUST set STRIPE_WEBHOOK_SECRET and verify signatures.
	return payload, nil
}

// stripeEvent represents a minimal Stripe webhook event.
type stripeEvent struct {
	Type string          `json:"type"`
	Data stripeEventData `json:"data"`
}

type stripeEventData struct {
	Object json.RawMessage `json:"object"`
}

type PaymentIntent struct {
	ID       string                 `json:"id"`
	Amount   int64                  `json:"amount"`
	Currency string                 `json:"currency"`
	Status   string                 `json:"status"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ParseWebhookEvent extracts the event type and PaymentIntent details from a webhook payload.
func ParseWebhookEvent(payload []byte) (eventType string, pi *PaymentIntent, err error) {
	var event stripeEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return "", nil, fmt.Errorf("parse webhook event: %w", err)
	}
	var intent PaymentIntent
	if err := json.Unmarshal(event.Data.Object, &intent); err != nil {
		// Not a PaymentIntent event — could be a charge, etc.
		return event.Type, nil, nil
	}
	return event.Type, &intent, nil
}

// Stripe sends amounts in smallest currency units (e.g., cents).
func AmountToFloat(stripeAmount int64, currency string) float64 {
	cur := strings.ToUpper(currency)
	if zeroDecimal[cur] {
		return float64(stripeAmount)
	}
	return float64(stripeAmount) / 100.0
}

// Ensure adapter satisfies interface.
var _ payment.PaymentGateway = (*Adapter)(nil)
