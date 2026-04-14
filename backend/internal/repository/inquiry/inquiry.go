package inquiry

import (
	modelsProduct "candypro/api/internal/models/product"
)

import (
	"context"

	"gorm.io/gorm"
)

// InquiryRepository handles inquiry data operations
type InquiryRepository struct {
	db *gorm.DB
}

// NewInquiryRepository creates a new InquiryRepository
func NewInquiryRepository(db *gorm.DB) *InquiryRepository {
	return &InquiryRepository{db: db}
}

// Create creates a new inquiry
func (r *InquiryRepository) Create(ctx context.Context, inquiry *modelsProduct.Inquiry) error {
	return r.db.WithContext(ctx).Create(inquiry).Error
}

// FindByID returns an inquiry by ID
func (r *InquiryRepository) FindByID(ctx context.Context, id string) (*modelsProduct.Inquiry, error) {
	var inquiry modelsProduct.Inquiry
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&inquiry).Error; err != nil {
		return nil, err
	}
	return &inquiry, nil
}

// FindAll returns all inquiries with pagination
func (r *InquiryRepository) FindAll(ctx context.Context, page, limit int) ([]modelsProduct.Inquiry, int64, error) {
	var inquiries []modelsProduct.Inquiry
	var total int64

	if err := r.db.WithContext(ctx).Model(&modelsProduct.Inquiry{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&inquiries).Error; err != nil {
		return nil, 0, err
	}

	return inquiries, total, nil
}

// FindByStatus returns inquiries by status
func (r *InquiryRepository) FindByStatus(ctx context.Context, status string, page, limit int) ([]modelsProduct.Inquiry, int64, error) {
	var inquiries []modelsProduct.Inquiry
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsProduct.Inquiry{}).Where("status = ?", status)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC").Offset(offset).Limit(limit).Find(&inquiries).Error; err != nil {
		return nil, 0, err
	}

	return inquiries, total, nil
}

// UpdateStatus updates inquiry status
func (r *InquiryRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	return r.db.WithContext(ctx).Model(&modelsProduct.Inquiry{}).Where("id = ?", id).Update("status", status).Error
}

// Update updates an inquiry record.
func (r *InquiryRepository) Update(ctx context.Context, inquiry *modelsProduct.Inquiry) error {
	return r.db.WithContext(ctx).Save(inquiry).Error
}

// Delete deletes an inquiry
func (r *InquiryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&modelsProduct.Inquiry{}, "id = ?", id).Error
}

// FindByUserID returns all inquiries for a specific user with pagination
func (r *InquiryRepository) FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsProduct.Inquiry, int64, error) {
	var inquiries []modelsProduct.Inquiry
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsProduct.Inquiry{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&inquiries).Error; err != nil {
		return nil, 0, err
	}

	return inquiries, total, nil
}

// CountAll returns total inquiry count.
func (r *InquiryRepository) CountAll(ctx context.Context) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&modelsProduct.Inquiry{}).Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// CountByStatus returns inquiry count with status.
func (r *InquiryRepository) CountByStatus(ctx context.Context, status string) (int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).
		Model(&modelsProduct.Inquiry{}).
		Where("status = ?", status).
		Count(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

// FindRecent returns latest inquiries.
func (r *InquiryRepository) FindRecent(ctx context.Context, limit int) ([]modelsProduct.Inquiry, error) {
	var inquiries []modelsProduct.Inquiry
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Find(&inquiries).Error; err != nil {
		return nil, err
	}
	return inquiries, nil
}

// ConversionByMonth returns monthly inquiry conversion data (total, won) for the last N months.
func (r *InquiryRepository) ConversionByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	type row struct {
		Month string
		Total int64
		Won   int64
	}
	var rows []row
	if err := r.db.WithContext(ctx).
		Model(&modelsProduct.Inquiry{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*) AS total, COUNT(*) FILTER (WHERE status = 'won') AS won").
		Where("created_at >= NOW() - INTERVAL '? months'", months).
		Group("month").
		Order("month").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, len(rows))
	for i, row := range rows {
		result[i] = map[string]interface{}{
			"month": row.Month,
			"total": row.Total,
			"won":   row.Won,
		}
	}
	return result, nil
}
