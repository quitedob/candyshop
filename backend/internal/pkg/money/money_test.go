package money

import "testing"

func TestMoneyCoversTotal(t *testing.T) {
	if !MoneyCoversTotal(100.0, 100.0) {
		t.Fatal("exact match")
	}
	if !MoneyCoversTotal(99.999999, 100.0) {
		t.Fatal("within epsilon")
	}
	if MoneyCoversTotal(99.0, 100.0) {
		t.Fatal("below should not cover")
	}
	// API audit: payment 6695.81 vs order total 6695.8125 after tax calc
	if !MoneyCoversTotal(6695.81, 6695.8125) {
		t.Fatal("2-decimal payment should cover 4-decimal order total")
	}
}

func TestRoundMoney(t *testing.T) {
	if got := RoundMoney(6695.8125); got != 6695.81 {
		t.Fatalf("RoundMoney = %v, want 6695.81", got)
	}
}
