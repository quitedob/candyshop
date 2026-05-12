package rag

import (
	"fmt"
	"os"
	"path/filepath"
)

// NewFromCorpus resolves the compliance corpus directory and creates a ComplianceRetriever.
// This is the shared factory used by AIService and TradeAgent to avoid code duplication.
func NewFromCorpus() (*ComplianceRetriever, error) {
	corpusDir, err := resolveCorpusDir()
	if err != nil {
		return nil, err
	}
	return NewComplianceRetriever(corpusDir)
}

func resolveCorpusDir() (string, error) {
	candidates := []string{
		filepath.Join("internal", "pkg", "eino", "corpus"),
		filepath.Join(".", "internal", "pkg", "eino", "corpus"),
		filepath.Join("backend", "internal", "pkg", "eino", "corpus"),
	}
	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("cannot find compliance corpus directory in known locations")
}
