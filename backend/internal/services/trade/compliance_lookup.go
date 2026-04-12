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

// LookupCompliance retrieves country-specific compliance references from local corpus.
func (s *AIService) LookupCompliance(country, query string, topK int) *ComplianceLookupResult {
	if s == nil || s.complianceRetriever == nil {
		return nil
	}

	normalizedQuery := strings.TrimSpace(query)
	if normalizedQuery == "" {
		normalizedQuery = "food labeling allergens additives import registration certifications packaging traceability"
	}

	matches := s.complianceRetriever.Search(country, normalizedQuery, topK)
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
		Answer:     s.complianceRetriever.FormatAnswer(country, normalizedQuery, matches),
		References: references,
	}
}
