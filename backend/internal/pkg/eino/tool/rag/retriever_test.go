package rag

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewComplianceRetrieverLoadsCorpus(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "fda.md"), `
# FDA FSMA
FSMA requires preventive controls and importer compliance plans.
`)
	mustWriteFile(t, filepath.Join(dir, "cfia.md"), `
# CFIA Labeling
Canada requires bilingual labeling in English and French.
`)

	retriever, err := NewComplianceRetriever(dir)
	if err != nil {
		t.Fatalf("NewComplianceRetriever() error = %v", err)
	}
	if len(retriever.chunks) == 0 {
		t.Fatalf("expected loaded chunks, got 0")
	}
}

func TestSearchReturnsRelevantSource(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "fda_fsma_overview.md"), `
# FSMA
The FDA Food Safety Modernization Act requires preventive controls for human food.
`)
	mustWriteFile(t, filepath.Join(dir, "halal.md"), `
# Saudi Halal
Saudi Arabia requires halal compliance and SFDA alignment.
`)

	retriever, err := NewComplianceRetriever(dir)
	if err != nil {
		t.Fatalf("NewComplianceRetriever() error = %v", err)
	}

	matches := retriever.Search("usa", "FSMA preventive controls", 3)
	if len(matches) == 0 {
		t.Fatalf("expected matches, got 0")
	}

	found := false
	for _, m := range matches {
		if m.Source == "fda_fsma_overview.md" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find source fda_fsma_overview.md, got %+v", matches)
	}
}

func TestSearchPrefersOfficialCountryEvidence(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "fssai_imports_and_labeling_2026.md"), `
# FSSAI Imports
India import clearance uses FICS and links with customs processes.
`)
	mustWriteFile(t, filepath.Join(dir, "notes.md"), `
# Notes
India labeling note from non-official commentary text.
`)

	retriever, err := NewComplianceRetriever(dir)
	if err != nil {
		t.Fatalf("NewComplianceRetriever() error = %v", err)
	}

	matches := retriever.Search("india", "import labeling fics", 3)
	if len(matches) == 0 {
		t.Fatalf("expected matches, got 0")
	}
	if !matches[0].Official {
		t.Fatalf("expected first match to be official for india, got %+v", matches[0])
	}
	if matches[0].Source != "fssai_imports_and_labeling_2026.md" {
		t.Fatalf("expected first official match from fssai corpus, got %s", matches[0].Source)
	}
}

func TestCanonicalCountryAlias(t *testing.T) {
	if got := canonicalCountry("United States"); got != "usa" {
		t.Fatalf("canonicalCountry alias mismatch: got %q, want %q", got, "usa")
	}
	if got := canonicalCountry("  KSA "); got != "saudi arabia" {
		t.Fatalf("canonicalCountry alias mismatch: got %q, want %q", got, "saudi arabia")
	}
}

func TestNewComplianceToolNilRetriever(t *testing.T) {
	tool, err := NewComplianceTool(nil)
	if err == nil {
		t.Fatalf("expected error for nil retriever, got nil")
	}
	if tool != nil {
		t.Fatalf("expected nil tool when retriever is nil")
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write test file %s: %v", path, err)
	}
}
