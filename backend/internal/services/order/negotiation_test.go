package order_test

import (
	"context"
	"errors"
	"testing"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	orderSvc "candypro/api/internal/services/order"

	"gorm.io/gorm"
)

// fakeNegotiationRepo implements the (unexported) negotiationRepository interface
// for service-level tests. FindByID returns gorm.ErrRecordNotFound for missing
// offers so the service can classify not-found via errors.Is.
type fakeNegotiationRepo struct {
	offers []modelsOrder.NegotiationOffer
}

func (r *fakeNegotiationRepo) FindByInquiryID(_ context.Context, inquiryID string) ([]modelsOrder.NegotiationOffer, error) {
	var out []modelsOrder.NegotiationOffer
	for _, o := range r.offers {
		if o.InquiryID == inquiryID {
			out = append(out, o)
		}
	}
	return out, nil
}

func (r *fakeNegotiationRepo) FindByID(_ context.Context, id string) (*modelsOrder.NegotiationOffer, error) {
	for i := range r.offers {
		if r.offers[i].ID == id {
			cp := r.offers[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *fakeNegotiationRepo) Create(_ context.Context, offer *modelsOrder.NegotiationOffer) error {
	r.offers = append(r.offers, *offer)
	return nil
}

func (r *fakeNegotiationRepo) Update(_ context.Context, offer *modelsOrder.NegotiationOffer) error {
	for i := range r.offers {
		if r.offers[i].ID == offer.ID {
			r.offers[i] = *offer
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (r *fakeNegotiationRepo) FindPendingByInquiryID(_ context.Context, inquiryID string) (*modelsOrder.NegotiationOffer, error) {
	for i := range r.offers {
		if r.offers[i].InquiryID == inquiryID && r.offers[i].Status == "pending" {
			cp := r.offers[i]
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// TransitionStatus scopes by id AND inquiry_id (H-3), matching the repository.
func (r *fakeNegotiationRepo) TransitionStatus(_ context.Context, id, inquiryID, fromStatus, toStatus string, updatedAt time.Time) (int64, error) {
	for i := range r.offers {
		if r.offers[i].ID == id && r.offers[i].InquiryID == inquiryID {
			if r.offers[i].Status != fromStatus {
				return 0, nil
			}
			r.offers[i].Status = toStatus
			r.offers[i].UpdatedAt = updatedAt
			return 1, nil
		}
	}
	return 0, nil
}

func statusOf(offers []modelsOrder.NegotiationOffer, id string) string {
	for _, o := range offers {
		if o.ID == id {
			return o.Status
		}
	}
	return ""
}

func TestNegotiationAcceptOffer_Success(t *testing.T) {
	now := time.Now()
	repo := &fakeNegotiationRepo{offers: []modelsOrder.NegotiationOffer{
		{ID: "offer-1", InquiryID: "inq-1", UserID: "admin-1", SenderType: "admin", Status: "pending", CreatedAt: now, UpdatedAt: now},
	}}
	svc := orderSvc.NewNegotiationService(repo)

	offer, err := svc.AcceptOffer(context.Background(), "offer-1", "inq-1")
	if err != nil {
		t.Fatalf("AcceptOffer failed: %v", err)
	}
	if offer.Status != "accepted" {
		t.Fatalf("expected accepted, got %s", offer.Status)
	}
}

// TestNegotiationAcceptOffer_InquiryMismatch H-3: 在询盘 X 上接受属于询盘 Y 的
// offer 必须返回 ErrNegotiationOfferNotFound 且不改状态。
func TestNegotiationAcceptOffer_InquiryMismatch(t *testing.T) {
	now := time.Now()
	repo := &fakeNegotiationRepo{offers: []modelsOrder.NegotiationOffer{
		{ID: "offer-2", InquiryID: "inq-y", UserID: "admin-1", SenderType: "admin", Status: "pending", CreatedAt: now, UpdatedAt: now},
	}}
	svc := orderSvc.NewNegotiationService(repo)

	_, err := svc.AcceptOffer(context.Background(), "offer-2", "inq-x")
	if !errors.Is(err, orderSvc.ErrNegotiationOfferNotFound) {
		t.Fatalf("expected ErrNegotiationOfferNotFound, got %v", err)
	}
	if got := statusOf(repo.offers, "offer-2"); got != "pending" {
		t.Fatalf("expected offer status still pending, got %s", got)
	}
}

func TestNegotiationAcceptOffer_NotPending(t *testing.T) {
	now := time.Now()
	repo := &fakeNegotiationRepo{offers: []modelsOrder.NegotiationOffer{
		{ID: "offer-3", InquiryID: "inq-1", UserID: "admin-1", SenderType: "admin", Status: "accepted", CreatedAt: now, UpdatedAt: now},
	}}
	svc := orderSvc.NewNegotiationService(repo)

	_, err := svc.AcceptOffer(context.Background(), "offer-3", "inq-1")
	if !errors.Is(err, orderSvc.ErrNegotiationOfferNotPending) {
		t.Fatalf("expected ErrNegotiationOfferNotPending, got %v", err)
	}
}

func TestNegotiationAcceptOffer_NotFound(t *testing.T) {
	svc := orderSvc.NewNegotiationService(&fakeNegotiationRepo{})

	_, err := svc.AcceptOffer(context.Background(), "missing", "inq-1")
	if !errors.Is(err, orderSvc.ErrNegotiationOfferNotFound) {
		t.Fatalf("expected ErrNegotiationOfferNotFound, got %v", err)
	}
}

func TestNegotiationRejectOffer_Success(t *testing.T) {
	now := time.Now()
	repo := &fakeNegotiationRepo{offers: []modelsOrder.NegotiationOffer{
		{ID: "offer-4", InquiryID: "inq-1", UserID: "admin-1", SenderType: "admin", Status: "pending", CreatedAt: now, UpdatedAt: now},
	}}
	svc := orderSvc.NewNegotiationService(repo)

	offer, err := svc.RejectOffer(context.Background(), "offer-4", "inq-1")
	if err != nil {
		t.Fatalf("RejectOffer failed: %v", err)
	}
	if offer.Status != "rejected" {
		t.Fatalf("expected rejected, got %s", offer.Status)
	}
}

// TestNegotiationRejectOffer_InquiryMismatch H-3: reject 路径之前完全没有归属
// 校验，现在跨询盘 reject 必须被拒绝。
func TestNegotiationRejectOffer_InquiryMismatch(t *testing.T) {
	now := time.Now()
	repo := &fakeNegotiationRepo{offers: []modelsOrder.NegotiationOffer{
		{ID: "offer-5", InquiryID: "inq-y", UserID: "admin-1", SenderType: "admin", Status: "pending", CreatedAt: now, UpdatedAt: now},
	}}
	svc := orderSvc.NewNegotiationService(repo)

	_, err := svc.RejectOffer(context.Background(), "offer-5", "inq-x")
	if !errors.Is(err, orderSvc.ErrNegotiationOfferNotFound) {
		t.Fatalf("expected ErrNegotiationOfferNotFound, got %v", err)
	}
	if got := statusOf(repo.offers, "offer-5"); got != "pending" {
		t.Fatalf("expected offer status still pending, got %s", got)
	}
}

func TestNegotiationRejectOffer_NotFound(t *testing.T) {
	svc := orderSvc.NewNegotiationService(&fakeNegotiationRepo{})

	_, err := svc.RejectOffer(context.Background(), "missing", "inq-1")
	if !errors.Is(err, orderSvc.ErrNegotiationOfferNotFound) {
		t.Fatalf("expected ErrNegotiationOfferNotFound, got %v", err)
	}
}
