package trade

import (
	"time"

	"gorm.io/datatypes"
)

// Quotation review lifecycle statuses.
const (
	QuotationReviewStatusPending  = "pending"
	QuotationReviewStatusApproved = "approved"
	QuotationReviewStatusRejected = "rejected"
)

// QuotationReview persists a quotation queued by the AI for sales-manager
// approval — the persistence layer for the submit_quotation_for_human_review
// tool. Previously the tool marshalled the payload to JSON and discarded it.
type QuotationReview struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	CustomerRef          string         `json:"customerRef" gorm:"index"`
	Currency             string         `json:"currency"`
	TotalAmount          float64        `json:"totalAmount"`
	LineItems            datatypes.JSON `json:"lineItems" gorm:"type:jsonb"`
	StrategyNotes        string         `json:"strategyNotes" gorm:"type:text"`
	SuggestedDiscountPct float64        `json:"suggestedDiscountPct"`
	Status               string         `json:"status" gorm:"index"`
	Notes                string         `json:"notes" gorm:"type:text"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
}

// TableName returns the plural table name for GORM.
func (QuotationReview) TableName() string { return "quotation_reviews" }
