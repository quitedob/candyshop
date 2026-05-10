package trade

import (
	"time"

	"gorm.io/datatypes"
)

// TradeProcess Statuses
const (
	TradeStatusDraft     = "DRAFT"
	TradeStatusPending   = "PENDING"
	TradeStatusConfirmed = "CONFIRMED"
	TradeStatusCancelled = "CANCELLED"
	TradeStatusPaid      = "PAID"
	TradeStatusShipped   = "SHIPPED"
	TradeStatusCompleted = "COMPLETED"
)

// validTradeStatusTransitions defines allowed trade status transitions.
var ValidTradeStatusTransitions = map[string]map[string]bool{
	TradeStatusDraft:     {TradeStatusPending: true, TradeStatusCancelled: true},
	TradeStatusPending:   {TradeStatusConfirmed: true, TradeStatusCancelled: true},
	TradeStatusConfirmed: {TradeStatusPaid: true, TradeStatusCancelled: true},
	TradeStatusPaid:      {TradeStatusShipped: true},
	TradeStatusShipped:   {TradeStatusCompleted: true},
	TradeStatusCompleted: {},
	TradeStatusCancelled: {},
}

// Document Types
const (
	DocTypeQuotation         = "QUOTATION"
	DocTypeProformaInvoice   = "PROFORMA_INVOICE"
	DocTypeSalesContract     = "SALES_CONTRACT"
	DocTypeCommercialInvoice = "COMMERCIAL_INVOICE"
	DocTypePackingList       = "PACKING_LIST"
	DocTypeBillOfLading      = "BILL_OF_LADING"
	DocTypeHealthCertificate = "HEALTH_CERTIFICATE"
	DocTypeOriginCertificate = "ORIGIN_CERTIFICATE"
)

// TradeTransaction represents an entire end-to-end B2B trade transaction flow
type TradeTransaction struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	OrderID     *string `gorm:"type:varchar(100);uniqueIndex" json:"orderId,omitempty"` // Link to confirmed order (one trade per order)
	InquiryID   *string `json:"inquiryId"`                                              // Link to early-stage inquiry
	UserID      string  `gorm:"not null;index" json:"userId"`
	Reference   string  `gorm:"type:varchar(100);uniqueIndex" json:"reference"` // E.g., TRD-2026-001
	Status      string  `gorm:"type:varchar(50);default:'DRAFT'" json:"status"`
	Currency    string  `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	TotalAmount float64 `json:"totalAmount"`
	Terms       string  `gorm:"type:varchar(50)" json:"terms"` // Incoterms e.g. FOB, CIF, EXW
	// CommercialNotes 询盘映射：包装要求、议定付款方式等
	CommercialNotes string `gorm:"type:text" json:"commercialNotes,omitempty"`

	// Relationships to documents
	Documents []TradeDocument `gorm:"foreignKey:TransactionID" json:"documents,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// TradeDocument manages the individual forms: PI, CI, PL, Sales Contract
type TradeDocument struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	TransactionID uint   `gorm:"not null" json:"transactionId"`
	Type          string `gorm:"type:varchar(50);not null" json:"type"`                   // e.g. DOC_TYPE_PROFORMA_INVOICE
	DocNumber     string `gorm:"type:varchar(100);not null;uniqueIndex" json:"docNumber"` // E.g., PI-2026-001
	Status        string `gorm:"type:varchar(50);default:'DRAFT'" json:"status"`

	// Dynamic content based on the document type (JSON helps flexibility)
	// For PI/CI: Contains items, quantities, prices, incoterms, bank details
	// For PL: Contains items, gross/net weight, volume/CBM, carton counts
	// For Contracts: Contains specific clauses, quality standards, etc.
	Content datatypes.JSON `json:"content"`

	// AI Generation metadata
	IsAIGenerated bool `gorm:"default:false" json:"isAiGenerated"`
	AIVersion     int  `gorm:"default:0" json:"aiVersion"`

	// SourceOrderID 派生单证时关联的订单 ID（审计/对账）
	SourceOrderID *string `gorm:"type:varchar(100);index" json:"sourceOrderId,omitempty"`
	// LineageSource 派生来源：manual | order_derived | ai_draft
	LineageSource string `gorm:"type:varchar(40);default:'manual'" json:"lineageSource"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ComplianceRequirement holds the destination country specific requirements
type ComplianceRequirement struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	TransactionID uint   `gorm:"not null" json:"transactionId"`
	CountryCode   string `gorm:"type:varchar(5);not null" json:"countryCode"`
	Language      string `gorm:"type:varchar(50)" json:"language"` // Label language requirement

	// JSON block for compliance checklist
	// E.g., Additives allowed/banned, nutritional formatting, allergen warnings
	Rules  datatypes.JSON `json:"rules"`
	Status string         `gorm:"type:varchar(50);default:'PENDING'" json:"status"` // PENDING, PASSED, FAILED

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
