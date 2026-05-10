package user

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// CompanyAddress represents a company address stored as JSONB.
type CompanyAddress struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zipCode"`
	Country string `json:"country"`
}

// Value implements driver.Valuer interface.
func (a CompanyAddress) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements sql.Scanner interface.
func (a *CompanyAddress) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan CompanyAddress: expected []byte")
	}
	return json.Unmarshal(bytes, a)
}

// Company represents a business organization in the B2B platform.
type Company struct {
	ID              string         `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"not null"`
	TaxID           string         `json:"taxId"`
	RegistrationNo  string         `json:"registrationNo"`
	BusinessLicense string         `json:"businessLicense"`
	Address         CompanyAddress `json:"address" gorm:"type:jsonb"`
	Phone           string         `json:"phone"`
	Website         string         `json:"website"`
	Status          string         `json:"status" gorm:"default:'pending'"` // pending, verified, rejected
	PriceListID     *string        `json:"priceListId" gorm:"index"`
	CreditLimit     float64        `json:"creditLimit" gorm:"default:0"`
	PaymentTerms    string         `json:"paymentTerms" gorm:"default:'NET_30'"` // NET_15, NET_30, NET_60, COD, PREPAID
	VerifiedAt      *time.Time     `json:"verifiedAt"`
	CreatedAt       time.Time      `json:"createdAt"`
	UpdatedAt       time.Time      `json:"updatedAt"`
}
