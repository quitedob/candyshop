package kyb

import (
	"math"
	"testing"
)

func TestUsdCapValue(t *testing.T) {
	rates := map[string]float64{
		"EUR": 0.8, // 1 USD = 0.8 EUR → 2 EUR = 2.5 USD
		"CNY": 7.2,
	}
	cases := []struct {
		name     string
		rates    map[string]float64
		currency string
		amount   float64
		want     float64
		wantInf  bool
	}{
		{"empty currency is USD", rates, "", 100, 100, false},
		{"USD passthrough", rates, "USD", 100, 100, false},
		{"lowercase usd passthrough", rates, "usd", 100, 100, false},
		{"known rate converts", rates, "EUR", 2, 2.5, false},
		{"known rate CNY", rates, "CNY", 720, 100, false},
		{"unknown currency fails closed", rates, "XYZ", 100, 0, true},
		{"zero rate fails closed", map[string]float64{"EUR": 0}, "EUR", 100, 0, true},
		{"negative rate fails closed", map[string]float64{"EUR": -0.5}, "EUR", 100, 0, true},
		{"nil rates fails closed", nil, "EUR", 100, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := UsdCapValue(c.rates, c.currency, c.amount)
			if c.wantInf {
				if !math.IsInf(got, 1) {
					t.Fatalf("expected +Inf, got %v", got)
				}
				// +Inf must never satisfy a finite cap.
				if got < 1000 {
					t.Fatalf("fail-closed value %v must exceed any finite cap", got)
				}
				return
			}
			if got != c.want {
				t.Fatalf("UsdCapValue(%v, %q, %v) = %v, want %v", c.rates, c.currency, c.amount, got, c.want)
			}
		})
	}
}
