package order

import (
	"context"
	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

// BuyerOrgRepository handles buyer organization data access.
type BuyerOrgRepository struct {
	db *gorm.DB
}

func NewBuyerOrgRepository(db *gorm.DB) *BuyerOrgRepository {
	return &BuyerOrgRepository{db: db}
}

func (r *BuyerOrgRepository) Create(ctx context.Context, org *modelsOrder.BuyerOrganization) error {
	return r.db.WithContext(ctx).Create(org).Error
}

func (r *BuyerOrgRepository) FindByID(ctx context.Context, id uint) (*modelsOrder.BuyerOrganization, error) {
	var org modelsOrder.BuyerOrganization
	err := r.db.WithContext(ctx).First(&org, id).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

// FindByIDs returns buyer organisations matching the given IDs in a single query.
// Used by ApprovalService to avoid an N+1 lookup when checking approval thresholds
// for users who belong to multiple orgs (M-2).
func (r *BuyerOrgRepository) FindByIDs(ctx context.Context, ids []uint) ([]modelsOrder.BuyerOrganization, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var orgs []modelsOrder.BuyerOrganization
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&orgs).Error; err != nil {
		return nil, err
	}
	return orgs, nil
}

func (r *BuyerOrgRepository) FindAll(ctx context.Context) ([]modelsOrder.BuyerOrganization, error) {
	var orgs []modelsOrder.BuyerOrganization
	err := r.db.WithContext(ctx).Order("name").Find(&orgs).Error
	return orgs, err
}

func (r *BuyerOrgRepository) Update(ctx context.Context, org *modelsOrder.BuyerOrganization) error {
	return r.db.WithContext(ctx).Save(org).Error
}

// OrgMemberRepository handles organization membership data access.
type OrgMemberRepository struct {
	db *gorm.DB
}

func NewOrgMemberRepository(db *gorm.DB) *OrgMemberRepository {
	return &OrgMemberRepository{db: db}
}

func (r *OrgMemberRepository) Create(ctx context.Context, m *modelsOrder.OrgMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *OrgMemberRepository) FindByUserID(ctx context.Context, userID string) ([]modelsOrder.OrgMember, error) {
	var members []modelsOrder.OrgMember
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&members).Error
	return members, err
}

func (r *OrgMemberRepository) FindByOrgID(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error) {
	var members []modelsOrder.OrgMember
	err := r.db.WithContext(ctx).Where("organization_id = ?", orgID).Find(&members).Error
	return members, err
}

func (r *OrgMemberRepository) FindApprovers(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error) {
	var members []modelsOrder.OrgMember
	err := r.db.WithContext(ctx).
		Where("organization_id = ? AND role IN ?", orgID, []string{"approver", "org_admin"}).
		Find(&members).Error
	return members, err
}

// FindApproversForOrgs returns approvers for many orgs in a single query (M-3).
func (r *OrgMemberRepository) FindApproversForOrgs(ctx context.Context, orgIDs []uint) (map[uint][]modelsOrder.OrgMember, error) {
	if len(orgIDs) == 0 {
		return nil, nil
	}
	var members []modelsOrder.OrgMember
	if err := r.db.WithContext(ctx).
		Where("organization_id IN ? AND role IN ?", orgIDs, []string{"approver", "org_admin"}).
		Find(&members).Error; err != nil {
		return nil, err
	}
	out := make(map[uint][]modelsOrder.OrgMember, len(orgIDs))
	for _, m := range members {
		out[m.OrganizationID] = append(out[m.OrganizationID], m)
	}
	return out, nil
}

func (r *OrgMemberRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsOrder.OrgMember{}, id).Error
}

// ApprovalActionRepository handles approval action records.
type ApprovalActionRepository struct {
	db *gorm.DB
}

func NewApprovalActionRepository(db *gorm.DB) *ApprovalActionRepository {
	return &ApprovalActionRepository{db: db}
}

func (r *ApprovalActionRepository) Create(ctx context.Context, a *modelsOrder.ApprovalAction) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *ApprovalActionRepository) FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.ApprovalAction, error) {
	var actions []modelsOrder.ApprovalAction
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at DESC").Find(&actions).Error
	return actions, err
}
