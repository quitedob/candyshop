package trade

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsTrade "candypro/api/internal/models/trade"
	"candypro/api/internal/pkg/crypto"
	"context"
	"errors"
	"strconv"
	"time"
)

// quotationReviewRepository abstracts the persistence dependency.
type quotationReviewRepository interface {
	Create(ctx context.Context, qr *modelsTrade.QuotationReview) error
	List(ctx context.Context, status string, page, limit int) ([]modelsTrade.QuotationReview, int64, error)
	GetByID(ctx context.Context, id uint) (*modelsTrade.QuotationReview, error)
	UpdateStatus(ctx context.Context, id uint, status, notes string) error
}

// quotationReviewAuditLogger records who performed a review decision (L1).
type quotationReviewAuditLogger interface {
	LogActivity(ctx context.Context, log *modelsCommon.ActivityLog) error
}

// QuotationReviewService provides business logic for the human-review queue.
type QuotationReviewService struct {
	repo  quotationReviewRepository
	audit quotationReviewAuditLogger
}

// NewQuotationReviewService creates a QuotationReviewService.
func NewQuotationReviewService(repo quotationReviewRepository, audit quotationReviewAuditLogger) *QuotationReviewService {
	return &QuotationReviewService{repo: repo, audit: audit}
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
// lives in the repository (only pending rows may be decided). operatorID records
// who performed the decision (L1); it is written to the audit log after a
// successful transition.
func (s *QuotationReviewService) Decide(ctx context.Context, id uint, status, notes, operatorID string) error {
	if status != modelsTrade.QuotationReviewStatusApproved && status != modelsTrade.QuotationReviewStatusRejected {
		return errors.New("invalid quotation review status")
	}
	if err := s.repo.UpdateStatus(ctx, id, status, notes); err != nil {
		return err
	}
	s.recordDecisionAudit(ctx, id, status, operatorID)
	return nil
}

// recordDecisionAudit writes an ActivityLog row so the approve/reject decision is
// attributable to an operator. Audit failures are non-fatal and never roll back
// the decision (matching the handler-level logActivityAudit convention).
func (s *QuotationReviewService) recordDecisionAudit(ctx context.Context, id uint, status, operatorID string) {
	if s.audit == nil || operatorID == "" {
		return
	}
	action := "quotation_review_approve"
	if status == modelsTrade.QuotationReviewStatusRejected {
		action = "quotation_review_reject"
	}
	uid := operatorID
	_ = s.audit.LogActivity(ctx, &modelsCommon.ActivityLog{
		ID:         crypto.GenerateID(),
		UserID:     &uid,
		Action:     action,
		EntityType: "quotation_review",
		EntityID:   strconv.FormatUint(uint64(id), 10),
		Details:    `{"newValue":"` + status + `"}`,
		CreatedAt:  time.Now(),
	})
}
