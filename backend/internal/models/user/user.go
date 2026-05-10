package user

import (
	"time"

	common "candypro/api/internal/models/common"

	"gorm.io/gorm"
)

// RoleSnapshot keeps user-role association without importing auth model package.
type RoleSnapshot struct {
	ID          string             `json:"id" gorm:"primaryKey"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Permissions common.StringArray `json:"permissions" gorm:"type:jsonb"`
	IsSystem    bool               `json:"isSystem"`
}

// TableName binds the snapshot to roles table.
func (RoleSnapshot) TableName() string {
	return "roles"
}

// User represents a system user
type User struct {
	ID                  string         `json:"id" gorm:"primaryKey"`
	Email               string         `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash        string         `json:"-" gorm:"not null"`
	FirstName           string         `json:"firstName"`
	LastName            string         `json:"lastName"`
	Phone               string         `json:"phone"`
	Company             string         `json:"company" gorm:"column:company_name"`
	CompanyID           *string        `json:"companyId" gorm:"index"`
	CompanyRef          *Company       `json:"companyRef,omitempty" gorm:"foreignKey:CompanyID"`
	Status              string         `json:"status" gorm:"default:'pending'"` // pending, active, suspended, deleted
	EmailVerified       bool           `json:"emailVerified" gorm:"default:false"`
	EmailVerifiedAt     *time.Time     `json:"emailVerifiedAt"`
	RoleID              string         `json:"roleId" gorm:"index"`
	Role                *RoleSnapshot  `json:"role,omitempty" gorm:"foreignKey:RoleID;references:ID"`
	LastLoginAt         *time.Time     `json:"lastLoginAt"`
	PasswordChangedAt   *time.Time     `json:"passwordChangedAt"`
	ResetToken          *string        `json:"-" gorm:"index"`
	ResetTokenExpiresAt *time.Time     `json:"-"`
	CreatedAt           time.Time      `json:"createdAt"`
	UpdatedAt           time.Time      `json:"updatedAt"`
	DeletedAt           gorm.DeletedAt `json:"-" gorm:"index"`
}
