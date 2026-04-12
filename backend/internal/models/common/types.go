package common

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringArray is a custom type for storing string arrays in PostgreSQL as JSON
type StringArray []string

// Value implements driver.Valuer interface
func (s StringArray) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan StringArray: expected []byte")
	}
	return json.Unmarshal(bytes, s)
}
