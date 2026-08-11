package order

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"candypro/api/internal/config"
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/payment"
	paypalAdapter "candypro/api/internal/pkg/payment/paypal"
	stripeAdapter "candypro/api/internal/pkg/payment/stripe"
)

// GatewayCheckoutResult 网关支付创建结果
type GatewayCheckoutResult struct {
	PaymentID            string  `json:"paymentId"`
	Method               string  `json:"method"`
	GatewayTransactionID string  `json:"gatewayTransactionId"`
	ClientSecret         string  `json:"clientSecret,omitempty"`
	ApprovalURL          string  `json:"approvalUrl,omitempty"`
	PublishableKey       string  `json:"publishableKey,omitempty"`
	Amount               float64 `json:"amount"`
	Currency             string  `json:"currency"`
}

// GatewayPaymentService 封装 Stripe/PayPal 网关操作
type GatewayPaymentService struct {
	payments       *PaymentService
	stripe         *stripeAdapter.Adapter
	paypal         *paypalAdapter.Adapter
	publishableKey string
	frontendURL    string
}

// NewGatewayPaymentService 创建网关支付服务
func NewGatewayPaymentService(cfg *config.Config, payments *PaymentService) *GatewayPaymentService {
	svc := &GatewayPaymentService{payments: payments}
	if cfg != nil {
		svc.stripe = stripeAdapter.New(cfg.Stripe.SecretKey, cfg.Stripe.WebhookSecret)
		svc.paypal = paypalAdapter.New(cfg.PayPal.ClientID, cfg.PayPal.ClientSecret, cfg.PayPal.WebhookID, cfg.PayPal.Sandbox)
		svc.publishableKey = cfg.Stripe.PublishableKey
		svc.frontendURL = cfg.Security.FrontendURL
	}
	return svc
}

func (s *GatewayPaymentService) gateway(method string) (payment.PaymentGateway, error) {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case modelsOrder.PaymentMethodStripe:
		if s.stripe == nil || !s.stripe.IsConfigured() {
			return nil, fmt.Errorf("stripe_not_configured")
		}
		return s.stripe, nil
	case modelsOrder.PaymentMethodPayPal:
		if s.paypal == nil || !s.paypal.IsConfigured() {
			return nil, fmt.Errorf("paypal_not_configured")
		}
		return s.paypal, nil
	default:
		return nil, fmt.Errorf("unsupported_gateway_method")
	}
}

