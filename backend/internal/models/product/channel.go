package product

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Channel represents a sales channel (webstore, marketplace, etc.).
type Channel struct {
	ID                 uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Code               string         `json:"code" gorm:"type:varchar(40);uniqueIndex;not null"`
	Name               string         `json:"name" gorm:"type:varchar(120);not null"`
	Type               string         `json:"type" gorm:"type:varchar(40);not null;default:'marketplace'"` // webstore, marketplace, b2b, retail
	Status             string         `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	Currency           string         `json:"currency" gorm:"type:varchar(10);default:'USD'"`
	TaxConfig          ChannelTaxJSON `json:"taxConfig" gorm:"type:jsonb"`
	FulfillmentMode    string         `json:"fulfillmentMode" gorm:"type:varchar(20);default:'self'"` // self, 3pl, dropship
	PriceMultiplier    float64        `json:"priceMultiplier" gorm:"default:1.0"`
	DefaultWarehouse   *string        `json:"defaultWarehouse,omitempty" gorm:"type:varchar(100)"`
	AutoConfirm        bool           `json:"autoConfirm" gorm:"default:false"`       // auto-confirm orders
	AllowUnpaid        bool           `json:"allowUnpaid" gorm:"default:false"`       // allow unpaid order creation
	DraftExpireMinutes int            `json:"draftExpireMinutes" gorm:"default:60"`   // channel-level draft expiry
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

// ChannelTaxJSON is stored as JSONB.
type ChannelTaxJSON struct {
	TaxIncluded bool    `json:"taxIncluded"`
	TaxRate     float64 `json:"taxRate"`
	TaxLabel    string  `json:"taxLabel"`
}

func (t ChannelTaxJSON) Value() (driver.Value, error) {
	return json.Marshal(t)
}

func (t *ChannelTaxJSON) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, t)
}
