package rag

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

var stopWords = map[string]struct{}{
	"a": {}, "an": {}, "the": {}, "and": {}, "or": {}, "to": {}, "for": {}, "of": {}, "in": {}, "on": {},
	"by": {}, "at": {}, "from": {}, "with": {}, "is": {}, "are": {}, "be": {}, "as": {}, "that": {}, "this": {},
	"it": {}, "its": {}, "their": {}, "your": {}, "you": {}, "about": {}, "into": {}, "than": {}, "can": {}, "must": {},
}

// NewComplianceRetriever loads markdown corpus into an in-memory retriever.
func NewComplianceRetriever(corpusDir string, opts ...Option) (*ComplianceRetriever, error) {
	cfg := applyOptions(opts...)
	chunks, err := loadCorpus(corpusDir, cfg)
	if err != nil {
		return nil, err
	}

	return &ComplianceRetriever{
		corpusDir: corpusDir,
		options:   cfg,
		chunks:    chunks,
	}, nil
}

// Search returns top-k matches based on country and query text.
func (r *ComplianceRetriever) Search(country, query string, topK int) []ComplianceMatch {
	r.mu.RLock()
	chunks := make([]indexedChunk, len(r.chunks))
	copy(chunks, r.chunks)
	r.mu.RUnlock()

	if len(chunks) == 0 {
		return nil
	}

	canonical := canonicalCountry(country)
	queryText := strings.TrimSpace(strings.Join([]string{canonical, query}, " "))
	queryTokens := tokenize(queryText)
	if len(queryTokens) == 0 {
		queryTokens = tokenize(query)
	}
	if len(queryTokens) == 0 {
		return nil
	}

	k := topK
	if k <= 0 {
		k = r.options.DefaultTopK
	}
	if k > r.options.MaxTopK {
		k = r.options.MaxTopK
	}

	matches := r.searchWithThreshold(chunks, queryTokens, canonical, r.options.MinScore)
	if len(matches) == 0 {
		matches = r.searchWithThreshold(chunks, queryTokens, canonical, 0)
	}
	if len(matches) == 0 {
		return nil
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Official != matches[j].Official {
			return matches[i].Official
		}
		if matches[i].Score == matches[j].Score {
			if matches[i].Source == matches[j].Source {
				return matches[i].Section < matches[j].Section
			}
			return matches[i].Source < matches[j].Source
		}
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > k {
		matches = matches[:k]
	}
	return matches
}

// FormatAnswer converts retrieval hits into a concise answer for model/tool output.
func (r *ComplianceRetriever) FormatAnswer(country, query string, matches []ComplianceMatch) string {
	if len(matches) == 0 {
		return "No exact match found in local compliance markdown corpus. Please verify with official regulatory websites."
	}

	target := strings.TrimSpace(country)
	if target == "" {
		target = "target market"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Compliance references for %s", target)
	if q := strings.TrimSpace(query); q != "" {
		fmt.Fprintf(&b, " (query: %s)", q)
	}
	b.WriteString(":\n")

	for i, m := range matches {
		marker := "reference"
		if m.Official {
			marker = "official"
		}
		if strings.TrimSpace(m.URL) != "" {
			fmt.Fprintf(&b, "%d. [%s | %s | %s] %s (%s)\n", i+1, m.Source, m.Section, marker, m.Snippet, m.URL)
			continue
		}
		fmt.Fprintf(&b, "%d. [%s | %s | %s] %s\n", i+1, m.Source, m.Section, marker, m.Snippet)
	}
	return strings.TrimSpace(b.String())
}

func (r *ComplianceRetriever) searchWithThreshold(chunks []indexedChunk, queryTokens []string, country string, threshold float64) []ComplianceMatch {
	matches := make([]ComplianceMatch, 0, len(chunks))
	for _, chunk := range chunks {
		official := isOfficialMatchForCountry(chunk, country)
		score := scoreChunk(chunk, queryTokens, country, official)
		if score <= threshold {
			continue
		}
		matches = append(matches, ComplianceMatch{
			Source:   chunk.Source,
			Section:  chunk.Section,
			Snippet:  buildSnippet(chunk.Text, queryTokens),
			Score:    score,
			URL:      chunk.SourceURL,
			Official: official,
		})
	}
	return matches
}

func scoreChunk(chunk indexedChunk, queryTokens []string, country string, official bool) float64 {
	if len(queryTokens) == 0 {
		return 0
	}

	matchCount := 0
	for _, token := range queryTokens {
		if _, ok := chunk.TokenSet[token]; ok {
			matchCount++
		}
	}
	if matchCount == 0 {
		return 0
	}

	score := float64(matchCount) / float64(len(queryTokens))
	if country != "" {
		if _, ok := chunk.TokenSet[country]; ok {
			score += 0.25
		}
		if strings.Contains(chunk.NormalizedSection, country) {
			score += 0.15
		}
		if strings.Contains(chunk.NormalizedText, country) {
			score += 0.10
		}
	}

	sectionBoost := 0.0
	for _, token := range queryTokens {
		if strings.Contains(chunk.NormalizedSection, token) {
			sectionBoost += 0.03
		}
	}
	if sectionBoost > 0.20 {
		sectionBoost = 0.20
	}
	if official {
		score += 0.20
	}
	return score + sectionBoost
}

func buildSnippet(text string, queryTokens []string) string {
	const maxRunes = 260

	cleanText := strings.TrimSpace(text)
	if cleanText == "" {
		return cleanText
	}

	runes := []rune(cleanText)
	if len(runes) <= maxRunes {
		return cleanText
	}

	lower := strings.ToLower(cleanText)
	tokenRunePos := -1
	for _, token := range queryTokens {
		if token == "" {
			continue
		}
		if idx := strings.Index(lower, token); idx >= 0 {
			tokenRunePos = utf8.RuneCountInString(lower[:idx])
			break
		}
	}

	if tokenRunePos < 0 {
		return strings.TrimSpace(string(runes[:maxRunes])) + "..."
	}

	start := tokenRunePos - (maxRunes / 3)
	if start < 0 {
		start = 0
	}
	end := start + maxRunes
	if end > len(runes) {
		end = len(runes)
		start = end - maxRunes
		if start < 0 {
			start = 0
		}
	}

	snippet := strings.TrimSpace(string(runes[start:end]))
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet += "..."
	}
	return snippet
}

func normalizeForSearch(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(' ')
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func tokenize(text string) []string {
	normalized := normalizeForSearch(text)
	if normalized == "" {
		return nil
	}

	parts := strings.Fields(normalized)
	tokens := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, token := range parts {
		if len([]rune(token)) < 2 {
			continue
		}
		if _, stop := stopWords[token]; stop {
			continue
		}
		if _, ok := seen[token]; ok {
			continue
		}
		seen[token] = struct{}{}
		tokens = append(tokens, token)
	}
	return tokens
}

func toSet(tokens []string) map[string]struct{} {
	set := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		set[token] = struct{}{}
	}
	return set
}
