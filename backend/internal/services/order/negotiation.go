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
	// TransitionStatus 条件更新：仅当 id、inquiry_id、status 三者匹配时才写入。
	// inquiry_id 限定 offer 归属（H-3）；返回受影响行数，0 表示已被并发修改
	// （A-4 TOCTOU 防护）或询盘不匹配。
	TransitionStatus(ctx context.Context, id, inquiryID, fromStatus, toStatus string, updatedAt time.Time) (int64, error)
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

// AcceptOffer 把 pending offer 置为 accepted，并生成后续订单。
//
// H-3 所有权限定：offer 只属于一个 inquiry。调用方（handler）已先校验路径
// inquiry 归属于当前用户（customer 场景），这里再校验传入的 inquiryID 与
// offer 实际归属一致，防止调用方在询盘 X 上操作属于询盘 Y 的 offer；不匹配
// 时返回 ErrNegotiationOfferNotFound（404），不泄漏 offer 是否存在。
//
// A-4: 写入仍然是原子条件 UPDATE（WHERE id AND inquiry_id AND status），
// 上面的 FindByID 仅用于错误分类，不构成 read→check→Save 的 TOCTOU 模式。
func (s *NegotiationService) AcceptOffer(ctx context.Context, id, inquiryID string) (*modelsOrder.NegotiationOffer, error) {
	offer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNegotiationOfferNotFound
		}
		return nil, err
	}
	if offer.InquiryID != inquiryID {
		return nil, ErrNegotiationOfferNotFound
	}
	now := time.Now()
	// A-4: 用条件 UPDATE … WHERE status='pending' 取代 read→check→Save，
	// 保证两个管理员并发 Accept 时只有第一个写入成功。
	rows, err := s.repo.TransitionStatus(ctx, id, inquiryID, "pending", "accepted", now)
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

// RejectOffer 把 pending offer 置为 rejected。所有权限定同 AcceptOffer（H-3）。
func (s *NegotiationService) RejectOffer(ctx context.Context, id, inquiryID string) (*modelsOrder.NegotiationOffer, error) {
	offer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNegotiationOfferNotFound
		}
		return nil, err
	}
	if offer.InquiryID != inquiryID {
		return nil, ErrNegotiationOfferNotFound
	}
	now := time.Now()
	rows, err := s.repo.TransitionStatus(ctx, id, inquiryID, "pending", "rejected", now)
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