// CreateCheckout persists a pending payment row (with atomic balance check) and
// THEN calls the gateway to create the intent/checkout. Persisting before the
// gateway call means a DB failure after gateway success can no longer orphan a
// collectible intent, and the payment amount is committed to the order's balance
// while concurrent checkout calls for the same order are serialized so two
// creates cannot both pass the balance check (M8). If the gateway call fails the
// persisted row is marked failed instead of leaving an orphaned intent.
func (s *GatewayPaymentService) CreateCheckout(ctx context.Context, order *modelsOrder.Order, userID, method string, amount float64) (*GatewayCheckoutResult, error) {
	if s.payments == nil || order == nil {
		return nil, fmt.Errorf("service_unavailable")
	}
	if order.UserID != userID {
		return nil, fmt.Errorf("forbidden")
	}
	gw, err := s.gateway(method)
	if err != nil {
		return nil, err
	}
	if amount <= 0 {
		amount = order.TotalAmount
	}
	paymentID := crypto.GenerateID()
	// Persist the payment row (status pending) first. CreatePaymentWithBalanceCheck
	// re-reads the order's confirmed/pending payments and inserts inside one DB
	// transaction, and serializes concurrent creations, so the amount is committed
	// to the remaining balance before we reach the gateway.
	rec := &modelsOrder.Payment{
		ID:        paymentID,
		OrderID:   order.ID,
		Amount:    amount,
		Currency:  order.Currency,
		Method:    method,
		Status:    modelsOrder.PaymentRecordStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.payments.CreatePaymentWithBalanceCheck(ctx, order.TotalAmount, rec); err != nil {
		return nil, err
	}
	// Only now create the gateway intent/checkout.
	returnURL := strings.TrimRight(s.frontendURL, "/") + "/customer/orders/" + order.ID
	req := payment.GatewayRequest{
		Amount:         amount,
		Currency:       order.Currency,
		Description:    "Order " + order.OrderNumber,
		OrderID:        order.ID,
		PaymentID:      paymentID,
		IdempotencyKey: paymentID,
		Metadata: map[string]string{
			"return_url": returnURL,
			"cancel_url": returnURL,
		},
	}
	resp, err := gw.Authorize(ctx, req)
	if err != nil {
		// Gateway failed after the row was persisted: mark the row failed rather
		// than leaving a collectible intent with no matching payment record.
		_ = s.payments.FailPayment(ctx, paymentID)
		return nil, err
	}
	txID := resp.TransactionID
	if err := s.payments.AttachGatewayResult(ctx, paymentID, txID, resp.RawResponse); err != nil {
		return nil, err
	}
	out := &GatewayCheckoutResult{
		PaymentID:            paymentID,
		Method:               method,
		GatewayTransactionID: txID,
		ClientSecret:         resp.ClientSecret,
		ApprovalURL:          resp.ApprovalURL,
		Amount:               amount,
		Currency:             order.Currency,
	}
	if method == modelsOrder.PaymentMethodStripe {
		out.PublishableKey = s.publishableKey
	}
	return out, nil
}

// CaptureAndConfirm 捕获已授权/待确认网关支付并确认
func (s *GatewayPaymentService) CaptureAndConfirm(ctx context.Context, pay *modelsOrder.Payment, confirmedBy string) error {
	if s.payments == nil || pay == nil {
		return fmt.Errorf("service_unavailable")
	}
	if pay.GatewayTransactionID == nil || strings.TrimSpace(*pay.GatewayTransactionID) == "" {
		return s.payments.ConfirmPayment(ctx, pay.ID, confirmedBy)
	}
	gw, err := s.gateway(pay.Method)
	if err != nil {
		return err
	}
	if pay.Status == modelsOrder.PaymentRecordStatusAuthorized || pay.Method == modelsOrder.PaymentMethodPayPal {
		if _, capErr := gw.Capture(ctx, *pay.GatewayTransactionID, pay.Amount); capErr != nil {
			return capErr
		}
	}
	if pay.Status == modelsOrder.PaymentRecordStatusAuthorized {
		return s.payments.ConfirmAuthorizedPayment(ctx, pay.ID, confirmedBy, pay.Amount)
	}
	return s.payments.ConfirmPayment(ctx, pay.ID, confirmedBy)
}

// RefundGateway 通过网关退款并更新本地记录
func (s *GatewayPaymentService) RefundGateway(ctx context.Context, pay *modelsOrder.Payment) error {
	if s.payments == nil || pay == nil {
		return fmt.Errorf("service_unavailable")
	}
	if pay.GatewayTransactionID != nil && strings.TrimSpace(*pay.GatewayTransactionID) != "" {
		if gw, err := s.gateway(pay.Method); err == nil {
			if _, refundErr := gw.Refund(ctx, *pay.GatewayTransactionID, pay.Amount); refundErr != nil {
				return refundErr
			}
		}
	}
	return s.payments.RefundPayment(ctx, pay.ID)
}

// ParsePayPalWebhookResource 解析 PayPal webhook 资源
func ParsePayPalWebhookResource(raw []byte) (paymentID, orderID string) {
	var payload struct {
		Resource struct {
			CustomID    string `json:"custom_id"`
			ReferenceID string `json:"reference_id"`
		} `json:"resource"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", ""
	}
	return payload.Resource.ReferenceID, payload.Resource.CustomID
}
