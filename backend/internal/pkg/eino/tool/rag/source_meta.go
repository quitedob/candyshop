package rag

import (
	"net/url"
	"strings"
)

type sourceMetadata struct {
	URL      string
	Official bool
}

var sourceMetadataByFile = map[string]sourceMetadata{
	"fda_fsma_overview.md":               {URL: "https://www.fda.gov/food/food-safety-modernization-act-fsma", Official: true},
	"cfia_bilingual_labeling.md":         {URL: "https://inspection.canada.ca/en/food-labels/labelling/industry/labelling-tool", Official: true},
	"fsanz_sweeteners.md":                {URL: "https://www.foodstandards.gov.au", Official: true},
	"gacc_imported_food_registration.md": {URL: "http://www.customs.gov.cn", Official: true},
	"global_regulatory_summary.md":       {URL: "", Official: false},
	"fssai_imports_and_labeling_2026.md": {URL: "https://fssai.gov.in/cms/imports.php", Official: true},
	"psqca_mandatory_and_halal_2026.md":  {URL: "https://www.psqca.com.pk/Notification/SRO_108_ITEMS.htm", Official: true},
	"pakistan_halal_authority_2026.md":   {URL: "https://www.pakistanhalalauthority.gov.pk/", Official: true},
}

var officialDomainsByCountry = map[string][]string{
	"india":       {"fssai.gov.in", "fics.fssai.gov.in"},
	"pakistan":    {"psqca.com.pk", "pakistanhalalauthority.gov.pk"},
	"usa":         {"fda.gov"},
	"canada":      {"inspection.canada.ca", "canada.ca"},
	"china":       {"customs.gov.cn"},
	"australia":   {"foodstandards.gov.au"},
	"new zealand": {"foodstandards.gov.au"},
}

func resolveSourceMetadata(source string) sourceMetadata {
	key := strings.TrimSpace(strings.ToLower(source))
	if key == "" {
		return sourceMetadata{}
	}
	if meta, ok := sourceMetadataByFile[key]; ok {
		return meta
	}
	return sourceMetadata{}
}

func sourceDomain(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || parsed.Host == "" {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Host))
	host = strings.TrimPrefix(host, "www.")
	return host
}

func officialDomainsForCountry(country string) []string {
	canonical := canonicalCountry(country)
	raw := officialDomainsByCountry[canonical]
	if len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, domain := range raw {
		trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(domain)), "www.")
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
}

func domainAllowed(domain string, allowlist []string) bool {
	if strings.TrimSpace(domain) == "" {
		return false
	}
	normalized := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(domain)), "www.")
	for _, allowed := range allowlist {
		if normalized == allowed {
			return true
		}
		if strings.HasSuffix(normalized, "."+allowed) {
			return true
		}
	}
	return false
}

func isOfficialMatchForCountry(chunk indexedChunk, country string) bool {
	if !chunk.IsOfficial {
		return false
	}
	allowlist := officialDomainsForCountry(country)
	if len(allowlist) == 0 {
		return true
	}
	return domainAllowed(chunk.SourceDomain, allowlist)
}

func hasOfficialMatches(matches []ComplianceMatch) bool {
	for _, match := range matches {
		if match.Official {
			return true
		}
	}
	return false
}
