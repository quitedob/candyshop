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
}

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
	offer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("offer not found")
		}
		return nil, err
	}
	if offer.Status != "pending" {
		return nil, errors.New("offer is not pending")
	}
	offer.Status = "accepted"
	offer.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, offer); err != nil {
		return nil, err
	}
	return offer, nil
}

func (s *NegotiationService) RejectOffer(ctx context.Context, id, userID string) (*modelsOrder.NegotiationOffer, error) {
	offer, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("offer not found")
		}
		return nil, err
	}
	if offer.Status != "pending" {
		return nil, errors.New("offer is not pending")
	}
	offer.Status = "rejected"
	offer.UpdatedAt = time.Now()
	if err := s.repo.Update(ctx, offer); err != nil {
		return nil, err
	}
	return offer, nil
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
