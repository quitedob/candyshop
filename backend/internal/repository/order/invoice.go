package order

import (
	modelsOrder "candypro/api/internal/models/order"
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

// FindByOrderID returns all invoices linked to an order.
func (r *InvoiceRepository) FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Invoice, error) {
	var invoices []modelsOrder.Invoice
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Order("created_at asc").Find(&invoices).Error; err != nil {
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

// Update saves changes to an invoice.
func (r *InvoiceRepository) Update(ctx context.Context, invoice *modelsOrder.Invoice) error {
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
	return fmt.Sprintf("inv-%d", time.Now().UnixNano())
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
