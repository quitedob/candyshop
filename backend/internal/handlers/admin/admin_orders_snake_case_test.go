package admin

import (
	"encoding/json"
	"testing"

	"candypro/api/internal/pkg/request"
)

func TestAdminUpdateOrderRequest_SnakeCasePaymentStatus(t *testing.T) {
	body := []byte(`{"payment_status":"paid"}`)
	var req adminUpdateOrderRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ps := request.FirstOptionalNonEmpty(req.PaymentStatus, req.PaymentStatusSnake)
	if ps == nil || *ps != "paid" {
		t.Fatalf("expected paid, got %v", ps)
	}
}

func TestAdminUpdateOrderRequest_SnakeCaseTrackingNumber(t *testing.T) {
	body := []byte(`{"tracking_number":"FEDEX-001"}`)
	var req adminUpdateOrderRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	tn := request.FirstOptionalNonEmpty(req.TrackingNumber, req.TrackingNumberSnake)
	if tn == nil || *tn != "FEDEX-001" {
		t.Fatalf("expected FEDEX-001, got %v", tn)
	}
}

func TestAdminUpdateOrderStatusRequest_SnakeCaseTrackingNumber(t *testing.T) {
	body := []byte(`{"status":"shipped","tracking_number":"TEST-001"}`)
	var req struct {
		Status              string `json:"status"`
		TrackingNumber      string `json:"trackingNumber"`
		TrackingNumberSnake string `json:"tracking_number"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	got := request.FirstNonEmpty(req.TrackingNumber, req.TrackingNumberSnake)
	if got != "TEST-001" {
		t.Fatalf("tracking = %q, want TEST-001", got)
	}
}

func TestAdminUpdateOrderRequest_CamelCasePreferred(t *testing.T) {
	body := []byte(`{"paymentStatus":"partial","payment_status":"paid"}`)
	var req adminUpdateOrderRequest
	if err := json.Unmarshal(body, &req); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	ps := request.FirstOptionalNonEmpty(req.PaymentStatus, req.PaymentStatusSnake)
	if ps == nil || *ps != "partial" {
		t.Fatalf("camelCase should win, got %v", ps)
	}
}
