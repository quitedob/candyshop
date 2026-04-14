package common

import "time"

// SystemSetting represents a key-value configuration entry.
type SystemSetting struct {
	Key       string    `json:"key" gorm:"primaryKey"`
	Value     string    `json:"value" gorm:"type:text"`
	Category  string    `json:"category" gorm:"index"`
	UpdatedBy *string   `json:"updatedBy"`
	UpdatedAt time.Time `json:"updatedAt"`
}
