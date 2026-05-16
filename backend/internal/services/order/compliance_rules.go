// Package order 的目的国基线合规规则以本文件为运行时来源；变更请在 Git 留痕并与 ProductMarketProfile 版本字段对齐治理。
package order

import (
	modelsProduct "candypro/api/internal/models/product"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// PaymentPolicyRule represents country-level payment policy constraints.
type PaymentPolicyRule struct {
	RequiresFullPrepayment bool     `json:"requiresFullPrepayment"`
	AllowedTerms           []string `json:"allowedTerms,omitempty"`
	Note                   string   `json:"note,omitempty"`
}

// ComplianceValidationResult contains validation output for destination-country checks.
type ComplianceValidationResult struct {
	Country    string            `json:"country"`
	Warnings   []string          `json:"warnings"`
	Violations []string          `json:"violations"`
	Payment    PaymentPolicyRule `json:"paymentPolicy"`
}

// ValidateCountryCompliance validates product set against destination country baseline rules.
func (s *OrderService) ValidateCountryCompliance(country string, products []modelsProduct.Product) ComplianceValidationResult {
	return ValidateCountryComplianceRules(country, products)
}

// ValidateCountryComplianceWithProfiles 合并硬编码规则与 ProductMarketProfile 配置
func (s *OrderService) ValidateCountryComplianceWithProfiles(country string, products []modelsProduct.Product, profiles []modelsProduct.ProductMarketProfile) ComplianceValidationResult {
	return ValidateCountryComplianceRulesWithProfiles(country, products, profiles)
}

// ValidateCountryComplianceRules is the standalone version of the compliance validator.
// It can be called from any handler scope without needing an OrderService instance.
func ValidateCountryComplianceRules(country string, products []modelsProduct.Product) ComplianceValidationResult {
	return validateCountryCompliance(country, products)
}

// ValidateCountryComplianceRulesWithProfiles 基线规则 + 数据库市场画像
func ValidateCountryComplianceRulesWithProfiles(country string, products []modelsProduct.Product, profiles []modelsProduct.ProductMarketProfile) ComplianceValidationResult {
	r := validateCountryCompliance(country, products)
	mergeProductMarketProfiles(&r, canonicalComplianceCountry(country), products, profiles)
	return r
}

func validateCountryCompliance(country string, products []modelsProduct.Product) ComplianceValidationResult {
	canonical := canonicalComplianceCountry(country)
	result := ComplianceValidationResult{
		Country:    canonical,
		Warnings:   make([]string, 0, 8),
		Violations: make([]string, 0, 8),
		Payment:    PaymentPolicyRule{},
	}

	if len(products) == 0 {
		result.Violations = append(result.Violations, "No products provided for compliance validation.")
		return result
	}

	for _, p := range products {
		pid := strings.TrimSpace(p.ID)
		if pid == "" {
			pid = strings.TrimSpace(p.Name)
		}
		if pid == "" {
			pid = "unknown-product"
		}

		ingredients := strings.ToLower(strings.TrimSpace(p.Ingredients))
		allergens := strings.ToLower(strings.TrimSpace(p.Allergens))
		certText := strings.ToLower(strings.Join(p.Certifications, " "))

		switch canonical {
		case "eu":
			if containsAny(ingredients, "titanium dioxide", "e171") {
				result.Violations = append(result.Violations, fmt.Sprintf("Product %s includes titanium dioxide (E171), which is not authorized for EU food.", pid))
			}
			if ingredients == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty ingredient declaration. EU labeling dossier should include full composition.", pid))
			}
		case "saudi arabia":
			if containsAny(ingredients, "porcine", "pork", "lard", "ethanol", "alcohol", "wine", "rum", "brandy") {
				result.Violations = append(result.Violations, fmt.Sprintf("Product %s contains prohibited non-halal indicators for Saudi Arabia.", pid))
			}
			if !p.HalalCertified {
				result.Violations = append(result.Violations, fmt.Sprintf("Product %s is not marked as halal certified for Saudi market.", pid))
			}
			if !containsAny(certText, "halal") {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s does not explicitly list halal certificate in certifications field.", pid))
			}
		case "usa":
			if containsAny(ingredients, "cyclamate") {
				result.Violations = append(result.Violations, fmt.Sprintf("Product %s includes cyclamate, which is not permitted in US food products.", pid))
			}
			if ingredients == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty ingredient declaration. US labeling requires full ingredient statement.", pid))
			}
			if allergens == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty allergen declaration. US labeling requires major allergen statement when applicable.", pid))
			}
		case "japan":
			if ingredients == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty ingredient declaration. Japan import notification review usually requires additive/ingredient clarity.", pid))
			}
			if allergens == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty allergen declaration. Japan labeling should include allergen information where applicable.", pid))
			}
		case "india":
			result.Payment = PaymentPolicyRule{
				RequiresFullPrepayment: true,
				AllowedTerms:           []string{"100% T/T before production"},
				Note:                   "India orders follow stricter risk-control policy: full prepayment is required before production scheduling.",
			}
			if ingredients == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty ingredient declaration. India (FSSAI) import review requires complete ingredient and additive disclosure.", pid))
			}
			if allergens == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty allergen declaration. India labeling should include allergen information where applicable.", pid))
			}
		case "pakistan":
			result.Payment = PaymentPolicyRule{
				RequiresFullPrepayment: true,
				AllowedTerms:           []string{"100% T/T before production"},
				Note:                   "Pakistan orders follow stricter risk-control policy: full prepayment is required before production scheduling.",
			}
			if containsAny(ingredients, "porcine", "pork", "lard", "ethanol", "alcohol", "wine", "rum", "brandy") {
				result.Violations = append(result.Violations, fmt.Sprintf("Product %s contains prohibited non-halal indicators for Pakistan market.", pid))
			}
			if !p.HalalCertified {
				result.Violations = append(result.Violations, fmt.Sprintf("Product %s is not marked as halal certified for Pakistan market.", pid))
			}
			if !containsAny(certText, "halal") {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s does not explicitly list halal certificate in certifications field.", pid))
			}
			if ingredients == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty ingredient declaration. Pakistan import dossier should include full composition.", pid))
			}
		default:
			if ingredients == "" {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s has empty ingredient declaration. Destination-market compliance may fail without full composition.", pid))
			}
		}
	}

	if result.Payment.RequiresFullPrepayment {
		result.Warnings = append(result.Warnings, result.Payment.Note)
	}
	result.Warnings = dedupeLower(result.Warnings)
	result.Violations = dedupeLower(result.Violations)
	return result
}

