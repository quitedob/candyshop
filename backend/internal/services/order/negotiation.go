package order

import (
	"context"
	"errors"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"

	"gorm.io/gorm"
)

type negotiationRepository interface {
	FindByInquiryID(ctx context.Context, inquiryID string) ([]modelsOrder.NegotiationOffer, error)
	FindByID(ctx context.Context, id string) (*modelsOrder.NegotiationOffer, error)
	Create(ctx context.Context, offer *modelsOrder.NegotiationOffer) error
	Update(ctx context.Context, offer *modelsOrder.NegotiationOffer) error
	FindPendingByInquiryID(ctx context.Context, inquiryID string) (*modelsOrder.NegotiationOffer, error)
	// TransitionStatus 条件更新：仅当 status=fromStatus 时把它改为 toStatus。
	// 返回受影响行数；0 表示已被并发修改（A-4 TOCTOU 防护）。
	TransitionStatus(ctx context.Context, id, fromStatus, toStatus string, updatedAt time.Time) (int64, error)
}

// Sentinel errors returned by NegotiationService. Handlers translate these to
// stable user-facing error codes; the service-level err.Error() string is not
// exposed in API responses (C-10).
var (
	ErrNegotiationOfferNotFound   = errors.New("negotiation: offer not found")
	ErrNegotiationOfferNotPending = errors.New("negotiation: offer is not pending")
)

type NegotiationService struct {
	repo negotiationRepository
}

func NewNegotiationService(repo negotiationRepository) *NegotiationService {
	return &NegotiationService{repo: repo}
}

func (s *NegotiationService) GetOffers(ctx context.Context, inquiryID string) ([]modelsOrder.NegotiationOffer, error) {
	return s.repo.FindByInquiryID(ctx, inquiryID)
}

func (s *NegotiationService) CreateOffer(ctx context.Context, inquiryID, userID, senderType string, req CreateOfferRequest) (*modelsOrder.NegotiationOffer, error) {
	now := time.Now()
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	offer := &modelsOrder.NegotiationOffer{
		ID:           crypto.GenerateID(),
		InquiryID:    inquiryID,
		UserID:       userID,
		SenderType:   senderType,
		Status:       "pending",
		UnitPrice:    req.UnitPrice,
		Quantity:     req.Quantity,
		TotalAmount:  req.TotalAmount,
		Currency:     currency,
		Incoterms:    req.Incoterms,
		PaymentTerms: req.PaymentTerms,
		DeliveryDate: req.DeliveryDate,
		ValidUntil:   req.ValidUntil,
		Message:      req.Message,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.repo.Create(ctx, offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *NegotiationService) AcceptOffer(ctx context.Context, id, userID string) (*modelsOrder.NegotiationOffer, error) {
	now := time.Now()
	// A-4: 用条件 UPDATE … WHERE status='pending' 取代 read→check→Save，
	// 保证两个管理员并发 Accept 时只有第一个写入成功。
	rows, err := s.repo.TransitionStatus(ctx, id, "pending", "accepted", now)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		// R2 A-13: previously we called FindByID and discarded the result via
		// `_ = offer`. The lookup is needed to distinguish "missing" from
		// "non-pending"; just check Err and return the right sentinel without
		// binding the unused value.
		if _, ferr := s.repo.FindByID(ctx, id); ferr != nil {
			if errors.Is(ferr, gorm.ErrRecordNotFound) {
				return nil, ErrNegotiationOfferNotFound
			}
			return nil, ferr
		}
		return nil, ErrNegotiationOfferNotPending
	}
	return s.repo.FindByID(ctx, id)
}

func (s *NegotiationService) RejectOffer(ctx context.Context, id, userID string) (*modelsOrder.NegotiationOffer, error) {
	now := time.Now()
	rows, err := s.repo.TransitionStatus(ctx, id, "pending", "rejected", now)
	if err != nil {
		return nil, err
	}
	if rows == 0 {
		if _, ferr := s.repo.FindByID(ctx, id); ferr != nil {
			if errors.Is(ferr, gorm.ErrRecordNotFound) {
				return nil, ErrNegotiationOfferNotFound
			}
			return nil, ferr
		}
		return nil, ErrNegotiationOfferNotPending
	}
	return s.repo.FindByID(ctx, id)
}

type CreateOfferRequest struct {
	UnitPrice    float64    `json:"unitPrice"`
	Quantity     int        `json:"quantity"`
	TotalAmount  float64    `json:"totalAmount"`
	Currency     string     `json:"currency"`
	Incoterms    string     `json:"incoterms"`
	PaymentTerms string     `json:"paymentTerms"`
	DeliveryDate string     `json:"deliveryDate"`
	ValidUntil   *time.Time `json:"validUntil"`
	Message      string     `json:"message"`
}
