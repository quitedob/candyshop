package database

import (
	modelsProduct "candypro/api/internal/models/product"
	"log"

	"gorm.io/gorm"
)

// productGrossWeightDefaults slug → kg per unit (used in checkout weight estimate: qty × weight).
var productGrossWeightDefaults = map[string]float64{
	"4d-fruit-gummy":             0.003,
	"crystal-hard-candy":         0.002,
	"rainbow-lollipop":           0.005,
	"sour-belt":                  0.003,
	"jelly-fruits":               0.004,
	"chewy-toffee":               0.004,
	"gummy-bear-classic":         0.003,
	"gummy-worms":                0.003,
	"fruit-slices":               0.003,
	"hard-candy-assorted-mix":    0.002,
	"hard-candy-fruit-bonbon":    0.003,
	"chocolate-premium-truffles": 0.008,
	"chocolate-variety-box":      0.010,
	"licorice-black-twists":      0.004,
	"licorice-fruit-ropes":       0.004,
	"sour-gummy-extreme":         0.003,
	"sour-belt-rainbow":          0.003,
	"sour-gummy-bears-zing":      0.003,
}

// catalogProductSlugs 种子目录产品 slug 列表（与 seedProducts 保持一致）
var catalogProductSlugs = []string{
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

// EnsureProductGrossWeights 启动时回填缺失的 gross_weight_per_carton，供运费估算
func EnsureProductGrossWeights(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var updated int64
	for slug, weight := range productGrossWeightDefaults {
		res := db.Model(&modelsProduct.Product{}).
			Where("slug = ? AND (gross_weight_per_carton = 0 OR gross_weight_per_carton IS NULL)", slug).
			Update("gross_weight_per_carton", weight)
		if res.Error != nil {
			return res.Error
		}
		updated += res.RowsAffected
	}
	if updated > 0 {
		log.Printf("Ensured gross_weight_per_carton for %d product(s)", updated)
	}
	return nil
}
