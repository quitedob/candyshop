package paypal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

	tokenMu     sync.Mutex
	accessToken string
	tokenExpiry time.Time
}

// New creates a PayPal adapter. sandbox=true uses PayPal sandbox API.
func New(clientID, clientSecret, webhookID string, sandbox bool) *Adapter {
	return &Adapter{
		clientID:     clientID,
		clientSecret: clientSecret,
		webhookID:    webhookID,
		sandbox:      sandbox,
		httpClient:   &http.Client{Timeout: 20 * time.Second},
	}
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
func (a *Adapter) Capture(ctx context.Context, transactionID string, _ float64) (*payment.GatewayResponse, error) {
	respBody, err := a.postJSON(ctx, "/v2/checkout/orders/"+transactionID+"/capture", nil)
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

var _ payment.PaymentGateway = (*Adapter)(nil)
