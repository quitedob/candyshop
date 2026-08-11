package paypal

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"candypro/api/internal/pkg/payment"
)

const (
	sandboxBase    = "https://api-m.sandbox.paypal.com"
	productionBase = "https://api-m.paypal.com"
)

// Adapter implements payment.PaymentGateway for PayPal Checkout Orders v2.
type Adapter struct {
	clientID     string
	clientSecret string
	webhookID    string
	sandbox      bool
	httpClient   *http.Client

	// certFetch fetches the PayPal signing certificate for a cert URL. It is
	// defaulted to the real HTTP fetcher in New and may be overridden in tests
	// to avoid the SSRF-restricted host check forcing external network calls.
	certFetch func(ctx context.Context, certURL string) (*x509.Certificate, error)

	tokenMu     sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

// New creates a PayPal adapter. sandbox=true uses PayPal sandbox API.
func New(clientID, clientSecret, webhookID string, sandbox bool) *Adapter {
	a := &Adapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		webhookID:    webhookID,
		sandbox:      sandbox,
		httpClient:   &http.Client{Timeout: 20 * time.Second},
	}
	a.certFetch = a.fetchCertWithContext
	return a
}

func (a *Adapter) Name() string { return "paypal" }

// IsConfigured returns true when credentials are present.
func (a *Adapter) IsConfigured() bool {
	return a.clientID != "" && a.clientSecret != ""
}

func (a *Adapter) baseURL() string {
	if a.sandbox {
		return sandboxBase
	}
	return productionBase
}

