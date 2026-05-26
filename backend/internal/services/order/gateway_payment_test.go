package order

import (
	"testing"

	modelsOrder "candypro/api/internal/models/order"
)

// TestParsePayPalWebhookResource 验证 PayPal webhook 解析
func TestParsePayPalWebhookResource(t *testing.T) {
	raw := []byte(`{"resource":{"reference_id":"pay-1","custom_id":"ord-1"}}`)
	payID, orderID := ParsePayPalWebhookResource(raw)
	if payID != "pay-1" || orderID != "ord-1" {
		t.Fatalf("unexpected parse: pay=%q order=%q", payID, orderID)
	}
}

// TestGatewayPaymentService_unconfigured 未配置网关时应返回错误
func TestGatewayPaymentService_unconfigured(t *testing.T) {
	svc := NewGatewayPaymentService(nil, nil)
	_, err := svc.gateway(modelsOrder.PaymentMethodStripe)
	if err == nil {
		t.Fatal("expected error for unconfigured stripe")
	}
}
