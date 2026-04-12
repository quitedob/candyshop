package common

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// OEMStep represents one step in an OEM service flow.
type OEMStep struct {
	Order        int         `json:"order"`
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	Duration     string      `json:"duration"`
	Deliverables StringArray `json:"deliverables"`
}

// OEMStepArray stores OEM steps in JSONB.
type OEMStepArray []OEMStep

// Value implements driver.Valuer.
func (s OEMStepArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner.
func (s *OEMStepArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan OEMStepArray: expected []byte")
	}
	return json.Unmarshal(bytes, s)
}

// Milestone represents one factory/history milestone.
type Milestone struct {
	Year        int    `json:"year"`
	Milestone   string `json:"milestone"`
	Description string `json:"description"`
}

// MilestoneArray stores milestones in JSONB.
type MilestoneArray []Milestone

// Value implements driver.Valuer.
func (s MilestoneArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner.
func (s *MilestoneArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan MilestoneArray: expected []byte")
	}
	return json.Unmarshal(bytes, s)
}
