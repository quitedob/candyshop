package product

import (
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
	ID           string          `json:"id" gorm:"primaryKey"`
	UserID       string          `json:"userId" gorm:"index;not null"`
	InquiryID    *string         `json:"inquiryId" gorm:"index"`
	ProductName  string          `json:"productName"`
	Status       string          `json:"status" gorm:"default:'inquiry'"` // inquiry, sampling, formulation, quotation, contract, production, delivery, completed
	CurrentStep  int             `json:"currentStep" gorm:"default:0"`
	Requirements OEMRequirements `json:"requirements" gorm:"type:jsonb"`
	Samples      OEMSampleArray  `json:"samples" gorm:"type:jsonb"`
	AssignedTo   *string         `json:"assignedTo" gorm:"index"`
	Notes        string          `json:"notes" gorm:"type:text"`
	CreatedAt    time.Time       `json:"createdAt"`
	UpdatedAt    time.Time       `json:"updatedAt"`
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
