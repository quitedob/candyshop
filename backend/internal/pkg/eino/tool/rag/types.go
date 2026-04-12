package rag

import "sync"

// ComplianceQueryInput is the argument schema for the compliance retrieval tool.
type ComplianceQueryInput struct {
	Country string `json:"country" jsonschema_description:"Target country or region, for example USA, Saudi Arabia, Indonesia, or EU."`
	Query   string `json:"query" jsonschema_description:"Compliance query keywords, for example labeling, certifications, additives, packaging."`
	TopK    int    `json:"top_k,omitempty" jsonschema_description:"Number of top references to return. Allowed range: 1-8."`
}

// ComplianceMatch is one retrieval hit from markdown corpus.
type ComplianceMatch struct {
	Source   string  `json:"source"`
	Section  string  `json:"section"`
	Snippet  string  `json:"snippet"`
	Score    float64 `json:"score"`
	URL      string  `json:"url,omitempty"`
	Official bool    `json:"official"`
}

// ComplianceQueryOutput is the tool response payload.
type ComplianceQueryOutput struct {
	Answer                string            `json:"answer"`
	Matches               []ComplianceMatch `json:"matches"`
	HasOfficialEvidence   bool              `json:"hasOfficialEvidence"`
	OfficialSourceDomains []string          `json:"officialSourceDomains,omitempty"`
}

type indexedChunk struct {
	Source            string
	Section           string
	Text              string
	NormalizedText    string
	NormalizedSection string
	TokenSet          map[string]struct{}
	SourceURL         string
	SourceDomain      string
	IsOfficial        bool
}

// ComplianceRetriever loads markdown corpus and retrieves relevant sections.
type ComplianceRetriever struct {
	corpusDir string
	options   Options
	mu        sync.RWMutex
	chunks    []indexedChunk
}
