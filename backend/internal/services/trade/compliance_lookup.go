// Package trade 的合规检索（LookupCompliance）仅作 RAG 解释层与内控参考，不构成法律意见；
// 订单/产品否决仍以硬编码规则与 ProductMarketProfile 为准。
package trade

import "strings"

// ComplianceReference is one compliance evidence snippet from local corpus retrieval.
type ComplianceReference struct {
	Source   string  `json:"source"`
	Section  string  `json:"section"`
	Snippet  string  `json:"snippet"`
	Score    float64 `json:"score"`
	URL      string  `json:"url,omitempty"`
	Official bool    `json:"official"`
}

// ComplianceLookupResult contains compliance summary and references.
type ComplianceLookupResult struct {
	Answer     string                `json:"answer"`
	References []ComplianceReference `json:"references"`
}

// LookupCompliance 从本地语料检索合规参考片段（非法律依据，不得单独作为放行条件）。
func (s *AIService) LookupCompliance(country, query string, topK int) *ComplianceLookupResult {
	if s == nil || s.ComplianceRetriever() == nil {
		return nil
	}

	normalizedQuery := strings.TrimSpace(query)
	if normalizedQuery == "" {
		normalizedQuery = "food labeling allergens additives import registration certifications packaging traceability"
	}

	matches := s.ComplianceRetriever().Search(country, normalizedQuery, topK)
	references := make([]ComplianceReference, 0, len(matches))
	for _, match := range matches {
		references = append(references, ComplianceReference{
			Source:   match.Source,
			Section:  match.Section,
			Snippet:  match.Snippet,
			Score:    match.Score,
			URL:      match.URL,
			Official: match.Official,
		})
	}

	return &ComplianceLookupResult{
		Answer:     s.ComplianceRetriever().FormatAnswer(country, normalizedQuery, matches),
		References: references,
	}
}
