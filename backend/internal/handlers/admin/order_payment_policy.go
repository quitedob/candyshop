package admin

import (
	"context"
	"strings"
)

// requiresFullPrepaymentCountry checks if the country requires full prepayment.
// Uses the database-driven policy service when available, falls back to hardcoded defaults.
func (h *Handler) requiresFullPrepaymentCountry(country string) bool {
	if h.countryPaymentPolicy != nil {
		return h.countryPaymentPolicy.RequiresFullPrepayment(country)
	}
	// Fallback for when the policy service is unavailable
	switch normalizePaymentPolicyCountry(country) {
	case "india", "pakistan":
		return true
	default:
		return false
	}
}

// requiresFullPrepaymentForOrder checks both the destination country policy and
// the user's company payment terms to determine if full prepayment is required.
func (h *Handler) requiresFullPrepaymentForOrder(ctx context.Context, country, userID string) bool {
	if h.requiresFullPrepaymentCountry(country) {
		return true
	}
	if h.services != nil && h.services.User != nil && h.services.Company != nil && userID != "" {
		usr, err := h.services.User.GetByID(ctx, userID)
		if err == nil && usr.CompanyID != nil {
			company, err := h.services.Company.GetCompany(ctx, *usr.CompanyID)
			if err == nil {
				return requiresPrepaymentByTerms(company.PaymentTerms)
			}
		}
	}
	return false
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

// isPaidOrPartial 已付款或部分付款视为满足执行阶段付款要求
func isPaidOrPartial(paymentStatus string) bool {
	switch strings.ToLower(strings.TrimSpace(paymentStatus)) {
	case "paid", "partial":
		return true
	default:
		return false
	}
}

// paymentInsufficientForExecution 进入 confirmed/production/shipped/delivered 前须 paid 或 partial
func paymentInsufficientForExecution(targetStatus, paymentStatus string) bool {
	if !requiresPaidBeforeExecution(targetStatus) {
		return false
	}
	return !isPaidOrPartial(paymentStatus)
}

func requiresPaidBeforeExecution(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "confirmed", "production", "shipped", "delivered":
		return true
	default:
		return false
	}
}
