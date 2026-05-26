package order

import (
	"context"
	"fmt"

	modelsOrder "candypro/api/internal/models/order"
)

type approvalOrgRepo interface {
	FindByID(ctx context.Context, id uint) (*modelsOrder.BuyerOrganization, error)
	FindByIDs(ctx context.Context, ids []uint) ([]modelsOrder.BuyerOrganization, error)
	FindAll(ctx context.Context) ([]modelsOrder.BuyerOrganization, error)
	Create(ctx context.Context, org *modelsOrder.BuyerOrganization) error
	Update(ctx context.Context, org *modelsOrder.BuyerOrganization) error
}

type approvalMemberRepo interface {
	FindByUserID(ctx context.Context, userID string) ([]modelsOrder.OrgMember, error)
	FindByOrgID(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error)
	FindApprovers(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error)
	FindApproversForOrgs(ctx context.Context, orgIDs []uint) (map[uint][]modelsOrder.OrgMember, error)
	Create(ctx context.Context, m *modelsOrder.OrgMember) error
	Delete(ctx context.Context, id uint) error
}

type approvalActionRepo interface {
	Create(ctx context.Context, a *modelsOrder.ApprovalAction) error
	FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.ApprovalAction, error)
}

// orderApprovalOps defines the subset of order repository methods needed for approval.
type orderApprovalOps interface {
	FindByID(ctx context.Context, id string) (*modelsOrder.Order, error)
	ApprovePendingOrder(ctx context.Context, id string) error
	ApprovePendingOrderWithAudit(ctx context.Context, id, userID, action, comment string) error
	CancelPendingApprovalOrder(ctx context.Context, orderID string) error
	CancelPendingApprovalOrderWithAudit(ctx context.Context, orderID, userID, action, comment string) error
	Update(ctx context.Context, order *modelsOrder.Order) error
}

// ApprovalService handles B2B purchase approval logic.
type ApprovalService struct {
	orgRepo    approvalOrgRepo
	memberRepo approvalMemberRepo
	actionRepo approvalActionRepo
	orderOps   orderApprovalOps
}

func NewApprovalService(orgRepo approvalOrgRepo, memberRepo approvalMemberRepo, actionRepo approvalActionRepo, orderOps orderApprovalOps) *ApprovalService {
	return &ApprovalService{orgRepo: orgRepo, memberRepo: memberRepo, actionRepo: actionRepo, orderOps: orderOps}
}

// ShouldRequireApproval checks if an order total exceeds the buyer org's approval threshold.
//
// M-2: organisations are loaded with a single FindByIDs query, regardless of how
// many memberships the purchaser has, instead of the previous per-member roundtrip.
func (s *ApprovalService) ShouldRequireApproval(ctx context.Context, userID string, orderTotal float64) (bool, *modelsOrder.BuyerOrganization, error) {
	members, err := s.memberRepo.FindByUserID(ctx, userID)
	if err != nil || len(members) == 0 {
		return false, nil, nil
	}
	purchaserOrgIDs := make([]uint, 0, len(members))
	for _, m := range members {
		if m.Role == modelsOrder.OrgRolePurchaser {
			purchaserOrgIDs = append(purchaserOrgIDs, m.OrganizationID)
		}
	}
	if len(purchaserOrgIDs) == 0 {
		return false, nil, nil
	}
	orgs, err := s.orgRepo.FindByIDs(ctx, purchaserOrgIDs)
	if err != nil {
		return false, nil, nil
	}
	for i := range orgs {
		org := orgs[i]
		if org.ApprovalThreshold > 0 && orderTotal > org.ApprovalThreshold {
			return true, &org, nil
		}
	}
	return false, nil, nil
}

// GetOrgApprovers returns users who can approve orders for the org.
func (s *ApprovalService) GetOrgApprovers(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error) {
	return s.memberRepo.FindApprovers(ctx, orgID)
}

// RecordApprovalAction records an approval decision.
func (s *ApprovalService) RecordApprovalAction(ctx context.Context, orderID, userID, action, comment string) error {
	a := &modelsOrder.ApprovalAction{
		OrderID: orderID,
		UserID:  userID,
		Action:  action,
		Comment: comment,
	}
	return s.actionRepo.Create(ctx, a)
}

// GetApprovalHistory returns all approval actions for an order.
func (s *ApprovalService) GetApprovalHistory(ctx context.Context, orderID string) ([]modelsOrder.ApprovalAction, error) {
	return s.actionRepo.FindByOrderID(ctx, orderID)
}

// CreateOrganization creates a new buyer organization.
func (s *ApprovalService) CreateOrganization(ctx context.Context, org *modelsOrder.BuyerOrganization) error {
	return s.orgRepo.Create(ctx, org)
}

// GetOrganizations returns all buyer organizations.
func (s *ApprovalService) GetOrganizations(ctx context.Context) ([]modelsOrder.BuyerOrganization, error) {
	return s.orgRepo.FindAll(ctx)
}

// GetOrganization returns a single buyer organization.
func (s *ApprovalService) GetOrganization(ctx context.Context, id uint) (*modelsOrder.BuyerOrganization, error) {
	return s.orgRepo.FindByID(ctx, id)
}

// UpdateOrganization updates a buyer organization.
func (s *ApprovalService) UpdateOrganization(ctx context.Context, org *modelsOrder.BuyerOrganization) error {
	return s.orgRepo.Update(ctx, org)
}

