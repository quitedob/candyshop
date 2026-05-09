package trade

import (
	"time"
)

// SalesContract represents a formalized B2B product sales agreement
type SalesContract struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TransactionID   uint       `gorm:"not null;uniqueIndex" json:"transactionId"`
	ContractNo      string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"contractNo"`
	BuyerName       string     `gorm:"type:varchar(255)" json:"buyerName"`
	SellerName      string     `gorm:"type:varchar(255)" json:"sellerName"`
	Incoterms       string     `gorm:"type:varchar(50)" json:"incoterms"` // e.g., FOB, CIF, EXW
	TermsOfPayment  string     `gorm:"type:text" json:"termsOfPayment"` // e.g., 30% T/T Advance, 70% L/C
	QualityStandard string     `gorm:"type:text" json:"qualityStandard"` // Detailed candy/food quality specs
	TotalAmount     float64    `json:"totalAmount"`
	Currency        string     `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	SignDate        *time.Time `json:"signDate"`
	ValidUntil      *time.Time `json:"validUntil"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

// PackingList represents the physical packing configuration of the shipment
type PackingList struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	TransactionID    uint      `gorm:"not null;uniqueIndex" json:"transactionId"`
	PLNumber         string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"plNumber"`
	TotalCartons     int       `json:"totalCartons"`
	TotalPallets     int       `json:"totalPallets"`
	TotalGrossWeight float64   `json:"totalGrossWeight"`                 // in kg
	TotalNetWeight   float64   `json:"totalNetWeight"`                   // in kg
	TotalVolume      float64   `json:"totalVolume"`                       // in CBM
	MarksAndNumbers  string    `gorm:"type:text" json:"marksAndNumbers"` // Shipping marks
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// CertificateOfOrigin represents a COO document
type CertificateOfOrigin struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	TransactionID      uint       `gorm:"not null;uniqueIndex" json:"transactionId"`
	CertificateNo      string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"certificateNo"`
	CountryOfOrigin    string     `gorm:"type:varchar(100);not null" json:"countryOfOrigin"`
	DestinationCountry string     `gorm:"type:varchar(100);not null" json:"destinationCountry"`
	FTAType            string     `gorm:"type:varchar(100)" json:"ftaType"` // e.g., FORM E, FORM A, RCEP
	IssueDate          *time.Time `json:"issueDate"`
	HSCode             string     `gorm:"type:varchar(50)" json:"hsCode"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// HealthCertificate represents food/sanitary/health certifications
type HealthCertificate struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	TransactionID      uint       `gorm:"not null;uniqueIndex" json:"transactionId"`
	CertificateNo      string     `gorm:"type:varchar(100);not null;uniqueIndex" json:"certificateNo"`
	IssuingAuthority   string     `gorm:"type:varchar(255)" json:"issuingAuthority"` // E.g., Customs/FDA
	Consignor          string     `gorm:"type:varchar(255)" json:"consignor"`
	Consignee          string     `gorm:"type:varchar(255)" json:"consignee"`
	DestinationCountry string     `gorm:"type:varchar(100);not null" json:"destinationCountry"`
	IssueDate          *time.Time `json:"issueDate"`
	ValidUntil         *time.Time `json:"validUntil"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
}

// ShipmentTracking represents logistics and transport events
type ShipmentTracking struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TransactionID   uint       `gorm:"not null" json:"transactionId"`
	BillOfLadingNo  string     `gorm:"type:varchar(100);index" json:"billOfLadingNo"`
	CarrierName     string     `gorm:"type:varchar(255)" json:"carrierName"`
	VesselFlight    string     `gorm:"type:varchar(100)" json:"vesselFlight"`
	PortOfLoading   string     `gorm:"type:varchar(255)" json:"portOfLoading"`
	PortOfDischarge string     `gorm:"type:varchar(255)" json:"portOfDischarge"`
	ETD             *time.Time `json:"etd"`                                              // Estimated Time of Departure
	ETA             *time.Time `json:"eta"`                                              // Estimated Time of Arrival
	Status           string     `gorm:"type:varchar(50);default:'PENDING'" json:"status"` // PENDING, DISPATCHED, IN_TRANSIT, DELIVERED
	DeliveredAt      *time.Time `json:"deliveredAt,omitempty"`
	DeliveryProofURL string     `gorm:"type:varchar(500)" json:"deliveryProofUrl,omitempty"`
	SignedBy         string     `gorm:"type:varchar(255)" json:"signedBy,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

