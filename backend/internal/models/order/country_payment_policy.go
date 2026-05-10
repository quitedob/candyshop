package order

// CountryPaymentPolicy defines payment requirements for a specific destination country.
// Replaces hardcoded India/Pakistan logic with a data-driven approach.
type CountryPaymentPolicy struct {
	ID                     uint   `gorm:"primaryKey" json:"id"`
	Country                string `gorm:"type:varchar(100);uniqueIndex;not null" json:"country"` // canonical name e.g. "india"
	RequiresFullPrepayment bool   `gorm:"default:false" json:"requiresFullPrepayment"`
	AllowedTerms           string `gorm:"type:text" json:"allowedTerms,omitempty"` // comma-separated e.g. "100% T/T before production"
	Note                   string `gorm:"type:text" json:"note,omitempty"`
}
