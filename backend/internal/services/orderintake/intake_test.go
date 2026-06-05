package orderintake

import "testing"

func TestParseQuantity(t *testing.T) {
	tests := map[string]int{
		"":                    1,
		"not specified":       1,
		"0 cartons":           1,
		"1,000 cartons":       1000,
		"about 2,500 kg":      2500,
		"estimated: 12.9 pcs": 12,
	}
	for input, want := range tests {
		if got := parseQuantity(input); got != want {
			t.Fatalf("parseQuantity(%q) = %d, want %d", input, got, want)
		}
	}
}
