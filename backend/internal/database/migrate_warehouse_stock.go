package database

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"encoding/json"
	"log"
	"time"

	"gorm.io/gorm"
)

const mainWarehouseCode = "MAIN"

// defaultProductWeightedAvgCosts 种子产品默认单位成本（约为标价的 40%）
var defaultProductWeightedAvgCosts = map[string]float64{
	"4d-fruit-gummy":             3.40,
	"crystal-hard-candy":         2.04,
	"rainbow-lollipop":           1.52,
	"sour-belt":                  1.68,
	"jelly-fruits":               2.40,
	"chewy-toffee":               3.00,
	"gummy-bear-classic":         1.98,
	"gummy-worms":                2.10,
	"fruit-slices":               1.80,
	"hard-candy-assorted-mix":    1.44,
	"hard-candy-fruit-bonbon":    1.92,
	"chocolate-premium-truffles": 4.80,
	"chocolate-variety-box":      6.00,
	"licorice-black-twists":      1.56,
	"licorice-fruit-ropes":       1.64,
	"sour-gummy-extreme":         2.20,
	"sour-belt-rainbow":          1.72,
	"sour-gummy-bears-zing":      2.00,
}

// EnsureDefaultWarehouseStock 启动时确保默认仓库及仓级库存行存在（兼容已有数据库）
func EnsureDefaultWarehouseStock(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	whID, err := ensureMainWarehouse(db)
	if err != nil {
		return err
	}
	var products []modelsProduct.Product
	if err := db.Select("id", "stock_quantity").Where("stock_quantity > 0").Find(&products).Error; err != nil {
		return err
	}
	var created int64
	for _, p := range products {
		var count int64
		if err := db.Model(&modelsProduct.WarehouseStock{}).
			Where("warehouse_id = ? AND product_id = ?", whID, p.ID).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		row := modelsProduct.WarehouseStock{
			WarehouseID: whID,
			ProductID:   p.ID,
			Quantity:    p.StockQuantity,
			Reserved:    0,
			UpdatedAt:   time.Now(),
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
		created++
	}
	if created > 0 {
		log.Printf("Ensured default warehouse stock rows (%d new)", created)
	}
	if err := BackfillProductWeightedAvgCost(db); err != nil {
		return err
	}
	if err := BackfillCrystalHardCandyHalal(db); err != nil {
		return err
	}
	return nil
}

// ensureMainWarehouse 返回默认 MAIN 仓库 ID，不存在则创建
func ensureMainWarehouse(db *gorm.DB) (string, error) {
	var w modelsProduct.Warehouse
	if err := db.Where("code = ? AND is_active = ?", mainWarehouseCode, true).First(&w).Error; err == nil {
		return w.ID, nil
	}
	now := time.Now()
	w = modelsProduct.Warehouse{
		ID:        crypto.GenerateID(),
		Name:      "Main Factory Warehouse",
		Code:      mainWarehouseCode,
		Type:      "factory",
		Country:   "CN",
		IsActive:  true,
		IsDefault: true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(&w).Error; err != nil {
		return "", err
	}
	log.Printf("Created default warehouse %s (%s)", w.Code, w.ID)
	return w.ID, nil
}

// seedWarehouses 为新 seed 数据库创建 MAIN 仓及仓级库存
func seedWarehouses(db *gorm.DB) error {
	whID, err := ensureMainWarehouse(db)
	if err != nil {
		return err
	}
	var products []modelsProduct.Product
	if err := db.Select("id", "stock_quantity").Find(&products).Error; err != nil {
		return err
	}
	for _, p := range products {
		if p.StockQuantity <= 0 {
			continue
		}
		row := modelsProduct.WarehouseStock{
			WarehouseID: whID,
			ProductID:   p.ID,
			Quantity:    p.StockQuantity,
			Reserved:    0,
			UpdatedAt:   time.Now(),
		}
		if err := db.Where(modelsProduct.WarehouseStock{WarehouseID: whID, ProductID: p.ID}).
			Assign(row).
			FirstOrCreate(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

// BackfillProductWeightedAvgCost 为 weighted_avg_cost=0 的产品补默认成本
func BackfillProductWeightedAvgCost(db *gorm.DB) error {
	var updated int64
	for slug, cost := range defaultProductWeightedAvgCosts {
		res := db.Model(&modelsProduct.Product{}).
			Where("slug = ? AND (weighted_avg_cost = 0 OR weighted_avg_cost IS NULL)", slug).
			Update("weighted_avg_cost", cost)
		if res.Error != nil {
			return res.Error
		}
		updated += res.RowsAffected
	}
	if updated > 0 {
		log.Printf("Backfilled product weighted avg cost (%d rows)", updated)
	}
	return nil
}

// BackfillCrystalHardCandyHalal 补全 Crystal Hard Candy 的 Halal 认证字段
func BackfillCrystalHardCandyHalal(db *gorm.DB) error {
	certs := modelsCommon.StringArray{"HACCP", "ISO 22000", "Halal"}
	raw, err := json.Marshal(certs)
	if err != nil {
		return err
	}
	res := db.Exec(`
		UPDATE products
		SET certifications = ?::jsonb,
		    halal_certified = true,
		    updated_at = NOW()
		WHERE slug = 'crystal-hard-candy'
		  AND NOT (certifications @> '["Halal"]'::jsonb)
	`, string(raw))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		log.Printf("Backfilled Halal certification for crystal-hard-candy")
	}
	return nil
}