// Authorize creates a PayPal checkout order (intent=CAPTURE).
func (a *Adapter) Authorize(ctx context.Context, req payment.GatewayRequest) (*payment.GatewayResponse, error) {
	body := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": req.PaymentID,
				"description":  req.Description,
				"custom_id":    req.OrderID,
				"amount": map[string]string{
					"currency_code": strings.ToUpper(req.Currency),
					"value":         formatAmount(req.Amount),
				},
			},
		},
		"application_context": map[string]string{
			"return_url": req.Metadata["return_url"],
			"cancel_url": req.Metadata["cancel_url"],
		},
	}
	respBody, err := a.postJSON(ctx, "/v2/checkout/orders", body)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Rel  string `json:"rel"`
			Href string `json:"href"`
		} `json:"links"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("paypal: parse order: %w", err)
	}
	approvalURL := ""
	for _, link := range parsed.Links {
		if link.Rel == "approve" || link.Rel == "payer-action" {
			approvalURL = link.Href
			break
		}
	}
	return &payment.GatewayResponse{
		TransactionID: parsed.ID,
		Status:        parsed.Status,
		RawResponse:   string(respBody),
		ApprovalURL:   approvalURL,
	}, nil
}

// Capture captures an approved PayPal order.
//
// The Orders v2 capture endpoint (`POST /v2/checkout/orders/{id}/capture`) does
// NOT accept an amount in the request body — it captures the full order amount
// fixed at order creation (verified against the Orders v2 spec). So the adapter
// cannot "send" the requested amount; instead it verifies that the capture the
// API actually performed matches the requested amount and fails otherwise (G16).
// This catches the case where the local payment record diverges from the PayPal
// order amount and money is taken for a different value than the one recorded.
func (a *Adapter) Capture(ctx context.Context, transactionID string, amount float64) (*payment.GatewayResponse, error) {
	// Send an explicit empty object: PayPal rejects null/empty bodies with
	// UNSUPPORTED_MEDIA_TYPE / INVALID_REQUEST, and `{}` is the documented form.
	respBody, err := a.postJSON(ctx, "/v2/checkout/orders/"+transactionID+"/capture", map[string]interface{}{})
	if err != nil {
		return nil, err
	}
	var parsed struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		PurchaseUnits []struct {
			Payments struct {
				Captures []struct {
					ID     string `json:"id"`
					Status string `json:"status"`
					Amount struct {
						CurrencyCode string `json:"currency_code"`
						Value        string `json:"value"`
					} `json:"amount"`
				} `json:"captures"`
			} `json:"payments"`
		} `json:"purchase_units"`
	}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("paypal: parse capture: %w", err)
	}
	if amount > 0 {
		if len(parsed.PurchaseUnits) == 0 || len(parsed.PurchaseUnits[0].Payments.Captures) == 0 {
			return nil, fmt.Errorf("paypal: capture response for order %s is missing capture details", transactionID)
		}
		captured := parsed.PurchaseUnits[0].Payments.Captures[0].Amount.Value
		if !amountsEqual(captured, amount) {
			return nil, fmt.Errorf("paypal: captured amount %s does not match requested %s (order %s)", captured, formatAmount(amount), transactionID)
		}
	}
	return &payment.GatewayResponse{
		TransactionID: parsed.ID,
		Status:        parsed.Status,
		RawResponse:   string(respBody),
	}, nil
}

// Refund refunds a captured PayPal payment (transactionID = capture ID).
func (a *Adapter) Refund(ctx context.Context, transactionID string, amount float64) (*payment.GatewayResponse, error) {
	body := map[string]interface{}{}
	if amount > 0 {
		body["amount"] = map[string]string{"value": formatAmount(amount)}
	}
	respBody, err := a.postJSON(ctx, "/v2/payments/captures/"+transactionID+"/refund", body)
	if err != nil {
		return nil, err
	}
	var parsed struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	_ = json.Unmarshal(respBody, &parsed)
	return &payment.GatewayResponse{
		TransactionID: parsed.ID,
		Status:        parsed.Status,
		RawResponse:   string(respBody),
	}, nil
}

// Void cancels an uncaptured PayPal order.
func (a *Adapter) Void(ctx context.Context, transactionID string) (*payment.GatewayResponse, error) {
	token, err := a.ensureToken(ctx)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL()+"/v2/checkout/orders/"+transactionID+"/void", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paypal void (%d): %s", resp.StatusCode, string(respBody))
	}
	return &payment.GatewayResponse{TransactionID: transactionID, Status: "VOIDED", RawResponse: string(respBody)}, nil
}

func (a *Adapter) ensureToken(ctx context.Context) (string, error) {
	a.tokenMu.Lock()
	defer a.tokenMu.Unlock()
	if a.accessToken != "" && time.Now().Before(a.tokenExpiry) {
		return a.accessToken, nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL()+"/v1/oauth2/token", strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(a.clientID, a.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("paypal oauth (%d): %s", resp.StatusCode, string(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil {
		return "", err
	}
	a.accessToken = tok.AccessToken
	a.tokenExpiry = time.Now().Add(time.Duration(tok.ExpiresIn-60) * time.Second)
	return a.accessToken, nil
}

func (a *Adapter) postJSON(ctx context.Context, path string, payload interface{}) ([]byte, error) {
	token, err := a.ensureToken(ctx)
	if err != nil {
		return nil, err
	}
	var body io.Reader
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL()+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paypal %s (%d): %s", path, resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

// amountsEqual reports whether the captured amount string reported by PayPal
// matches the requested amount. PayPal reports values with the currency's
// decimal places (e.g. "100.00" for USD, "1234" for JPY), so a plain string
// compare against formatAmount would misfire on zero-decimal currencies; a
// numeric compare with a half-cent tolerance is currency-agnostic (G16).
func amountsEqual(value string, want float64) bool {
	got, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return false
	}
	return math.Abs(got-want) < 0.005
}

// VerifyWebhook validates a PayPal webhook transmission against PayPal's
// signature headers before any payment mutation is performed (mirrors
// Stripe's ValidateWebhookPayload). Requires the adapter to have been
// constructed with the registered webhook ID (New's webhookID argument).
//
// PayPal signs the concatenation
//
//	transmission_id + "|" + transmission_time + "|" + webhook_id + "|" + raw_body
//
// with RSA-SHA256 using the public certificate published at PayPal-Cert-Url.
// The cert URL is restricted to PayPal-owned hosts to avoid SSRF, and the
// transmission timestamp is required to be within ±5 minutes.
func (a *Adapter) VerifyWebhook(payload []byte, transmissionID, transmissionTime, transmissionSig, certURL, authAlgo string) ([]byte, error) {
	if a.webhookID == "" {
		return nil, fmt.Errorf("paypal: webhook id not configured")
	}
	if strings.TrimSpace(transmissionID) == "" || strings.TrimSpace(transmissionTime) == "" || strings.TrimSpace(transmissionSig) == "" || strings.TrimSpace(certURL) == "" {
		return nil, fmt.Errorf("paypal: missing webhook transmission headers")
	}
	algo := strings.TrimSpace(authAlgo)
	if algo != "" && !strings.EqualFold(algo, "SHA256withRSA") {
		return nil, fmt.Errorf("paypal: unsupported auth algo: %s", algo)
	}

	// Only accept certificates from PayPal-owned hosts (SSRF guard).
	u, err := url.Parse(strings.TrimSpace(certURL))
	if err != nil {
		return nil, fmt.Errorf("paypal: invalid cert url: %w", err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("paypal: cert url must be https")
	}
	host := strings.ToLower(u.Hostname())
	if host != "api.paypal.com" && host != "api.sandbox.paypal.com" {
		return nil, fmt.Errorf("paypal: cert url host not allowed: %s", host)
	}

	// Reject stale transmissions.
	tt, err := time.Parse(time.RFC3339, strings.TrimSpace(transmissionTime))
	if err != nil {
		return nil, fmt.Errorf("paypal: invalid transmission time: %w", err)
	}
	diff := time.Since(tt)
	if diff > 5*time.Minute || diff < -5*time.Minute {
		return nil, fmt.Errorf("paypal: transmission time outside tolerance window")
	}

	cert, err := a.certFetch(context.Background(), u.String())
	if err != nil {
		return nil, err
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("paypal: certificate public key is not RSA")
	}
	if err := verifyTransmissionSignature(pub, payload, transmissionID, transmissionTime, a.webhookID, transmissionSig); err != nil {
		return nil, err
	}
	return payload, nil
}

// verifyTransmissionSignature performs the pure cryptographic check: builds the
// signed transmission string, URL-unescapes and base64-decodes the signature,
// and verifies it with RSA-SHA256 against the certificate's public key.
func verifyTransmissionSignature(pub *rsa.PublicKey, payload []byte, transmissionID, transmissionTime, webhookID, transmissionSig string) error {
	message := strings.Join([]string{
		strings.TrimSpace(transmissionID),
		strings.TrimSpace(transmissionTime),
		webhookID,
		string(payload),
	}, "|")

	// PayPal URL-encodes the signature header value; base64 may omit padding.
	sig := transmissionSig
	if unescaped, uerr := url.PathUnescape(transmissionSig); uerr == nil {
		sig = unescaped
	}
	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		sigBytes, err = base64.RawStdEncoding.DecodeString(sig)
	}
	if err != nil {
		return fmt.Errorf("paypal: decode transmission signature: %w", err)
	}

	digest := sha256.Sum256([]byte(message))
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, digest[:], sigBytes); err != nil {
		return fmt.Errorf("paypal: signature verification failed: %w", err)
	}
	return nil
}

// fetchCertWithContext is the default certFetch implementation. It ignores the
// context (matching the historical behavior of fetchCert, which made an
// uncancellable request) and delegates to the cached HTTP fetcher.
func (a *Adapter) fetchCertWithContext(_ context.Context, certURL string) (*x509.Certificate, error) {
	return a.fetchCert(certURL)
}

// certCache caches PayPal signing certificates keyed by cert URL. Certs rotate
// rarely, so a 24h TTL avoids a network round-trip on every webhook.
var (
	certCacheMu sync.Mutex
	certCache   = map[string]certCacheEntry{}
)

type certCacheEntry struct {
	cert *x509.Certificate
	at   time.Time
}

func (a *Adapter) fetchCert(certURL string) (*x509.Certificate, error) {
	certCacheMu.Lock()
	if e, ok := certCache[certURL]; ok && time.Since(e.at) < 24*time.Hour {
		certCacheMu.Unlock()
		return e.cert, nil
	}
	certCacheMu.Unlock()

	req, err := http.NewRequest(http.MethodGet, certURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("paypal: fetch cert: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("paypal: fetch cert (%d): %s", resp.StatusCode, certURL)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(body)
	if block == nil {
		return nil, fmt.Errorf("paypal: no PEM block in certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("paypal: parse certificate: %w", err)
	}

	certCacheMu.Lock()
	certCache[certURL] = certCacheEntry{cert: cert, at: time.Now()}
	certCacheMu.Unlock()
	return cert, nil
}

var _ payment.PaymentGateway = (*Adapter)(nil)
