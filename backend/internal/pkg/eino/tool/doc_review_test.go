package tool

import "testing"

func TestGenerateTradeDocumentsRequiresReview(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		args     string
		want     bool
	}{
		{
			name:     "non-matching tool is exempt",
			toolName: "track_shipment",
			args:     `{"total_amount":99999}`,
			want:     false,
		},
		{
			name:     "high-value doc generation is gated",
			toolName: "generate_trade_documents",
			args:     `{"trade_id":1,"total_amount":10000}`,
			want:     true,
		},
		{
			name:     "above threshold is gated",
			toolName: "generate_trade_documents",
			args:     `{"total_amount":25000.5}`,
			want:     true,
		},
		{
			name:     "below threshold passes through",
			toolName: "generate_trade_documents",
			args:     `{"total_amount":9999.99}`,
			want:     false,
		},
		{
			name:     "omitted total_amount passes through",
			toolName: "generate_trade_documents",
			args:     `{"trade_id":1,"doc_types":["COMMERCIAL_INVOICE"]}`,
			want:     false,
		},
		{
			name:     "malformed arguments pass through",
			toolName: "generate_trade_documents",
			args:     `{not json}`,
			want:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := GenerateTradeDocumentsRequiresReview(tc.toolName, tc.args); got != tc.want {
				t.Fatalf("GenerateTradeDocumentsRequiresReview(%q, %q) = %v, want %v",
					tc.toolName, tc.args, got, tc.want)
			}
		})
	}
}
