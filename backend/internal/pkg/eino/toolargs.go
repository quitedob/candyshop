package eino

import (
	"encoding/json"
	"strings"
)

// FallbackTradeDocumentType is used as the SSE document_type anchor when a
// generate_trade_documents tool call's doc_types argument is absent, empty, or
// malformed. Shared by the admin and system SSE event processors so both
// fall back identically.
const FallbackTradeDocumentType = "TRADE_DOCUMENT"

// FirstDocTypeFromArgs returns the first non-empty doc_types entry (trimmed)
// from a generate_trade_documents tool call's arguments JSON, or "" when the
// arguments do not parse or the array is empty. Callers fall back to
// FallbackTradeDocumentType on "".
func FirstDocTypeFromArgs(argsJSON string) string {
	if strings.TrimSpace(argsJSON) == "" {
		return ""
	}
	var args struct {
		DocTypes []string `json:"doc_types"`
	}
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ""
	}
	for _, docType := range args.DocTypes {
		if strings.TrimSpace(docType) != "" {
			return strings.TrimSpace(docType)
		}
	}
	return ""
}
