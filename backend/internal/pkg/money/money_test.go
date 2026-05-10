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
}
