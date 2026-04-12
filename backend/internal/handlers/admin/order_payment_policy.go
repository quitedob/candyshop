package admin

import "strings"

func requiresFullPrepaymentCountry(country string) bool {
	switch normalizePaymentPolicyCountry(country) {
	case "india", "pakistan":
		return true
	default:
		return false
	}
}

func normalizePaymentPolicyCountry(country string) string {
	normalized := strings.ToLower(strings.TrimSpace(country))
	switch normalized {
	case "in", "india", "bharat":
		return "india"
	case "pk", "pakistan", "islamic republic of pakistan":
		return "pakistan"
	default:
		return normalized
	}
}

// requiresPrepaymentByTerms returns true when the company payment terms require
// payment before the order can move to execution stages.
// PREPAID and COD require payment upfront; NET_15/30/60 allow credit.
func requiresPrepaymentByTerms(paymentTerms string) bool {
	term := strings.ToUpper(strings.TrimSpace(paymentTerms))
	switch term {
	case "PREPAID", "COD":
		return true
	default:
		return false
	}
}

func isPaidInFull(paymentStatus string) bool {
	return strings.EqualFold(strings.TrimSpace(paymentStatus), "paid")
}

func requiresPaidBeforeExecution(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "confirmed", "production", "shipped", "delivered":
		return true
	default:
		return false
	}
}
