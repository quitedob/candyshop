package database

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/catalog"
	"log"

	"gorm.io/gorm"
)

// EnsureProductBasePrices syncs catalog reference base_price for known product slugs.
// Only unset prices (base_price = 0 / NULL) are filled, matching the in-seed path, so
// admin edits to seeded products' base_price are not silently reverted on restart.
func EnsureProductBasePrices(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	var updated int64
	catalog.EachProductBasePriceDefault(func(slug string, price float64) {
		res := db.Model(&modelsProduct.Product{}).
			Where("slug = ? AND (base_price = 0 OR base_price IS NULL)", slug).
			Update("base_price", price)
		if res.Error != nil {
			log.Printf("EnsureProductBasePrices: slug %q: %v", slug, res.Error)
			return
		}
		updated += res.RowsAffected
	})
	if updated > 0 {
		log.Printf("Ensured base_price for %d product(s)", updated)
	}
	return nil
}