// CanonicalComplianceCountry 将用户输入的目的国归一化为内部规则键（供画像 market 匹配）
func CanonicalComplianceCountry(country string) string {
	return canonicalComplianceCountry(country)
}

// ComplianceProfileMarketCode 将目的国映射为 ProductMarketProfile.market_code（EU/US/GCC…）
func ComplianceProfileMarketCode(country string) string {
	return profileMarketFromCanon(canonicalComplianceCountry(country))
}

func canonicalComplianceCountry(country string) string {
	normalized := strings.ToLower(strings.TrimSpace(country))
	switch normalized {
	case "us", "usa", "united states", "united states of america", "america":
		return "usa"
	case "eu", "europe", "european union", "germany", "france", "italy", "spain", "netherlands", "belgium", "poland":
		return "eu"
	case "saudi", "saudi arabia", "ksa":
		return "saudi arabia"
	case "jp", "japan", "nippon":
		return "japan"
	case "in", "india", "bharat":
		return "india"
	case "pk", "pakistan", "islamic republic of pakistan":
		return "pakistan"
	default:
		return normalized
	}
}

func profileMarketFromCanon(canon string) string {
	switch canon {
	case "eu":
		return "EU"
	case "usa":
		return "US"
	case "saudi arabia":
		return "GCC"
	default:
		return strings.ToUpper(strings.TrimSpace(canon))
	}
}

func mergeProductMarketProfiles(result *ComplianceValidationResult, canon string, products []modelsProduct.Product, profiles []modelsProduct.ProductMarketProfile) {
	if result == nil || len(profiles) == 0 {
		return
	}
	want := profileMarketFromCanon(canon)
	for _, p := range products {
		pid := strings.TrimSpace(p.ID)
		if pid == "" {
			pid = strings.TrimSpace(p.Name)
		}
		if pid == "" {
			pid = "unknown-product"
		}
		ingredients := strings.ToLower(strings.TrimSpace(p.Ingredients))
		certText := strings.ToLower(strings.Join(p.Certifications, " "))
		for _, prof := range profiles {
			if prof.ProductID != pid || !strings.EqualFold(strings.TrimSpace(prof.MarketCode), want) {
				continue
			}
			var blocked []string
			if ue := json.Unmarshal(prof.BlockedIngredientPatterns, &blocked); ue != nil {
				log.Printf("compliance_rules: unmarshal BlockedIngredientPatterns failed for product %s: %v", prof.ProductID, ue)
			}
			for _, pat := range blocked {
				pt := strings.ToLower(strings.TrimSpace(pat))
				if pt != "" && strings.Contains(ingredients, pt) {
					result.Violations = append(result.Violations, fmt.Sprintf("Product %s violates market profile ban on '%s' for %s.", pid, pat, want))
				}
			}
			var req []string
			if ue := json.Unmarshal(prof.RequiredCertKeywords, &req); ue != nil {
				log.Printf("compliance_rules: unmarshal RequiredCertKeywords failed for product %s: %v", prof.ProductID, ue)
			}
			for _, kw := range req {
				k := strings.ToLower(strings.TrimSpace(kw))
				if k != "" && !strings.Contains(certText, k) {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Product %s market profile expects certification keyword '%s'.", pid, kw))
				}
			}
		}
	}
	result.Violations = dedupeLower(result.Violations)
	result.Warnings = dedupeLower(result.Warnings)
}

func containsAny(text string, tokens ...string) bool {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return false
	}
	for _, t := range tokens {
		token := strings.ToLower(strings.TrimSpace(t))
		if token == "" {
			continue
		}
		if strings.Contains(text, token) {
			return true
		}
	}
	return false
}

func dedupeLower(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}
