package common

import "time"

type Translation struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Key       string    `json:"key" gorm:"not null;uniqueIndex:idx_trans_key_locale"`
	Locale    string    `json:"locale" gorm:"not null;size:10;uniqueIndex:idx_trans_key_locale"`
	Value     string    `json:"value" gorm:"type:text"`
	Group     string    `json:"group" gorm:"size:50;index"`
	IsActive  bool      `json:"isActive" gorm:"default:true"`
	UpdatedBy *string   `json:"updatedBy"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
