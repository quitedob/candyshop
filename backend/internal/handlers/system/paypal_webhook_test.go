package system

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	paypalAdapter "candypro/api/internal/pkg/payment/paypal"
	orderRepo "candypro/api/internal/repository/order"
	servicesCommon "candypro/api/internal/services/common"
	orderSvc "candypro/api/internal/services/order"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// TestHandlePayPalWebhook_unconfigured ensures the endpoint fails closed with
// 503 when the PayPal adapter has no credentials configured.
func TestHandlePayPalWebhook_unconfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{
		services:      &servicesCommon.SystemServices{Payment: &orderSvc.PaymentService{}},
		paypalAdapter: paypalAdapter.New("", "", "", false), // no client id/secret
	}

	r := gin.New()
	r.POST("/paypal-webhook", h.HandlePayPalWebhook)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/paypal-webhook", strings.NewReader(`{}`))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d (body: %s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "paypal_not_configured") {
		t.Fatalf("expected paypal_not_configured body, got %q", w.Body.String())
	}
}

// TestHandlePayPalWebhook_nilAdapter ensures a nil adapter also fails closed.
func TestHandlePayPalWebhook_nilAdapter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{
		services:      &servicesCommon.SystemServices{Payment: &orderSvc.PaymentService{}},
		paypalAdapter: nil,
	}

	r := gin.New()
	r.POST("/paypal-webhook", h.HandlePayPalWebhook)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/paypal-webhook", strings.NewReader(`{}`))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d (body: %s)", w.Code, w.Body.String())
	}
}

// TestHandlePayPalWebhook_missingHeaders ensures a configured adapter still
// rejects a webhook that carries no PayPal transmission headers (400).
func TestHandlePayPalWebhook_missingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &Handler{
		services:      &servicesCommon.SystemServices{Payment: &orderSvc.PaymentService{}},
		paypalAdapter: paypalAdapter.New("client-id", "client-secret", "wh_test", false), // configured
	}

	r := gin.New()
	r.POST("/paypal-webhook", h.HandlePayPalWebhook)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/paypal-webhook", strings.NewReader(`{"event_type":"CHECKOUT.ORDER.APPROVED"}`))
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (body: %s)", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "invalid_signature") {
		t.Fatalf("expected invalid_signature body, got %q", w.Body.String())
	}
}

// ---- G16: APPROVED-vs-captured distinction + event replay dedupe -----------

// newPaymentServiceWithDB builds a PaymentService backed by an in-memory sqlite
// DB with one pending PayPal payment row and its parent order.
func newPaymentServiceWithDB(t *testing.T) (*orderSvc.PaymentService, *modelsOrder.Payment) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsOrder.Order{}, &modelsOrder.Payment{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	paySvc := orderSvc.NewPaymentService(orderRepo.NewPaymentRepository(db), orderRepo.NewOrderRepository(db))

	order := &modelsOrder.Order{
		ID:            "ord-g16",
		OrderNumber:   "ORD-G16-1001",
		UserID:        "user-g16",
		Status:        "pending",
		PaymentStatus: "unpaid",
		TotalAmount:   100.0,
		Currency:      "USD",
		Items:         modelsOrder.OrderItemArray{},
	}
	if err := db.Create(order).Error; err != nil {
		t.Fatalf("create order: %v", err)
	}
	pay := &modelsOrder.Payment{
		ID:       "pay-g16",
		OrderID:  order.ID,
		Amount:   100.0,
		Currency: "USD",
		Method:   modelsOrder.PaymentMethodPayPal,
		Status:   modelsOrder.PaymentRecordStatusPending,
	}
	if err := db.Create(pay).Error; err != nil {
		t.Fatalf("create payment: %v", err)
	}
	return paySvc, pay
}

func getPaymentStatus(t *testing.T, paySvc *orderSvc.PaymentService, id string) string {
	t.Helper()
	got, err := paySvc.GetPayment(context.Background(), id)
	if err != nil {
		t.Fatalf("get payment %s: %v", id, err)
	}
	return got.Status
}

// TestHandlePayPalEvent_ApprovedDoesNotConfirm is the core G16 regression test:
// CHECKOUT.ORDER.APPROVED only means the buyer approved — funds are NOT captured
// — so the payment must NOT be confirmed when there is no gateway path to
// actually capture money. On the pre-fix path the shared APPROVED/CAPTURE branch
// fell through to a bare ConfirmPayment when the gateway was nil, confirming the
// payment and shipping goods before any money moved.
func TestHandlePayPalEvent_ApprovedDoesNotConfirm(t *testing.T) {
	paySvc, pay := newPaymentServiceWithDB(t)
	h := &Handler{
		// GatewayPayment deliberately nil: there is no way to capture money.
		services: &servicesCommon.SystemServices{Payment: paySvc},
	}

	h.handlePayPalPaymentEvent(context.Background(), "CHECKOUT.ORDER.APPROVED", pay.ID)

	if st := getPaymentStatus(t, paySvc, pay.ID); st != modelsOrder.PaymentRecordStatusPending {
		t.Fatalf("APPROVED must NOT confirm the payment without a successful capture; got status %q", st)
	}
}

