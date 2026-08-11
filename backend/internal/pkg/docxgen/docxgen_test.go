package docxgen

import (
	"math"
	"testing"
)

// TestFormatMoney is the G26 regression: FormatMoney must round to the nearest
// cent before formatting instead of truncating (19.999 used to render as
// "USD 19.99", off by one cent).
func TestFormatMoney(t *testing.T) {
	cases := []struct {
		name string
		amt  float64
		cur  string
		want string
	}{
		{"clean value", 19.99, "USD", "USD 19.99"},
		{"round up off-by-one-cent", 19.999, "USD", "USD 20.00"},
		{"sub-cent rounds down", 19.994, "USD", "USD 19.99"},
		{"sub-cent rounds up", 19.9951, "USD", "USD 20.00"},
		{"thousands separator", 1234.5, "USD", "USD 1,234.50"},
		{"millions separator", 1000000.0, "USD", "USD 1,000,000.00"},
		{"negative rounds", -19.999, "USD", "USD -20.00"},
		{"zero", 0, "USD", "USD 0.00"},
		{"nan guard", math.NaN(), "USD", "USD 0.00"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatMoney(tc.amt, tc.cur); got != tc.want {
				t.Fatalf("FormatMoney(%v, %q) = %q, want %q", tc.amt, tc.cur, got, tc.want)
			}
		})
	}
}