// AddMember adds a user to an organization.
func (s *ApprovalService) AddMember(ctx context.Context, m *modelsOrder.OrgMember) error {
	return s.memberRepo.Create(ctx, m)
}

// RemoveMember removes a user from an organization.
func (s *ApprovalService) RemoveMember(ctx context.Context, id uint) error {
	return s.memberRepo.Delete(ctx, id)
}

// GetUserOrganizations returns organizations a user belongs to.
func (s *ApprovalService) GetUserOrganizations(ctx context.Context, userID string) ([]modelsOrder.OrgMember, error) {
	return s.memberRepo.FindByUserID(ctx, userID)
}

// GetOrgMembers returns members of a single buyer organization.
// Backed by an indexed FindByOrgID query rather than a full table scan + Go-side filter
// (H-8). This replaces the previous misuse of FindByUserID("") that loaded every member.
func (s *ApprovalService) GetOrgMembers(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error) {
	return s.memberRepo.FindByOrgID(ctx, orgID)
}

// UserCanApproveOrder checks whether approverID may approve/reject order placed by a purchaser in the same org.
//
// M-3: approver lists for all the purchaser's orgs are loaded with a single
// FindApproversForOrgs query rather than one per org.
func (s *ApprovalService) UserCanApproveOrder(ctx context.Context, approverID string, order *modelsOrder.Order) (bool, error) {
	purchaserMembers, err := s.memberRepo.FindByUserID(ctx, order.UserID)
	if err != nil {
		return false, err
	}
	if len(purchaserMembers) == 0 {
		return false, nil
	}
	orgIDs := make([]uint, 0, len(purchaserMembers))
	for _, pm := range purchaserMembers {
		if pm.Role == modelsOrder.OrgRolePurchaser {
			orgIDs = append(orgIDs, pm.OrganizationID)
		}
	}
	if len(orgIDs) == 0 {
		return false, nil
	}
	approversByOrg, err := s.memberRepo.FindApproversForOrgs(ctx, orgIDs)
	if err != nil {
		return false, err
	}
	for _, list := range approversByOrg {
		for _, a := range list {
			if a.UserID == approverID {
				return true, nil
			}
		}
	}
	return false, nil
}

// ApproveOrderByAdmin approves without org membership check (admin portal).
func (s *ApprovalService) ApproveOrderByAdmin(ctx context.Context, orderID, userID, comment string) error {
	order, err := s.orderOps.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != modelsOrder.OrderStatusPendingApproval {
		return fmt.Errorf("order is not in pending_approval status")
	}
	// H-16: status flip + audit row written atomically.
	return s.orderOps.ApprovePendingOrderWithAudit(ctx, orderID, userID, "approved", comment)
}

// RejectOrderByAdmin rejects without org membership check (admin portal).
func (s *ApprovalService) RejectOrderByAdmin(ctx context.Context, orderID, userID, comment string) error {
	order, err := s.orderOps.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != modelsOrder.OrderStatusPendingApproval {
		return fmt.Errorf("order is not in pending_approval status")
	}
	return s.orderOps.CancelPendingApprovalOrderWithAudit(ctx, orderID, userID, "rejected", comment)
}

// ApproveOrder transitions an order from pending_approval to pending_confirmation.
func (s *ApprovalService) ApproveOrder(ctx context.Context, orderID, userID, comment string) error {
	order, err := s.orderOps.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != modelsOrder.OrderStatusPendingApproval {
		return fmt.Errorf("order is not in pending_approval status")
	}
	can, err := s.UserCanApproveOrder(ctx, userID, order)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("user is not authorized to approve this order")
	}
	return s.orderOps.ApprovePendingOrderWithAudit(ctx, orderID, userID, "approved", comment)
}

// RejectOrder transitions an order from pending_approval to cancelled and releases reserved stock.
func (s *ApprovalService) RejectOrder(ctx context.Context, orderID, userID, comment string) error {
	order, err := s.orderOps.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != modelsOrder.OrderStatusPendingApproval {
		return fmt.Errorf("order is not in pending_approval status")
	}
	can, err := s.UserCanApproveOrder(ctx, userID, order)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("user is not authorized to reject this order")
	}
	return s.orderOps.CancelPendingApprovalOrderWithAudit(ctx, orderID, userID, "rejected", comment)
}

// ApproveOrderWithModifications approves after optionally adjusting line items.
func (s *ApprovalService) ApproveOrderWithModifications(ctx context.Context, orderID, userID, comment string, items []modelsOrder.OrderItem) error {
	order, err := s.orderOps.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != modelsOrder.OrderStatusPendingApproval {
		return fmt.Errorf("order is not in pending_approval status")
	}
	can, err := s.UserCanApproveOrder(ctx, userID, order)
	if err != nil {
		return err
	}
	if !can {
		return fmt.Errorf("user is not authorized to approve this order")
	}
	if len(items) > 0 {
		order.Items = items
		var subtotal float64
		for _, it := range items {
			subtotal += float64(it.Quantity) * it.UnitPrice
		}
		order.Subtotal = subtotal
		order.TotalAmount = subtotal + order.TaxAmount + order.ShippingAmount
		if err := s.orderOps.Update(ctx, order); err != nil {
			return err
		}
	}
	if err := s.orderOps.ApprovePendingOrder(ctx, orderID); err != nil {
		return err
	}
	action := "approved"
	if len(items) > 0 {
		action = "modified"
	}
	return s.RecordApprovalAction(ctx, orderID, userID, action, comment)
}
