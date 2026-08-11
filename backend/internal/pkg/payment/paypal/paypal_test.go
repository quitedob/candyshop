package paypal

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// testMaterial carries a fresh 2048-bit RSA keypair plus its self-signed cert.
type testMaterial struct {
	priv *rsa.PrivateKey
	cert *x509.Certificate
	pem  []byte
}

func makeTestMaterial(t *testing.T) *testMaterial {
	t.Helper()
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "api.paypal.com"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return &testMaterial{priv: priv, cert: cert, pem: pemBytes}
}

// certServer serves the certificate PEM at /cert.pem.
func certServer(t *testing.T, pemBytes []byte) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cert.pem" {
			_, _ = w.Write(pemBytes)
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// fetchFromServer returns a certFetch func that downloads and parses the PEM
// certificate served by srv at /cert.pem. The injected certURL is ignored
// because VerifyWebhook already validated it against the SSRF host guard.
func fetchFromServer(srv *httptest.Server) func(context.Context, string) (*x509.Certificate, error) {
	return func(_ context.Context, _ string) (*x509.Certificate, error) {
		resp, err := http.Get(srv.URL + "/cert.pem")
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(body)
		if block == nil {
			return nil, fmt.Errorf("paypal: no PEM block in certificate")
		}
		return x509.ParseCertificate(block.Bytes)
	}
}

