package database

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"log"
	"time"

	"gorm.io/gorm"
)

// EnsureDefaultTaxShippingRates 启动时确保默认税费/运费费率存在（兼容已有数据库）
func EnsureDefaultTaxShippingRates(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	now := time.Now()

	taxSeeds := []modelsOrder.TaxRate{
		{ID: crypto.GenerateID(), Country: "US", Region: "", Rate: 0, Name: "US Export (B2B)", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: crypto.GenerateID(), Country: "US", Region: "NY", Rate: 0.08875, Name: "NY Sales Tax", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: crypto.GenerateID(), Country: "CN", Region: "", Rate: 0.13, Name: "VAT", IsActive: true, CreatedAt: now, UpdatedAt: now},
	}
	for _, tr := range taxSeeds {
		var count int64
		q := db.Model(&modelsOrder.TaxRate{}).Where("country = ? AND region = ?", tr.Country, tr.Region)
		if err := q.Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&tr).Error; err != nil {
				return err
			}
		}
	}

	shipSeeds := []modelsOrder.ShippingRate{
		{
			ID: crypto.GenerateID(), Destination: "US", MinWeightKg: 0, MaxWeightKg: 99999,
			BaseCost: 500, CostPerKg: 2.5, Currency: "USD", Carrier: "sea",
			EstimatedDays: 25, IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
		{
			ID: crypto.GenerateID(), Destination: "CN", MinWeightKg: 0, MaxWeightKg: 99999,
			BaseCost: 200, CostPerKg: 1.2, Currency: "USD", Carrier: "express",
			EstimatedDays: 7, IsActive: true, CreatedAt: now, UpdatedAt: now,
		},
	}
	for _, sr := range shipSeeds {
		var count int64
		if err := db.Model(&modelsOrder.ShippingRate{}).Where("destination = ? AND carrier = ?", sr.Destination, sr.Carrier).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			if err := db.Create(&sr).Error; err != nil {
				return err
			}
		}
	}

	log.Println("Default tax/shipping rates ensured")
	return nil
}

// seedTaxShippingRates 全量 seed 时写入税费/运费（与 Ensure 逻辑一致）
func seedTaxShippingRates(db *gorm.DB) error {
	return EnsureDefaultTaxShippingRates(db)
}
