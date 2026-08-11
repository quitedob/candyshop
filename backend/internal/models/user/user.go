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
	ID string `json:"id" gorm:"primaryKey"`
	// Email is unique only among non-deleted rows (partial unique index on
	// deleted_at IS NULL) so a soft-deleted account's address can be re-registered
	// instead of 5xx-ing on a full-table unique constraint (G24a).
	Email               string         `json:"email" gorm:"uniqueIndex:idx_users_email_active,where:deleted_at IS NULL;not null"`
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

// LegacyUniqueIndexEmail is the pre-G24a full-table unique index name that GORM
// created from the former unnamed `uniqueIndex` tag on User.Email. GORM
// AutoMigrate never drops indexes that were removed from the model, so on any
// already-migrated database this index still spans soft-deleted rows and blocks
// re-registering a deleted email with a unique-constraint 5xx. The replacement
// partial index is idx_users_email_active (WHERE deleted_at IS NULL).
const LegacyUniqueIndexEmail = "idx_users_email"

// DropLegacyEmailUniqueIndex removes the pre-G24a full-table unique index on
// users.email if it is still present. It is a no-op when the index does not
// exist, so it is safe to call on every startup and on fresh databases. Call it
// once after AutoMigrate on already-migrated deployments so a soft-deleted
// email can be re-registered instead of raising the G24a 5xx.
func DropLegacyEmailUniqueIndex(db *gorm.DB) error {
	if db.Migrator().HasIndex(&User{}, LegacyUniqueIndexEmail) {
		return db.Migrator().DropIndex(&User{}, LegacyUniqueIndexEmail)
	}
	return nil
}