// signTransmission reproduces VerifyWebhook's signed message construction and
// returns the base64-encoded RSA-SHA256 signature.
func signTransmission(t *testing.T, priv *rsa.PrivateKey, transmissionID, transmissionTime, webhookID string, payload []byte) string {
	t.Helper()
	message := strings.Join([]string{transmissionID, transmissionTime, webhookID, string(payload)}, "|")
	digest := sha256.Sum256([]byte(message))
	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign transmission: %v", err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}

// staticCertFetcher returns a certFetch that records whether it was invoked.
func staticCertFetcher(m *testMaterial, fetched *bool) func(context.Context, string) (*x509.Certificate, error) {
	return func(_ context.Context, _ string) (*x509.Certificate, error) {
		*fetched = true
		return m.cert, nil
	}
}

const testCertURL = "https://api.paypal.com/cert.pem"

func TestVerifyWebhookHappyPath(t *testing.T) {
	m := makeTestMaterial(t)
	srv := certServer(t, m.pem)

	a := New("client-id", "client-secret", "wh_test", false)
	// Bypass the SSRF host guard by injecting a fetcher that pulls the cert
	// from the local test server instead of api.paypal.com.
	a.certFetch = fetchFromServer(srv)

	payload := []byte(`{"event_type":"CHECKOUT.ORDER.APPROVED","resource":{"id":"cap1","reference_id":"pay-1"}}`)
	transmissionID := "txn-123"
	transmissionTime := time.Now().Format(time.RFC3339)
	sig := signTransmission(t, m.priv, transmissionID, transmissionTime, "wh_test", payload)

	got, err := a.VerifyWebhook(payload, transmissionID, transmissionTime, sig, testCertURL, "SHA256withRSA")
	if err != nil {
		t.Fatalf("expected verification to succeed, got: %v", err)
	}
	if string(got) != string(payload) {
		t.Fatalf("expected payload echoed back, got %q", string(got))
	}
}

func TestVerifyWebhookHappyPathRawUnpaddedSig(t *testing.T) {
	m := makeTestMaterial(t)
	a := New("client-id", "client-secret", "wh_test", false)
	a.certFetch = staticCertFetcher(m, new(bool))

	payload := []byte(`{"event_type":"PAYMENT.CAPTURE.COMPLETED"}`)
	transmissionID := "txn-raw"
	transmissionTime := time.Now().Format(time.RFC3339)
	sig := signTransmission(t, m.priv, transmissionID, transmissionTime, "wh_test", payload)
	rawSig := strings.TrimRight(sig, "=") // base64 without padding

	_, err := a.VerifyWebhook(payload, transmissionID, transmissionTime, rawSig, testCertURL, "SHA256withRSA")
	if err != nil {
		t.Fatalf("expected unpadded signature to verify, got: %v", err)
	}
}

func TestVerifyWebhookBadSignature(t *testing.T) {
	m := makeTestMaterial(t)
	a := New("client-id", "client-secret", "wh_test", false)
	a.certFetch = staticCertFetcher(m, new(bool))

	payload := []byte(`{"event_type":"CHECKOUT.ORDER.APPROVED"}`)
	transmissionID := "txn-1"
	transmissionTime := time.Now().Format(time.RFC3339)
	sig := signTransmission(t, m.priv, transmissionID, transmissionTime, "wh_test", payload)

	// Corrupt a byte of the signature before base64-encoding.
	raw, err := base64.StdEncoding.DecodeString(sig)
	if err != nil {
		t.Fatalf("decode signature: %v", err)
	}
	raw[0] ^= 0xFF
	badSig := base64.StdEncoding.EncodeToString(raw)

	_, err = a.VerifyWebhook(payload, transmissionID, transmissionTime, badSig, testCertURL, "SHA256withRSA")
	if err == nil {
		t.Fatal("expected corrupted signature to be rejected")
	}
	if !strings.Contains(err.Error(), "signature verification failed") {
		t.Fatalf("expected signature failure error, got: %v", err)
	}
}

func TestVerifyWebhookStaleTimestamp(t *testing.T) {
	m := makeTestMaterial(t)
	a := New("client-id", "client-secret", "wh_test", false)
	var fetched bool
	a.certFetch = staticCertFetcher(m, &fetched)

	payload := []byte(`{"event_type":"CHECKOUT.ORDER.APPROVED"}`)

	for _, tt := range []struct {
		name string
		at   time.Time
	}{
		{"too old", time.Now().Add(-6 * time.Minute)},
		{"too far in the future", time.Now().Add(6 * time.Minute)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			fetched = false
			transmissionID := "txn-1"
			transmissionTime := tt.at.Format(time.RFC3339)
			sig := signTransmission(t, m.priv, transmissionID, transmissionTime, "wh_test", payload)

			_, err := a.VerifyWebhook(payload, transmissionID, transmissionTime, sig, testCertURL, "SHA256withRSA")
			if err == nil {
				t.Fatal("expected stale transmission to be rejected")
			}
			if !strings.Contains(err.Error(), "outside tolerance window") {
				t.Fatalf("expected tolerance window error, got: %v", err)
			}
			if fetched {
				t.Fatal("cert fetch should NOT have been attempted for a stale transmission")
			}
		})
	}
}

func TestVerifyWebhookMissingHeaders(t *testing.T) {
	m := makeTestMaterial(t)
	a := New("client-id", "client-secret", "wh_test", false)
	var fetched bool
	a.certFetch = staticCertFetcher(m, &fetched)

	_, err := a.VerifyWebhook([]byte(`{}`), "", "", "", "", "")
	if err == nil {
		t.Fatal("expected missing transmission headers to be rejected")
	}
	if !strings.Contains(err.Error(), "missing webhook transmission headers") {
		t.Fatalf("expected missing headers error, got: %v", err)
	}
	if fetched {
		t.Fatal("cert fetch should NOT have been attempted with missing headers")
	}
}

func TestVerifyWebhookMissingWebhookID(t *testing.T) {
	a := New("client-id", "client-secret", "", false) // no registered webhook id

	payload := []byte(`{"event_type":"CHECKOUT.ORDER.APPROVED"}`)
	_, err := a.VerifyWebhook(payload, "txn-1", time.Now().Format(time.RFC3339), "c2ln", testCertURL, "SHA256withRSA")
	if err == nil {
		t.Fatal("expected missing webhook id to be rejected")
	}
	if !strings.Contains(err.Error(), "webhook id not configured") {
		t.Fatalf("expected webhook id error, got: %v", err)
	}
}

