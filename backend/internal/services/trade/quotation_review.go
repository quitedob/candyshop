package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"errors"
)

// quotationReviewRepository abstracts the persistence dependency.
type quotationReviewRepository interface {
	Create(ctx context.Context, qr *modelsTrade.QuotationReview) error
	List(ctx context.Context, status string, page, limit int) ([]modelsTrade.QuotationReview, int64, error)
	GetByID(ctx context.Context, id uint) (*modelsTrade.QuotationReview, error)
	UpdateStatus(ctx context.Context, id uint, status, notes string) error
}

// QuotationReviewService provides business logic for the human-review queue.
type QuotationReviewService struct {
	repo quotationReviewRepository
}

// NewQuotationReviewService creates a QuotationReviewService.
func NewQuotationReviewService(repo quotationReviewRepository) *QuotationReviewService {
	return &QuotationReviewService{repo: repo}
}

// SaveQuotationReview persists a review queued by the AI tool. It satisfies the
// einotool.QuotationReviewSaver interface used by NewQuotationHumanReviewTool.
func (s *QuotationReviewService) SaveQuotationReview(ctx context.Context, qr *modelsTrade.QuotationReview) error {
	if qr.Status == "" {
		qr.Status = modelsTrade.QuotationReviewStatusPending
	}
	return s.repo.Create(ctx, qr)
}

// List returns quotation reviews, optionally filtered by status.
func (s *QuotationReviewService) List(ctx context.Context, status string, page, limit int) ([]modelsTrade.QuotationReview, int64, error) {
	return s.repo.List(ctx, status, page, limit)
}

// GetByID returns a single quotation review.
func (s *QuotationReviewService) GetByID(ctx context.Context, id uint) (*modelsTrade.QuotationReview, error) {
	return s.repo.GetByID(ctx, id)
}

// Decide approves or rejects a pending quotation review. The transition guard
// lives in the repository (only pending rows may be decided).
func (s *QuotationReviewService) Decide(ctx context.Context, id uint, status, notes string) error {
	if status != modelsTrade.QuotationReviewStatusApproved && status != modelsTrade.QuotationReviewStatusRejected {
		return errors.New("invalid quotation review status")
	}
	return s.repo.UpdateStatus(ctx, id, status, notes)
}
