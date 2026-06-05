package product

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsUser "candypro/api/internal/models/user"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// OEM Project status constants
const (
	OEMStatusInquiry     = "inquiry"
	OEMStatusSampling    = "sampling"
	OEMStatusFormulation = "formulation"
	OEMStatusQuotation   = "quotation"
	OEMStatusContract    = "contract"
	OEMStatusProduction  = "production"
	OEMStatusDelivery    = "delivery"
	OEMStatusCompleted   = "completed"
	OEMStatusCancelled   = "cancelled"
)

// ValidOEMStatusTransitions defines allowed OEM project status transitions.
var ValidOEMStatusTransitions = map[string]map[string]bool{
	OEMStatusInquiry:     {OEMStatusSampling: true, OEMStatusCancelled: true},
	OEMStatusSampling:    {OEMStatusFormulation: true, OEMStatusCancelled: true},
	OEMStatusFormulation: {OEMStatusQuotation: true, OEMStatusCancelled: true},
	OEMStatusQuotation:   {OEMStatusContract: true, OEMStatusCancelled: true},
	OEMStatusContract:    {OEMStatusProduction: true, OEMStatusCancelled: true},
	OEMStatusProduction:  {OEMStatusDelivery: true, OEMStatusCancelled: true},
	OEMStatusDelivery:    {OEMStatusCompleted: true},
	OEMStatusCompleted:   {},
	OEMStatusCancelled:   {},
}

// ValidateOEMStatusTransition checks if a status transition is allowed.
func ValidateOEMStatusTransition(from, to string) error {
	allowed, known := ValidOEMStatusTransitions[from]
	if !known {
		return nil // unknown current status — allow for legacy data
	}
	if allowed[to] {
		return nil
	}
	return fmt.Errorf("cannot transition OEM project from '%s' to '%s'", from, to)
}

type OEMProject struct {
	ID        string           `json:"id" gorm:"primaryKey"`
	UserID    string           `json:"userId" gorm:"index;not null"`
	User      *modelsUser.User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	InquiryID *string          `json:"inquiryId" gorm:"index"`
	// OrderID links the OEM project to the draft order created when the project
	// is converted (P0.2 / G-OEM-2). Nil until conversion; used for idempotency
	// so a project cannot be converted into multiple orders.
	OrderID     *string `json:"orderId" gorm:"index"`
	ProductName string  `json:"productName"`
	// QuotedUnitPrice / QuotedQuantity capture the agreed commercial terms an
	// admin sets at the quotation stage. They seed the order created at
	// conversion when no explicit override is supplied.
	QuotedUnitPrice float64                  `json:"quotedUnitPrice" gorm:"default:0"`
	QuotedQuantity  int                      `json:"quotedQuantity" gorm:"default:0"`
	ProductID       *string                  `json:"productId" gorm:"index"`          // optional catalog product the OEM maps to
	Status          string                   `json:"status" gorm:"default:'inquiry'"` // inquiry, sampling, formulation, quotation, contract, production, delivery, completed
	CurrentStep     int                      `json:"currentStep" gorm:"default:0"`
	Requirements    OEMRequirements          `json:"requirements" gorm:"type:jsonb"`
	Samples         OEMSampleArray           `json:"samples" gorm:"type:jsonb"`
	Attachments     modelsCommon.StringArray `json:"attachments" gorm:"type:jsonb"`
	AssignedTo      *string                  `json:"assignedTo" gorm:"index"`
	Notes           string                   `json:"notes" gorm:"type:text"`      // customer requirement notes (set at creation)
	AdminNotes      string                   `json:"adminNotes" gorm:"type:text"` // internal admin/team remarks
	Version         uint                     `json:"version" gorm:"not null;default:1"`
	CreatedAt       time.Time                `json:"createdAt"`
	UpdatedAt       time.Time                `json:"updatedAt"`
}

type OEMRequirements struct {
	Flavor         string   `json:"flavor"`
	Shape          string   `json:"shape"`
	Packaging      string   `json:"packaging"`
	TargetMarket   string   `json:"targetMarket"`
	Certifications []string `json:"certifications"`
	MOQ            int      `json:"moq"`
}

func (r OEMRequirements) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *OEMRequirements) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan OEMRequirements")
	}
	return json.Unmarshal(bytes, r)
}

type OEMSample struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Status     string     `json:"status"` // requested, shipped, received, approved, rejected
	SentAt     *time.Time `json:"sentAt"`
	ReceivedAt *time.Time `json:"receivedAt"`
	Feedback   string     `json:"feedback"`
}

type OEMSampleArray []OEMSample

// OEM sample status constants.
const (
	OEMSampleStatusRequested = "requested"
	OEMSampleStatusShipped   = "shipped"
	OEMSampleStatusReceived  = "received"
	OEMSampleStatusApproved  = "approved"
	OEMSampleStatusRejected  = "rejected"
)

// ValidOEMSampleStatuses lists the recognised sample lifecycle states.
var ValidOEMSampleStatuses = map[string]bool{
	OEMSampleStatusRequested: true,
	OEMSampleStatusShipped:   true,
	OEMSampleStatusReceived:  true,
	OEMSampleStatusApproved:  true,
	OEMSampleStatusRejected:  true,
}

// IsValidOEMSampleStatus reports whether s is a recognised sample status.
func IsValidOEMSampleStatus(s string) bool {
	return ValidOEMSampleStatuses[s]
}

func (s OEMSampleArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

func (s *OEMSampleArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan OEMSampleArray")
	}
	return json.Unmarshal(bytes, s)
}
