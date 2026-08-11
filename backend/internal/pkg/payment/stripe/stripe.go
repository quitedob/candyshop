package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
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
//
// amount is in the gateway-facing minor unit (dollars for USD, yen for JPY), so
// the raw value passed to Stripe must be scaled currency-aware: x100 for
// two-decimal currencies, raw for zero-decimal currencies (JPY, KRW, ...). The
// currency is not part of the PaymentGateway interface, so it is read from the
// PaymentIntent itself rather than guessed (G16). A zero amount captures the
// full authorized amount and needs no scaling.
func (a *Adapter) Capture(ctx context.Context, transactionID string, amount float64) (*payment.GatewayResponse, error) {
	params := map[string]string{}
	if amount > 0 {
		cur, err := a.intentCurrency(ctx, transactionID)
		if err != nil {
			return nil, err
		}
		params["amount_to_capture"] = stripeAmount(amount, cur)
	}
	return a.postForm(ctx, "/payment_intents/"+transactionID+"/capture", params, "")
}

// Refund refunds a payment (full or partial).
//
// Same currency-aware scaling as Capture: the PaymentIntent's currency decides
// whether amount is in cents (x100) or the raw minor unit (zero-decimal). A zero
// amount refunds the full payment and needs no scaling.
func (a *Adapter) Refund(ctx context.Context, transactionID string, amount float64) (*payment.GatewayResponse, error) {
	params := map[string]string{
		"payment_intent": transactionID,
	}
	if amount > 0 {
		cur, err := a.intentCurrency(ctx, transactionID)
		if err != nil {
			return nil, err
		}
		params["amount"] = stripeAmount(amount, cur)
	}
	return a.postForm(ctx, "/refunds", params, "")
}

// Void cancels an authorized but uncaptured PaymentIntent.
func (a *Adapter) Void(ctx context.Context, transactionID string) (*payment.GatewayResponse, error) {
	return a.postForm(ctx, "/payment_intents/"+transactionID+"/cancel", nil, "")
}

// get sends a GET request to the Stripe API and returns the raw response body.
func (a *Adapter) get(ctx context.Context, path string) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("stripe: build request: %w", err)
	}
	httpReq.SetBasicAuth(a.secretKey, "")
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
	return respBody, nil
}

// intentCurrency fetches the PaymentIntent to learn its currency. Capture/Refund
// need it to scale partial amounts correctly, because the PaymentGateway
// interface does not carry a currency (G16). Reading it from the intent is
// authoritative and works across processes/restarts, unlike caching it from a
// prior Authorize call.
func (a *Adapter) intentCurrency(ctx context.Context, transactionID string) (string, error) {
	body, err := a.get(ctx, "/payment_intents/"+transactionID)
	if err != nil {
		return "", err
	}
	var pi struct {
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(body, &pi); err != nil {
		return "", fmt.Errorf("stripe: parse payment intent %s: %w", transactionID, err)
	}
	if pi.Currency == "" {
		return "", fmt.Errorf("stripe: payment intent %s has no currency", transactionID)
	}
	return pi.Currency, nil
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
		ID           string `json:"id"`
		Status       string `json:"status"`
		ClientSecret string `json:"client_secret"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("stripe: parse response: %w", err)
	}

	return &payment.GatewayResponse{
		TransactionID: result.ID,
		Status:        result.Status,
		RawResponse:   string(respBody),
		ClientSecret:  result.ClientSecret,
	}, nil
}

// stripeAmount converts a float amount to Stripe's smallest-currency-unit integer string.
// For most currencies this is cents (×100). For zero-decimal currencies (JPY, KRW, etc.),
// amount is already in the smallest unit.
//
// L6: math.Round (round-half-away-from-zero) instead of Sprintf("%.0f", ...),
// which could round a fractional-cent partial down a cent.
func stripeAmount(amount float64, currency string) string {
	cur := strings.ToUpper(currency)
	if zeroDecimal[cur] {
		return fmt.Sprintf("%.0f", math.Round(amount))
	}
	return fmt.Sprintf("%.0f", math.Round(amount*100))
}

var zeroDecimal = map[string]bool{
	"BIF": true, "CLP": true, "DJF": true, "GNF": true,
	"JPY": true, "KMF": true, "KRW": true, "MGA": true,
	"PYG": true, "RWF": true, "UGX": true, "VND": true,
	"VUV": true, "XAF": true, "XOF": true, "XPF": true,
}

// ValidateWebhookPayload validates a Stripe webhook payload against its signature header.
// The signature header format is: t=<timestamp>,v1=<signature>[,v1=<signature>...]
// Signature is computed as HMAC-SHA256(webhookSecret, timestamp + "." + payload).
func (a *Adapter) ValidateWebhookPayload(payload []byte, signatureHeader string) ([]byte, error) {
	if signatureHeader == "" {
		return nil, fmt.Errorf("missing Stripe-Signature header")
	}
	if a.webhookSecret == "" {
		return nil, fmt.Errorf("webhook secret not configured")
	}

	// Parse signature header: "t=1492774577,v1=abc123,v1=def456"
	var timestamp string
	var signatures []string
	for _, part := range strings.Split(signatureHeader, ",") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, "t=") {
			timestamp = strings.TrimPrefix(part, "t=")
		} else if strings.HasPrefix(part, "v1=") {
			signatures = append(signatures, strings.TrimPrefix(part, "v1="))
		}
	}

	if timestamp == "" || len(signatures) == 0 {
		return nil, fmt.Errorf("invalid signature header format")
	}

	// Verify timestamp is within tolerance (±5 minutes)
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp in signature header")
	}
	if diff := time.Now().Unix() - ts; diff > 300 || diff < -300 {
		return nil, fmt.Errorf("webhook timestamp outside tolerance window (diff=%ds)", diff)
	}

	// Compute expected signature
	signedPayload := fmt.Sprintf("%s.%s", timestamp, string(payload))
	mac := hmac.New(sha256.New, []byte(a.webhookSecret))
	mac.Write([]byte(signedPayload))
	expected := hex.EncodeToString(mac.Sum(nil))

	// Compare against each provided signature (constant-time comparison)
	for _, sig := range signatures {
		if hmac.Equal([]byte(sig), []byte(expected)) {
			return payload, nil
		}
	}

	return nil, fmt.Errorf("no matching signature found")
}

// stripeEvent represents a minimal Stripe webhook event. ID is the Stripe event
// id ("evt_...") that lets webhook consumers dedupe replayed deliveries (G16);
// it was previously dropped by ParseWebhookEvent, so replays were indistinguishable.
type stripeEvent struct {
	ID   string          `json:"id"`
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

// EventID extracts the Stripe event id ("evt_...") from a webhook payload so
// consumers can skip replayed deliveries. Returns "" for unparseable payloads
// or events without an id. Combined with the payment status state machine (which
// rejects confirm/authorize/fail transitions on already-processed records), this
// makes webhook replays idempotent (G16).
func EventID(payload []byte) string {
	var event stripeEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return ""
	}
	return event.ID
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
