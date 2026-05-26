package order

import (
	"testing"

	modelsProduct "candypro/api/internal/models/product"
)

func TestChannelService_ApplyPriceMultiplier(t *testing.T) {
	s := NewChannelService(nil)
	got := s.ApplyPriceMultiplier(100, &modelsProduct.Channel{PriceMultiplier: 1.1})
	if got < 109.99 || got > 110.01 {
		t.Fatalf("expected ~110 got %v", got)
	}
}

func TestChannelService_ComputeTaxAmount(t *testing.T) {
	s := NewChannelService(nil)
	ch := &modelsProduct.Channel{TaxConfig: modelsProduct.ChannelTaxJSON{TaxRate: 0.1}}
	if got := s.ComputeTaxAmount(200, ch, 0); got != 20 {
		t.Fatalf("expected 20 got %v", got)
	}
	if got := s.ComputeTaxAmount(200, ch, 5); got != 5 {
		t.Fatalf("client tax preserved, got %v", got)
	}
}

func TestChannelService_ResolveInitialOrderStatus(t *testing.T) {
	s := NewChannelService(nil)
	if st := s.ResolveInitialOrderStatus(&modelsProduct.Channel{AutoConfirm: true}, false); st != "confirmed" {
		t.Fatalf("auto confirm expected confirmed got %s", st)
	}
	if st := s.ResolveInitialOrderStatus(&modelsProduct.Channel{AutoConfirm: true}, true); st != "pending_approval" {
		t.Fatalf("approval overrides auto confirm")
	}
}

func TestChannelService_ResolvePaymentStatus(t *testing.T) {
	s := NewChannelService(nil)
	if got := s.ResolvePaymentStatus(nil); got != "unpaid" {
		t.Fatalf("nil channel expected unpaid got %s", got)
	}
	if got := s.ResolvePaymentStatus(&modelsProduct.Channel{AllowUnpaid: false}); got != "unpaid" {
		t.Fatalf("prepay channel expected unpaid got %s", got)
	}
	if got := s.ResolvePaymentStatus(&modelsProduct.Channel{AllowUnpaid: true, Type: "b2b"}); got != "partial" {
		t.Fatalf("b2b NET channel expected partial got %s", got)
	}
}
