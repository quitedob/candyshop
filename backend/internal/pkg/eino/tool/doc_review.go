package tool

import "encoding/json"

// GenerateTradeDocumentsReviewThreshold is the total_amount floor (in the
// transaction's currency) at or above which a generate_trade_documents call must
// pass through the human review-and-edit gate before its PersistDocs node writes
// TradeDocument rows. generate_trade_documents is the only state-changing tool
// without its own DB approval queue (submit_quotation_for_human_review already
// queues into the QuotationReview list), so it is the single tool the HITL gate
// can meaningfully protect.
const GenerateTradeDocumentsReviewThreshold = 10000

// GenerateTradeDocumentsRequiresReview returns a ReviewPolicyFunc that gates only
// generate_trade_documents calls whose total_amount is at least the threshold.
// Every other tool — including the read-only reviewExemptTools — passes through
// immediately. A call whose arguments do not parse, or which omit total_amount,
// is not gated: a malformed/partial call must not deadlock behind a review it
// can never satisfy.
func GenerateTradeDocumentsRequiresReview(toolName, argumentsInJSON string) bool {
	if toolName != "generate_trade_documents" {
		return false
	}
	var args struct {
		TotalAmount float64 `json:"total_amount"`
	}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return false
	}
	return args.TotalAmount >= GenerateTradeDocumentsReviewThreshold
}
