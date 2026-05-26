package order

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSignPayloadHMAC(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"orderId":"123"}`)
	sig := signPayload(secret, payload)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))
	if sig != expected {
		t.Fatalf("signature mismatch: got %s want %s", sig, expected)
	}
}
