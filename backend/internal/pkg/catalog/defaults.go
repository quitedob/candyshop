package catalog

import "strings"

// productBasePriceDefaults maps product slug → catalog list price used as COGS scaling reference.
var productBasePriceDefaults = map[string]float64{
	"4d-fruit-gummy":             8.50,
	"crystal-hard-candy":         5.10,
	"rainbow-lollipop":           3.80,
	"sour-belt":                  4.20,
	"jelly-fruits":               6.00,
	"chewy-toffee":               7.50,
	"gummy-bear-classic":         4.95,
	"gummy-worms":                5.25,
	"fruit-slices":               4.50,
	"hard-candy-assorted-mix":    3.60,
	"hard-candy-fruit-bonbon":    4.80,
	"chocolate-premium-truffles": 12.00,
	"chocolate-variety-box":      15.00,
	"licorice-black-twists":      3.90,
	"licorice-fruit-ropes":       4.10,
	"sour-gummy-extreme":         5.50,
	"sour-belt-rainbow":          4.30,
	"sour-gummy-bears-zing":      5.00,
}

// CatalogProductSlugs returns seed catalog product slugs (stable order for tests/migrations).
func CatalogProductSlugs() []string {
	return []string{
		"4d-fruit-gummy",
		"crystal-hard-candy",
		"rainbow-lollipop",
		"sour-belt",
		"jelly-fruits",
		"chewy-toffee",
		"gummy-bear-classic",
		"gummy-worms",
		"fruit-slices",
		"hard-candy-assorted-mix",
		"hard-candy-fruit-bonbon",
		"chocolate-premium-truffles",
		"chocolate-variety-box",
		"licorice-black-twists",
		"licorice-fruit-ropes",
		"sour-gummy-extreme",
		"sour-belt-rainbow",
		"sour-gummy-bears-zing",
	}
}

// CatalogBasePrice returns the catalog reference price for COGS scaling by product slug.
func CatalogBasePrice(slug string) (float64, bool) {
	price, ok := productBasePriceDefaults[strings.TrimSpace(slug)]
	return price, ok && price > 0
}

// COGSReferencePrice resolves catalog reference price for COGS unit-cost scaling.
func COGSReferencePrice(slug string, basePrice, weightedAvgCost float64) float64 {
	if ref, ok := CatalogBasePrice(slug); ok {
		return ref
	}
	if basePrice > 0 && (weightedAvgCost <= 0 || basePrice > weightedAvgCost) {
		return basePrice
	}
	return 0
}

// EachProductBasePriceDefault invokes fn for every catalog slug → base price pair.
func EachProductBasePriceDefault(fn func(slug string, price float64)) {
	for slug, price := range productBasePriceDefaults {
		fn(slug, price)
	}
}
