package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

// QuotationReviewRepository persists AI-submitted quotation reviews.
type QuotationReviewRepository struct {
	db *gorm.DB
}

// NewQuotationReviewRepository creates a new QuotationReviewRepository.
func NewQuotationReviewRepository(db *gorm.DB) *QuotationReviewRepository {
	return &QuotationReviewRepository{db: db}
}

// Create inserts a new quotation review.
func (r *QuotationReviewRepository) Create(ctx context.Context, qr *modelsTrade.QuotationReview) error {
	if qr.Status == "" {
		qr.Status = modelsTrade.QuotationReviewStatusPending
	}
	return r.db.WithContext(ctx).Create(qr).Error
}

// List returns quotation reviews, optionally filtered by status, newest first.
func (r *QuotationReviewRepository) List(ctx context.Context, status string, page, limit int) ([]modelsTrade.QuotationReview, int64, error) {
	q := r.db.WithContext(ctx).Model(&modelsTrade.QuotationReview{})
	if s := strings.TrimSpace(status); s != "" {
		q = q.Where("status = ?", s)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	var rows []modelsTrade.QuotationReview
	if err := q.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// GetByID returns a single quotation review.
func (r *QuotationReviewRepository) GetByID(ctx context.Context, id uint) (*modelsTrade.QuotationReview, error) {
	var qr modelsTrade.QuotationReview
	if err := r.db.WithContext(ctx).First(&qr, id).Error; err != nil {
		return nil, err
	}
	return &qr, nil
}

// UpdateStatus sets the review status and admin notes. Only pending rows may be
// decided; a conditional update makes concurrent double-decisions fail loudly.
func (r *QuotationReviewRepository) UpdateStatus(ctx context.Context, id uint, status, notes string) error {
	res := r.db.WithContext(ctx).
		Model(&modelsTrade.QuotationReview{}).
		Where("id = ? AND status = ?", id, modelsTrade.QuotationReviewStatusPending).
		Updates(map[string]any{
			"status":     status,
			"notes":      notes,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("quotation review not pending or not found")
	}
	return nil
}
