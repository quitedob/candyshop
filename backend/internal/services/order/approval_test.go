package order_test

import (
	"context"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	orderSvc "candypro/api/internal/services/order"
)

type fakeApprovalOrderRepo struct {
	order *modelsOrder.Order
}

func (f *fakeApprovalOrderRepo) FindByID(_ context.Context, id string) (*modelsOrder.Order, error) {
	if f.order != nil && f.order.ID == id {
		copy := *f.order
		return &copy, nil
	}
	return nil, context.Canceled
}

func (f *fakeApprovalOrderRepo) ApprovePendingOrder(_ context.Context, id string) error {
	if f.order == nil || f.order.ID != id {
		return context.Canceled
	}
	if f.order.Status != modelsOrder.OrderStatusPendingApproval {
		return context.Canceled
	}
	f.order.Status = modelsOrder.OrderStatusPendingConfirm
	return nil
}

func (f *fakeApprovalOrderRepo) CancelPendingApprovalOrder(_ context.Context, orderID string) error {
	if f.order == nil || f.order.ID != orderID {
		return context.Canceled
	}
	f.order.Status = modelsOrder.OrderStatusCancelled
	return nil
}

func (f *fakeApprovalOrderRepo) Update(_ context.Context, order *modelsOrder.Order) error {
	f.order = order
	return nil
}

func (f *fakeApprovalOrderRepo) ApprovePendingOrderWithAudit(_ context.Context, id, _, _, _ string) error {
	return f.ApprovePendingOrder(context.Background(), id)
}

func (f *fakeApprovalOrderRepo) CancelPendingApprovalOrderWithAudit(_ context.Context, orderID, _, _, _ string) error {
	return f.CancelPendingApprovalOrder(context.Background(), orderID)
}

type fakeMemberRepo struct {
	purchaser modelsOrder.OrgMember
	approver  modelsOrder.OrgMember
}

func (f *fakeMemberRepo) FindByUserID(_ context.Context, userID string) ([]modelsOrder.OrgMember, error) {
	if userID == f.purchaser.UserID {
		return []modelsOrder.OrgMember{f.purchaser}, nil
	}
	if userID == f.approver.UserID {
		return []modelsOrder.OrgMember{f.approver}, nil
	}
	return nil, nil
}
func (f *fakeMemberRepo) FindByOrgID(context.Context, uint) ([]modelsOrder.OrgMember, error) {
	return nil, nil
}
func (f *fakeMemberRepo) FindApprovers(_ context.Context, orgID uint) ([]modelsOrder.OrgMember, error) {
	if orgID == f.approver.OrganizationID {
		return []modelsOrder.OrgMember{f.approver}, nil
	}
	return nil, nil
}

func (f *fakeMemberRepo) FindApproversForOrgs(_ context.Context, orgIDs []uint) (map[uint][]modelsOrder.OrgMember, error) {
	out := make(map[uint][]modelsOrder.OrgMember, len(orgIDs))
	for _, id := range orgIDs {
		if id == f.approver.OrganizationID {
			out[id] = []modelsOrder.OrgMember{f.approver}
		}
	}
	return out, nil
}
func (f *fakeMemberRepo) Create(context.Context, *modelsOrder.OrgMember) error { return nil }
func (f *fakeMemberRepo) Delete(context.Context, uint) error                   { return nil }

type fakeOrgRepoImpl struct{}

func (f *fakeOrgRepoImpl) FindByID(context.Context, uint) (*modelsOrder.BuyerOrganization, error) {
	return nil, nil
}
func (f *fakeOrgRepoImpl) FindByIDs(_ context.Context, ids []uint) ([]modelsOrder.BuyerOrganization, error) {
	return nil, nil
}
func (f *fakeOrgRepoImpl) FindAll(context.Context) ([]modelsOrder.BuyerOrganization, error) {
	return nil, nil
}
func (f *fakeOrgRepoImpl) Create(context.Context, *modelsOrder.BuyerOrganization) error { return nil }
func (f *fakeOrgRepoImpl) Update(context.Context, *modelsOrder.BuyerOrganization) error { return nil }

type fakeActionRepo struct{}

func (f *fakeActionRepo) Create(context.Context, *modelsOrder.ApprovalAction) error { return nil }
func (f *fakeActionRepo) FindByOrderID(context.Context, string) ([]modelsOrder.ApprovalAction, error) {
	return nil, nil
}

func TestApproveOrderPendingToPendingConfirm(t *testing.T) {
	order := &modelsOrder.Order{
		ID:     "ord-1",
		UserID: "buyer-1",
		Status: modelsOrder.OrderStatusPendingApproval,
	}
	repo := &fakeApprovalOrderRepo{order: order}
	members := &fakeMemberRepo{
		purchaser: modelsOrder.OrgMember{UserID: "buyer-1", OrganizationID: 1, Role: modelsOrder.OrgRolePurchaser},
		approver:  modelsOrder.OrgMember{UserID: "mgr-1", OrganizationID: 1, Role: modelsOrder.OrgRoleApprover},
	}
	svc := orderSvc.NewApprovalService(&fakeOrgRepoImpl{}, members, &fakeActionRepo{}, repo)

	if err := svc.ApproveOrder(context.Background(), "ord-1", "mgr-1", "ok"); err != nil {
		t.Fatalf("ApproveOrder failed: %v", err)
	}
	if order.Status != modelsOrder.OrderStatusPendingConfirm {
		t.Fatalf("expected pending_confirmation, got %s", order.Status)
	}
}

func TestRejectOrderPendingApproval(t *testing.T) {
	order := &modelsOrder.Order{
		ID:     "ord-2",
		UserID: "buyer-1",
		Status: modelsOrder.OrderStatusPendingApproval,
	}
	repo := &fakeApprovalOrderRepo{order: order}
	members := &fakeMemberRepo{
		purchaser: modelsOrder.OrgMember{UserID: "buyer-1", OrganizationID: 1, Role: modelsOrder.OrgRolePurchaser},
		approver:  modelsOrder.OrgMember{UserID: "mgr-1", OrganizationID: 1, Role: modelsOrder.OrgRoleApprover},
	}
	svc := orderSvc.NewApprovalService(&fakeOrgRepoImpl{}, members, &fakeActionRepo{}, repo)

	if err := svc.RejectOrder(context.Background(), "ord-2", "mgr-1", "budget exceeded"); err != nil {
		t.Fatalf("RejectOrder failed: %v", err)
	}
	if order.Status != modelsOrder.OrderStatusCancelled {
		t.Fatalf("expected cancelled, got %s", order.Status)
	}
}
