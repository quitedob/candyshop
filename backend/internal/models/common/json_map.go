package common

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap stores locale-keyed translations as JSONB.
// Structure: {"zh": {"name": "中文"}, "en": {"name": "English"}}
type JSONMap map[string]map[string]string

func (m JSONMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan JSONMap: expected []byte")
	}
	return json.Unmarshal(bytes, m)
}
