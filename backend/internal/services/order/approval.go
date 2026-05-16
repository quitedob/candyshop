package order

import (
	"context"
	"fmt"
	"time"

	modelsOrder "candypro/api/internal/models/order"
)

type approvalOrgRepo interface {
	FindByID(ctx context.Context, id uint) (*modelsOrder.BuyerOrganization, error)
	FindAll(ctx context.Context) ([]modelsOrder.BuyerOrganization, error)
	Create(ctx context.Context, org *modelsOrder.BuyerOrganization) error
	Update(ctx context.Context, org *modelsOrder.BuyerOrganization) error
}

type approvalMemberRepo interface {
	FindByUserID(ctx context.Context, userID string) ([]modelsOrder.OrgMember, error)
	FindByOrgID(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error)
	FindApprovers(ctx context.Context, orgID uint) ([]modelsOrder.OrgMember, error)
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
	ConfirmPendingOrder(ctx context.Context, id string, confirmedAt time.Time) error
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
func (s *ApprovalService) ShouldRequireApproval(ctx context.Context, userID string, orderTotal float64) (bool, *modelsOrder.BuyerOrganization, error) {
	members, err := s.memberRepo.FindByUserID(ctx, userID)
	if err != nil || len(members) == 0 {
		return false, nil, nil
	}
	for _, m := range members {
		if m.Role != modelsOrder.OrgRolePurchaser {
			continue
		}
		org, err := s.orgRepo.FindByID(ctx, m.OrganizationID)
		if err != nil {
			continue
		}
		if org.ApprovalThreshold > 0 && orderTotal > org.ApprovalThreshold {
			return true, org, nil
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

// ApproveOrder transitions an order from pending_approval to pending_confirmation.
func (s *ApprovalService) ApproveOrder(ctx context.Context, orderID, userID, comment string) error {
	order, err := s.orderOps.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != modelsOrder.OrderStatusPendingApproval {
		return fmt.Errorf("order is not in pending_approval status")
	}
	now := time.Now()
	if err := s.orderOps.ConfirmPendingOrder(ctx, orderID, now); err != nil {
		return err
	}
	return s.RecordApprovalAction(ctx, orderID, userID, "approved", comment)
}

// RejectOrder transitions an order from pending_approval to cancelled.
func (s *ApprovalService) RejectOrder(ctx context.Context, orderID, userID, comment string) error {
	order, err := s.orderOps.FindByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order.Status != modelsOrder.OrderStatusPendingApproval {
		return fmt.Errorf("order is not in pending_approval status")
	}
	order.Status = modelsOrder.OrderStatusCancelled
	if err := s.orderOps.Update(ctx, order); err != nil {
		return err
	}
	return s.RecordApprovalAction(ctx, orderID, userID, "rejected", comment)
}
