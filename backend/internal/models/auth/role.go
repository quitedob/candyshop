package auth

import (
	common "candypro/api/internal/models/common"
	"time"
)

// Role name constants used across the platform.
const (
	User       = "customer"
	Admin      = "admin"
	SuperAdmin = "superadmin"
	Supplier   = "supplier"
)

func UserPortal() []string {
	return []string{User}
}

func AdminPortal() []string {
	return []string{Admin, SuperAdmin}
}

func SupplierPortal() []string {
	return []string{Supplier}
}

// Role represents a user role and its permissions
type Role struct {
	ID          string             `json:"id" gorm:"primaryKey"`
	Name        string             `json:"name" gorm:"uniqueIndex;not null"`
	Description string             `json:"description"`
	Permissions common.StringArray `json:"permissions" gorm:"type:jsonb"`
	IsSystem    bool               `json:"isSystem" gorm:"default:false"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
}
