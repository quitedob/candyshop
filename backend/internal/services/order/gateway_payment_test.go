package order

import (
	"context"
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

// paymentRepoStub 记录是否触发了本地退款标记，用于 RefundGateway 失败闭合测试
type paymentRepoStub struct {
	refundCalled bool
}

func (p *paymentRepoStub) FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Payment, error) {
	return nil, nil
}
func (p *paymentRepoStub) FindByID(ctx context.Context, id string) (*modelsOrder.Payment, error) {
	return nil, nil
}
func (p *paymentRepoStub) FindByGatewayTransactionID(ctx context.Context, txID string) (*modelsOrder.Payment, error) {
	return nil, nil
}
func (p *paymentRepoStub) Create(ctx context.Context, payment *modelsOrder.Payment) error {
	return nil
}
func (p *paymentRepoStub) CreateWithBalanceCheck(ctx context.Context, orderTotalAmount float64, payment *modelsOrder.Payment) error {
	return nil
}
func (p *paymentRepoStub) Update(ctx context.Context, payment *modelsOrder.Payment) error {
	return nil
}
func (p *paymentRepoStub) ConfirmPayment(ctx context.Context, id, confirmedBy string) error {
	return nil
}
func (p *paymentRepoStub) ConfirmPaymentAndRecomputeOrderStatus(ctx context.Context, id, confirmedBy string) (string, string, error) {
	return "", "", nil
}
func (p *paymentRepoStub) MarkPaymentRefunded(ctx context.Context, id string) error {
	return nil
}
func (p *paymentRepoStub) MarkPaymentRefundedAndRecomputeOrderStatus(ctx context.Context, id string) (string, string, error) {
	p.refundCalled = true
	return "", "", nil
}
func (p *paymentRepoStub) UpdateStatus(ctx context.Context, id, currentStatus, newStatus string) error {
	return nil
}
func (p *paymentRepoStub) ConfirmAuthorizedPayment(ctx context.Context, id, confirmedBy string, capturedAmount float64) error {
	return nil
}
func (p *paymentRepoStub) ConfirmAuthorizedPaymentAndRecomputeOrderStatus(ctx context.Context, id, confirmedBy string, capturedAmount float64) (string, string, error) {
	return "", "", nil
}
func (p *paymentRepoStub) CountByStatus(ctx context.Context, status string) (int64, error) {
	return 0, nil
}
func (p *paymentRepoStub) StatusBreakdown(ctx context.Context) (map[string]int64, error) {
	return nil, nil
}

// TestRefundGateway_UnresolvedGateway_FailsClosed 网关不可解析（未配置/未知方法）时，
// 本地支付不得被标记为 refunded，钱仍留在网关（H8）
func TestRefundGateway_UnresolvedGateway_FailsClosed(t *testing.T) {
	tests := []struct {
		name   string
		method string
	}{
		{name: "unconfigured stripe", method: modelsOrder.PaymentMethodStripe},
		{name: "unconfigured paypal", method: modelsOrder.PaymentMethodPayPal},
		{name: "unknown method", method: "bitcoin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &paymentRepoStub{}
			payments := NewPaymentService(repo, nil)
			svc := NewGatewayPaymentService(nil, payments) // stripe/paypal nil => unconfigured
			txID := "tx-123"
			pay := &modelsOrder.Payment{
				ID:                   "pay-1",
				Method:               tt.method,
				GatewayTransactionID: &txID,
				Amount:               100,
			}
			err := svc.RefundGateway(context.Background(), pay)
			if err == nil {
				t.Fatal("expected error when gateway cannot be resolved")
			}
			if repo.refundCalled {
				t.Fatal("local payment must NOT be marked refunded when the gateway cannot be reached")
			}
		})
	}
}

// TestRefundGateway_NoGatewayTransaction_MarksLocalRefunded 无网关交易 ID 的支付
// （如线下/手动退款）不经过网关，仍正常标记本地 refunded，行为保持不变
func TestRefundGateway_NoGatewayTransaction_MarksLocalRefunded(t *testing.T) {
	repo := &paymentRepoStub{}
	payments := NewPaymentService(repo, nil)
	svc := NewGatewayPaymentService(nil, payments)
	pay := &modelsOrder.Payment{
		ID:     "pay-2",
		Method: modelsOrder.PaymentMethodBankTransfer,
		Amount: 50,
	}
	if err := svc.RefundGateway(context.Background(), pay); err != nil {
		t.Fatalf("RefundGateway: %v", err)
	}
	if !repo.refundCalled {
		t.Fatal("expected local payment to be marked refunded for non-gateway payment")
	}
}
