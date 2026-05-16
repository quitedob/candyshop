package order

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// Event represents a lifecycle event in the system.
type Event struct {
	ID        uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string          `json:"name" gorm:"type:varchar(100);not null;index"`
	Payload   json.RawMessage `json:"payload" gorm:"type:jsonb"`
	CreatedAt time.Time       `json:"createdAt"`
}

func (Event) TableName() string { return "events" }

// HookConfig holds a registered hook (internal or plugin).
type HookConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name      string    `json:"name" gorm:"type:varchar(120);not null"`
	EventName string    `json:"eventName" gorm:"type:varchar(100);not null;index"`
	Type      string    `json:"type" gorm:"type:varchar(20);not null;default:'webhook'"` // webhook, plugin, internal
	Config    JSONMap   `json:"config" gorm:"type:jsonb"`
	Status    string    `json:"status" gorm:"type:varchar(20);not null;default:'active'"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (HookConfig) TableName() string { return "hook_configs" }

// JSONMap is a map[string]any stored as JSONB.
type JSONMap map[string]any

func (m JSONMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *JSONMap) Scan(value any) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, m)
}

// HookExecution logs each hook invocation.
type HookExecution struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	HookID       uint       `json:"hookId" gorm:"not null;index"`
	EventID      uint       `json:"eventId" gorm:"not null;index"`
	Status       string    `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	ResponseCode *int       `json:"responseCode,omitempty"`
	ResponseBody string    `json:"responseBody,omitempty" gorm:"type:text"`
	Error        string    `json:"error,omitempty" gorm:"type:text"`
	DurationMs   int       `json:"durationMs" gorm:"default:0"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (HookExecution) TableName() string { return "hook_executions" }
