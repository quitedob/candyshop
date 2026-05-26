package catalog

import "testing"

func TestCOGSReferencePrice_UsesCatalogSlug(t *testing.T) {
	got := COGSReferencePrice("4d-fruit-gummy", 0.99, 3.40)
	if got != 8.50 {
		t.Fatalf("COGSReferencePrice = %v, want 8.50", got)
	}
}

func TestCOGSReferencePrice_FallbackBasePrice(t *testing.T) {
	got := COGSReferencePrice("custom-oem", 12.0, 4.0)
	if got != 12.0 {
		t.Fatalf("COGSReferencePrice = %v, want 12.0", got)
	}
}

func TestCOGSReferencePrice_RejectsPollutedBase(t *testing.T) {
	got := COGSReferencePrice("custom-oem", 0.99, 3.40)
	if got != 0 {
		t.Fatalf("COGSReferencePrice = %v, want 0 for polluted base", got)
	}
}
