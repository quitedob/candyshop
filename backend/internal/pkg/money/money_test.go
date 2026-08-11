package money

import (
	"math"
	"testing"
)

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

func TestMoneyToCentsInt(t *testing.T) {
	cases := []struct {
		name string
		amt  float64
		want int64
	}{
		{"clean value", 19.99, 1999},
		{"round up at third decimal", 19.999, 2000},
		{"sub-cent rounds down", 19.994, 1999},
		{"sub-cent rounds up", 19.9951, 2000},
		{"negative", -19.99, -1999},
		{"negative rounds away", -19.999, -2000},
		{"zero", 0, 0},
		{"matches RoundMoney semantics", 6695.8125, 669581},
		{"nan guards to zero", math.NaN(), 0},
		{"inf guards to zero", math.Inf(1), 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := MoneyToCentsInt(tc.amt); got != tc.want {
				t.Fatalf("MoneyToCentsInt(%v) = %d, want %d", tc.amt, got, tc.want)
			}
		})
	}
}
