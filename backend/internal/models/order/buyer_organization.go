package order

import "time"

// BuyerOrganization represents a customer's company for B2B procurement.
type BuyerOrganization struct {
	ID                uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name              string    `json:"name" gorm:"not null"`
	CreditLimit       float64   `json:"creditLimit" gorm:"default:0"`
	PaymentTerms      string    `json:"paymentTerms" gorm:"default:'net_30'"`
	ApprovalThreshold float64   `json:"approvalThreshold" gorm:"default:0"`
	IsActive          bool      `json:"isActive" gorm:"default:true"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (BuyerOrganization) TableName() string { return "buyer_organizations" }

// OrgMemberRole defines roles within a buyer organization.
type OrgMemberRole string

const (
	OrgRolePurchaser OrgMemberRole = "purchaser"
	OrgRoleApprover  OrgMemberRole = "approver"
	OrgRoleOrgAdmin  OrgMemberRole = "org_admin"
)

// OrgMember links a user to an organization with a role.
type OrgMember struct {
	ID             uint          `json:"id" gorm:"primaryKey;autoIncrement"`
	OrganizationID uint          `json:"organizationId" gorm:"not null;index"`
	UserID         string        `json:"userId" gorm:"not null;index"`
	Role           OrgMemberRole `json:"role" gorm:"not null;default:'purchaser'"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
}

func (OrgMember) TableName() string { return "org_members" }

// ApprovalAction records an approval decision on a pending-approval order.
type ApprovalAction struct {
	ID      uint   `json:"id" gorm:"primaryKey;autoIncrement"`
	OrderID string `json:"orderId" gorm:"not null;index"`
	// R2 E-12: indexed so audit views and approver-history lookups don't
	// trigger full-table scans of approval_actions.
	UserID    string    `json:"userId" gorm:"not null;index"`
	Action    string    `json:"action" gorm:"not null"` // approved, rejected, modified
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

func (ApprovalAction) TableName() string { return "approval_actions" }