// TestHandlePayPalEvent_ApprovedNoCaptureAttemptedWithoutTxID ensures that an
// APPROVED event for a payment that carries a gateway transaction id but no
// gateway service leaves the payment pending (no bare confirm).
func TestHandlePayPalEvent_ApprovedNoCaptureAttemptedWithoutTxID(t *testing.T) {
	paySvc, pay := newPaymentServiceWithDB(t)
	txID := "paypal-order-1"
	if err := paySvc.AttachGatewayResult(context.Background(), pay.ID, txID, `{}`); err != nil {
		t.Fatalf("attach gateway result: %v", err)
	}
	h := &Handler{services: &servicesCommon.SystemServices{Payment: paySvc}} // GatewayPayment nil

	h.handlePayPalPaymentEvent(context.Background(), "CHECKOUT.ORDER.APPROVED", pay.ID)
	if st := getPaymentStatus(t, paySvc, pay.ID); st != modelsOrder.PaymentRecordStatusPending {
		t.Fatalf("APPROVED must not confirm without a gateway capture; got status %q", st)
	}
}

// TestHandlePayPalEvent_CaptureCompletedConfirms verifies that PAYMENT.CAPTURE.
// COMPLETED — which PayPal emits only AFTER funds are captured — confirms the
// local payment directly.
func TestHandlePayPalEvent_CaptureCompletedConfirms(t *testing.T) {
	paySvc, pay := newPaymentServiceWithDB(t)
	h := &Handler{services: &servicesCommon.SystemServices{Payment: paySvc}}

	h.handlePayPalPaymentEvent(context.Background(), "PAYMENT.CAPTURE.COMPLETED", pay.ID)

	if st := getPaymentStatus(t, paySvc, pay.ID); st != modelsOrder.PaymentRecordStatusConfirmed {
		t.Fatalf("CAPTURE.COMPLETED must confirm the payment; got status %q", st)
	}
}

// TestHandlePayPalEvent_CaptureCompletedIdempotent verifies a replayed
// CAPTURE.COMPLETED event does not double-process: the second confirm is
// rejected by the payment state machine (payment already confirmed).
func TestHandlePayPalEvent_CaptureCompletedIdempotent(t *testing.T) {
	paySvc, pay := newPaymentServiceWithDB(t)
	h := &Handler{services: &servicesCommon.SystemServices{Payment: paySvc}}

	h.handlePayPalPaymentEvent(context.Background(), "PAYMENT.CAPTURE.COMPLETED", pay.ID)
	h.handlePayPalPaymentEvent(context.Background(), "PAYMENT.CAPTURE.COMPLETED", pay.ID)

	if st := getPaymentStatus(t, paySvc, pay.ID); st != modelsOrder.PaymentRecordStatusConfirmed {
		t.Fatalf("payment must remain confirmed after replay; got status %q", st)
	}
}

// TestHandlePayPalEvent_DeniedFails verifies PAYMENT.CAPTURE.DENIED fails the
// payment.
func TestHandlePayPalEvent_DeniedFails(t *testing.T) {
	paySvc, pay := newPaymentServiceWithDB(t)
	h := &Handler{services: &servicesCommon.SystemServices{Payment: paySvc}}

	h.handlePayPalPaymentEvent(context.Background(), "PAYMENT.CAPTURE.DENIED", pay.ID)

	if st := getPaymentStatus(t, paySvc, pay.ID); st != modelsOrder.PaymentRecordStatusFailed {
		t.Fatalf("CAPTURE.DENIED must fail the payment; got status %q", st)
	}
}

// TestPayPalEventDedupe verifies the event-id replay dedupe helpers (G16).
func TestPayPalEventDedupe(t *testing.T) {
	markPayPalEventProcessed("evt-g16-1")
	if !isPayPalEventProcessed("evt-g16-1") {
		t.Fatal("expected evt-g16-1 to be marked processed")
	}
	if isPayPalEventProcessed("evt-g16-unknown") {
		t.Fatal("expected an unknown event id to not be processed")
	}
	if isPayPalEventProcessed("") {
		t.Fatal("empty event id must not be treated as processed")
	}
	markPayPalEventProcessed("") // must be a no-op
}
