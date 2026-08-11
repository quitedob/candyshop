package eino

import "testing"

func TestFirstDocTypeFromArgs(t *testing.T) {
	cases := []struct {
		name string
		args string
		want string
	}{
		{"valid first entry", `{"doc_types": ["COMMERCIAL_INVOICE", "PACKING_LIST"]}`, "COMMERCIAL_INVOICE"},
		{"single entry", `{"doc_types": ["PROFORMA_INVOICE"]}`, "PROFORMA_INVOICE"},
		{"empty array", `{"doc_types": []}`, ""},
		{"malformed json", `{"doc_types": [`, ""},
		{"blank string", "", ""},
		{"whitespace string", "   ", ""},
		{"non-array doc_types", `{"doc_types": "COMMERCIAL_INVOICE"}`, ""},
		{"leading blank entries skipped", `{"doc_types": ["", "PACKING_LIST"]}`, "PACKING_LIST"},
		{"entry with surrounding whitespace", `{"doc_types": ["  PACKING_LIST  "] }`, "PACKING_LIST"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := FirstDocTypeFromArgs(c.args); got != c.want {
				t.Fatalf("FirstDocTypeFromArgs(%q) = %q, want %q", c.args, got, c.want)
			}
		})
	}
}

func TestFallbackTradeDocumentType(t *testing.T) {
	if FallbackTradeDocumentType != "TRADE_DOCUMENT" {
		t.Fatalf("unexpected fallback sentinel %q", FallbackTradeDocumentType)
	}
}
