package country

import "testing"

func TestNormalizeCountryCode(t *testing.T) {
	cases := map[string]string{
		"United States": "US",
		"usa":           "US",
		"US":            "US",
		"Germany":       "DE",
		"CN":            "CN",
		"":              "",
	}
	for in, want := range cases {
		if got := NormalizeCountryCode(in); got != want {
			t.Fatalf("NormalizeCountryCode(%q) = %q, want %q", in, got, want)
		}
	}
}