func TestVerifyWebhookUnsupportedAlgo(t *testing.T) {
	a := New("client-id", "client-secret", "wh_test", false)
	var fetched bool
	a.certFetch = staticCertFetcher(makeTestMaterial(t), &fetched)

	_, err := a.VerifyWebhook([]byte(`{}`), "txn-1", time.Now().Format(time.RFC3339), "c2ln", testCertURL, "MD5")
	if err == nil {
		t.Fatal("expected unsupported auth algo to be rejected")
	}
	if !strings.Contains(err.Error(), "unsupported auth algo") {
		t.Fatalf("expected unsupported algo error, got: %v", err)
	}
	if fetched {
		t.Fatal("cert fetch should NOT have been attempted with an unsupported algo")
	}
}

// ---- G16: capture amount verification -------------------------------------

// captureRoundTripper rewrites every request to the fake PayPal server so the
// adapter (which always targets api-m.paypal.com) talks to the test server.
type captureRoundTripper struct {
	base *url.URL
}

func (t *captureRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req2 := req.Clone(req.Context())
	req2.URL.Scheme = t.base.Scheme
	req2.URL.Host = t.base.Host
	return http.DefaultTransport.RoundTrip(req2)
}

// captureServer returns a fake PayPal API whose capture endpoint reports the
// given captured value/currency, plus an OAuth token endpoint.
func captureServer(t *testing.T, capturedValue, currency string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/oauth2/token":
			_, _ = io.WriteString(w, `{"access_token":"tok","expires_in":3600}`)
		case "/v2/checkout/orders/ORD/capture":
			_, _ = io.WriteString(w, fmt.Sprintf(
				`{"id":"ORD","status":"COMPLETED","purchase_units":[{"payments":{"captures":[{"id":"CAP1","status":"COMPLETED","amount":{"currency_code":"%s","value":"%s"}}]}}]}`,
				currency, capturedValue))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func captureAdapter(t *testing.T, srv *httptest.Server) *Adapter {
	t.Helper()
	u, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parse test server url: %v", err)
	}
	a := New("cid", "csec", "wh_test", false)
	a.httpClient = &http.Client{Transport: &captureRoundTripper{base: u}}
	return a
}

// TestCaptureAmountMismatch is the G16 regression test: when PayPal captures a
// different amount than the one requested, the adapter must FAIL (the local
// payment would otherwise record money taken for the wrong value). On the
// pre-fix path the amount argument was dropped and the capture "succeeded"
// silently, returning nil.
func TestCaptureAmountMismatch(t *testing.T) {
	a := captureAdapter(t, captureServer(t, "95.00", "USD"))
	_, err := a.Capture(context.Background(), "ORD", 100.00)
	if err == nil {
		t.Fatal("expected capture to fail when PayPal captured a different amount")
	}
	if !strings.Contains(err.Error(), "does not match") {
		t.Fatalf("expected amount mismatch error, got: %v", err)
	}
}

// TestCaptureAmountMatches ensures a capture whose reported amount equals the
// requested amount still succeeds.
func TestCaptureAmountMatches(t *testing.T) {
	a := captureAdapter(t, captureServer(t, "100.00", "USD"))
	if _, err := a.Capture(context.Background(), "ORD", 100.00); err != nil {
		t.Fatalf("expected capture to succeed, got: %v", err)
	}
}

// TestCaptureZeroDecimalAmountMatches verifies the amount comparison is
// currency-agnostic: PayPal reports zero-decimal amounts without a fractional
// part ("1234" for JPY), which must still match a requested 1234.
func TestCaptureZeroDecimalAmountMatches(t *testing.T) {
	a := captureAdapter(t, captureServer(t, "1234", "JPY"))
	if _, err := a.Capture(context.Background(), "ORD", 1234); err != nil {
		t.Fatalf("expected JPY capture to succeed, got: %v", err)
	}
}
