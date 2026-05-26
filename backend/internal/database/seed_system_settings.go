package database

import (
	modelsCommon "candypro/api/internal/models/common"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type defaultSystemSetting struct {
	Key      string
	Value    string
	Category string
}

var defaultSystemSettings = []defaultSystemSetting{
	{Key: "site_name", Value: "CandyPro", Category: "general"},
	{Key: "default_currency", Value: "USD", Category: "general"},
	{Key: "support_email", Value: "info@candypro.com", Category: "general"},
	{Key: "smtp_host", Value: "", Category: "email"},
	{Key: "smtp_port", Value: "587", Category: "email"},
	{Key: "smtp_from", Value: "noreply@candypro.com", Category: "email"},
	{Key: "default_incoterms", Value: "FOB", Category: "trade"},
	{Key: "default_payment_terms", Value: "T/T 30% deposit, 70% before shipment", Category: "trade"},
	{Key: "notify_new_inquiry", Value: "true", Category: "notifications"},
	{Key: "notify_new_order", Value: "true", Category: "notifications"},
}

// EnsureDefaultSystemSettings upserts baseline admin settings when missing.
func EnsureDefaultSystemSettings(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	now := time.Now()
	for _, item := range defaultSystemSettings {
		row := modelsCommon.SystemSetting{
			Key:       item.Key,
			Value:     item.Value,
			Category:  item.Category,
			UpdatedAt: now,
		}
		if err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}},
			DoNothing: true,
		}).Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}
