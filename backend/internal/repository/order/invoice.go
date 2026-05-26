package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/crypto"
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InvoiceRepository handles invoice persistence.
type InvoiceRepository struct {
	db *gorm.DB
}

// NewInvoiceRepository creates a new InvoiceRepository.
func NewInvoiceRepository(db *gorm.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// FindAll returns paginated invoices with optional status filter.
func (r *InvoiceRepository) FindAll(ctx context.Context, page, pageSize int, status string) ([]modelsOrder.Invoice, int64, error) {
	var invoices []modelsOrder.Invoice
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.Invoice{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at desc").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}
	return invoices, total, nil
}

// FindByID returns an invoice by primary key.
func (r *InvoiceRepository) FindByID(ctx context.Context, id string) (*modelsOrder.Invoice, error) {
	var invoice modelsOrder.Invoice
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&invoice).Error; err != nil {
		return nil, err
	}
	return &invoice, nil
}

// FindByUserID returns all invoices for orders owned by a user (single query, no N+1).
func (r *InvoiceRepository) FindByUserID(ctx context.Context, userID string) ([]modelsOrder.Invoice, error) {
	var invoices []modelsOrder.Invoice
	err := r.db.WithContext(ctx).
		Joins("JOIN orders ON orders.id = invoices.order_id").
		Where("orders.user_id = ?", userID).
		Order("invoices.created_at DESC").
		Find(&invoices).Error
	return invoices, err
}

// FindByOrderID returns all invoices linked to an order.
func (r *InvoiceRepository) FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Invoice, error) {
	var invoices []modelsOrder.Invoice
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at asc").Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// FindByTradeID returns all invoices linked to a trade transaction.
func (r *InvoiceRepository) FindByTradeID(ctx context.Context, tradeID uint) ([]modelsOrder.Invoice, error) {
	var invoices []modelsOrder.Invoice
	if err := r.db.WithContext(ctx).Where("trade_id = ?", tradeID).Order("created_at asc").Find(&invoices).Error; err != nil {
		return nil, err
	}
	return invoices, nil
}

// Create inserts a new invoice, generating an ID and invoice number if absent.
func (r *InvoiceRepository) Create(ctx context.Context, invoice *modelsOrder.Invoice) error {
	if invoice.ID == "" {
		invoice.ID = generateInvoiceID()
	}
	if invoice.InvoiceNo == "" {
		invoice.InvoiceNo = generateInvoiceNo(invoice.Type)
	}
	return r.db.WithContext(ctx).Create(invoice).Error
}

// Update saves changes to an invoice. M-9: bumps Version on every save.
func (r *InvoiceRepository) Update(ctx context.Context, invoice *modelsOrder.Invoice) error {
	if invoice != nil {
		invoice.Version++
	}
	return r.db.WithContext(ctx).Save(invoice).Error
}

// Delete removes an invoice by ID.
func (r *InvoiceRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&modelsOrder.Invoice{}).Error
}

// Stats returns summary counts by status.
func (r *InvoiceRepository) Stats(ctx context.Context) (map[string]int64, error) {
	type result struct {
		Status string
		Count  int64
	}
	var rows []result
	if err := r.db.WithContext(ctx).Model(&modelsOrder.Invoice{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[string]int64, len(rows))
	for _, row := range rows {
		m[row.Status] = row.Count
	}
	return m, nil
}

func generateInvoiceID() string {
	return crypto.GenerateID()
}

func generateInvoiceNo(invoiceType string) string {
	prefix := "CI"
	switch invoiceType {
	case modelsOrder.InvoiceTypeProforma:
		prefix = "PI"
	case modelsOrder.InvoiceTypeCreditNote:
		prefix = "CN"
	}
	return fmt.Sprintf("%s-%s-%06d", prefix, time.Now().UTC().Format("200601"), time.Now().UnixNano()%1000000)
}

// SumByStatus returns total amount for invoices of a given status.
func (r *InvoiceRepository) SumByStatus(ctx context.Context, status string) (float64, error) {
	var result struct {
		Amount float64
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Invoice{}).
		Where("status = ?", status).
		Select("COALESCE(SUM(total_amount), 0) AS amount").
		Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Amount, nil
}

// OverdueCount counts invoices past their due date with status not 'paid' or 'voided'.
func (r *InvoiceRepository) OverdueCount(ctx context.Context) (int64, error) {
	var result struct {
		Count int64
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Invoice{}).
		Where("due_date IS NOT NULL AND due_date < NOW() AND status NOT IN ?", []string{"paid", "voided"}).
		Select("COUNT(*) AS count").
		Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Count, nil
}

// OverdueTotal returns total amount of overdue invoices.
func (r *InvoiceRepository) OverdueTotal(ctx context.Context) (float64, error) {
	var result struct {
		Amount float64
	}
	if err := r.db.WithContext(ctx).
		Model(&modelsOrder.Invoice{}).
		Where("due_date IS NOT NULL AND due_date < NOW() AND status NOT IN ?", []string{"paid", "voided"}).
		Select("COALESCE(SUM(total_amount), 0) AS amount").
		Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.Amount, nil
}

// FindByStatus returns invoices filtered by status with pagination.
func (r *InvoiceRepository) FindByStatus(ctx context.Context, status string, page, limit int) ([]modelsOrder.Invoice, int64, error) {
	var invoices []modelsOrder.Invoice
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.Invoice{}).Where("status = ?", status)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at desc").Find(&invoices).Error; err != nil {
		return nil, 0, err
	}
	return invoices, total, nil
}
