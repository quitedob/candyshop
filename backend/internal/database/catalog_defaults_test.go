package database

import (
	"testing"

	"candypro/api/internal/pkg/catalog"
)

func TestCatalogProductDefaultsCoverage(t *testing.T) {
	for _, slug := range catalog.CatalogProductSlugs() {
		if _, ok := catalog.CatalogBasePrice(slug); !ok {
			t.Fatalf("catalog.CatalogBasePrice missing slug %q", slug)
		}
		if _, ok := productGrossWeightDefaults[slug]; !ok {
			t.Fatalf("productGrossWeightDefaults missing slug %q", slug)
		}
		if _, ok := defaultProductWeightedAvgCosts[slug]; !ok {
			t.Fatalf("defaultProductWeightedAvgCosts missing slug %q", slug)
		}
	}
	catalog.EachProductBasePriceDefault(func(slug string, _ float64) {
		if !containsSlug(catalog.CatalogProductSlugs(), slug) {
			t.Fatalf("catalog base price defaults has unknown slug %q", slug)
		}
	})
}

func containsSlug(slugs []string, target string) bool {
	for _, s := range slugs {
		if s == target {
			return true
		}
	}
	return false
}