// SettlementRecord represents payment milestones and statuses
type SettlementRecord struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	TransactionID uint       `gorm:"not null" json:"transactionId"`
	PaymentMethod string     `gorm:"type:varchar(50)" json:"paymentMethod"` // T/T, L/C, D/P
	AmountDue     float64    `json:"amountDue"`
	AmountPaid    float64    `json:"amountPaid"`
	Currency      string     `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	DueDate       *time.Time `json:"dueDate"`
	ExpectedDate  *time.Time `json:"expectedDate"`
	PaymentDate   *time.Time `json:"paymentDate"`
	Status        string     `gorm:"type:varchar(50);default:'UNPAID'" json:"status"` // UNPAID, PARTIAL, PAID
	LCReference   string     `gorm:"type:varchar(100)" json:"lcReference,omitempty"` // For L/C payments
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ProformaInvoice is the pre-shipment invoice used to confirm order terms before production.
type ProformaInvoice struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TransactionID   uint      `gorm:"not null;uniqueIndex" json:"transactionId"`
	PINumber        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"piNumber"`
	BuyerName       string    `gorm:"type:varchar(255)" json:"buyerName"`
	SellerName      string    `gorm:"type:varchar(255)" json:"sellerName"`
	Incoterms       string    `gorm:"type:varchar(50)" json:"incoterms"`
	TermsOfPayment  string    `gorm:"type:text" json:"termsOfPayment"`
	TotalAmount     float64   `json:"totalAmount"`
	Currency        string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	ValidUntil      *time.Time `json:"validUntil"`
	BankDetails     string    `gorm:"type:text" json:"bankDetails"`
	Notes           string    `gorm:"type:text" json:"notes"`
	Status          string    `gorm:"type:varchar(50);default:'DRAFT'" json:"status"` // DRAFT, SENT, CONFIRMED
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// CommercialInvoice is the final invoice issued after shipment for customs and payment.
type CommercialInvoice struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TransactionID   uint      `gorm:"not null;uniqueIndex" json:"transactionId"`
	CINumber        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"ciNumber"`
	PINumber        string    `gorm:"type:varchar(100)" json:"piNumber"` // Reference to originating PI
	BuyerName       string    `gorm:"type:varchar(255)" json:"buyerName"`
	SellerName      string    `gorm:"type:varchar(255)" json:"sellerName"`
	Incoterms       string    `gorm:"type:varchar(50)" json:"incoterms"`
	TermsOfPayment  string    `gorm:"type:text" json:"termsOfPayment"`
	TotalAmount     float64   `json:"totalAmount"`
	Currency        string    `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	ShipDate        *time.Time `json:"shipDate"`
	BankDetails     string    `gorm:"type:text" json:"bankDetails"`
	Notes           string    `gorm:"type:text" json:"notes"`
	Status          string    `gorm:"type:varchar(50);default:'DRAFT'" json:"status"` // DRAFT, ISSUED, PAID
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// BillOfLading is the transport document and title of goods issued by the carrier.
type BillOfLading struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TransactionID   uint      `gorm:"not null;uniqueIndex" json:"transactionId"`
	BLNumber        string    `gorm:"type:varchar(100);not null;uniqueIndex" json:"blNumber"`
	Shipper         string    `gorm:"type:varchar(255)" json:"shipper"`
	Consignee       string    `gorm:"type:varchar(255)" json:"consignee"`
	NotifyParty     string    `gorm:"type:varchar(255)" json:"notifyParty"`
	CarrierName     string    `gorm:"type:varchar(255)" json:"carrierName"`
	VesselVoyage    string    `gorm:"type:varchar(100)" json:"vesselVoyage"`
	PortOfLoading   string    `gorm:"type:varchar(255)" json:"portOfLoading"`
	PortOfDischarge string    `gorm:"type:varchar(255)" json:"portOfDischarge"`
	PlaceOfDelivery string    `gorm:"type:varchar(255)" json:"placeOfDelivery"`
	OnBoardDate     *time.Time `json:"onBoardDate"`
	FreightTerms    string    `gorm:"type:varchar(50)" json:"freightTerms"` // PREPAID, COLLECT
	GoodsDescription string   `gorm:"type:text" json:"goodsDescription"`
	GrossWeight     float64   `json:"grossWeight"` // kg
	Measurement     float64   `json:"measurement"` // CBM
	NumberOfPackages int      `json:"numberOfPackages"`
	Status          string    `gorm:"type:varchar(50);default:'DRAFT'" json:"status"` // DRAFT, ISSUED, SURRENDERED
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}
