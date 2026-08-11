package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"errors"
	"strings"
	"time"
)

type invoiceRepository interface {
	FindAll(ctx context.Context, page, pageSize int, status string) ([]modelsOrder.Invoice, int64, error)
	FindByID(ctx context.Context, id string) (*modelsOrder.Invoice, error)
	FindByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Invoice, error)
	FindByTradeID(ctx context.Context, tradeID uint) ([]modelsOrder.Invoice, error)
	Create(ctx context.Context, invoice *modelsOrder.Invoice) error
	Update(ctx context.Context, invoice *modelsOrder.Invoice) error
	Delete(ctx context.Context, id string) error
	Stats(ctx context.Context) (map[string]int64, error)
	SumByStatus(ctx context.Context, status string) (float64, error)
	OverdueCount(ctx context.Context) (int64, error)
	OverdueTotal(ctx context.Context) (float64, error)
	FindByStatus(ctx context.Context, status string, page, limit int) ([]modelsOrder.Invoice, int64, error)
	FindByUserID(ctx context.Context, userID string) ([]modelsOrder.Invoice, error)
}

// orderByIDReader 按 ID 读取订单（发票从订单派生时用；可传 nil）
type orderByIDReader interface {
	FindByID(ctx context.Context, id string) (*modelsOrder.Order, error)
}

// documentAdjustmentWriter 写入单证调整审计（可传 nil 则跳过审计）
type documentAdjustmentWriter interface {
	Create(ctx context.Context, row *modelsOrder.DocumentAdjustment) error
}

// InvoiceService provides invoice business logic.
type InvoiceService struct {
	repo   invoiceRepository
	orders orderByIDReader
	adj    documentAdjustmentWriter
}

// NewInvoiceService creates an InvoiceService（orders/adj 可为 nil）
func NewInvoiceService(repo invoiceRepository, orders orderByIDReader, adj documentAdjustmentWriter) *InvoiceService {
	return &InvoiceService{repo: repo, orders: orders, adj: adj}
}

// ListInvoices returns paginated invoices.
func (s *InvoiceService) ListInvoices(ctx context.Context, page, limit int, status string) ([]modelsOrder.Invoice, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.FindAll(ctx, page, limit, status)
}

// GetInvoice returns a single invoice by ID.
func (s *InvoiceService) GetInvoice(ctx context.Context, id string) (*modelsOrder.Invoice, error) {
	return s.repo.FindByID(ctx, id)
}

// GetByUserID returns all invoices for a user's orders in one query.
func (s *InvoiceService) GetByUserID(ctx context.Context, userID string) ([]modelsOrder.Invoice, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// GetByOrderID returns all invoices for an order.
func (s *InvoiceService) GetByOrderID(ctx context.Context, orderID string) ([]modelsOrder.Invoice, error) {
	return s.repo.FindByOrderID(ctx, orderID)
}

// GetByTradeID returns all invoices for a trade transaction.
func (s *InvoiceService) GetByTradeID(ctx context.Context, tradeID uint) ([]modelsOrder.Invoice, error) {
	return s.repo.FindByTradeID(ctx, tradeID)
}

// CreateInvoice creates a new invoice with defaults applied.
func (s *InvoiceService) CreateInvoice(ctx context.Context, invoice *modelsOrder.Invoice) error {
	if invoice.Type == "" {
		invoice.Type = modelsOrder.InvoiceTypeCommercial
	}
	if invoice.Status == "" {
		invoice.Status = modelsOrder.InvoiceStatusDraft
	}
	if invoice.Currency == "" {
		invoice.Currency = "USD"
	}
	// H9: preserve a caller-computed total (order-derived invoices already
	// include ShippingAmount in TotalAmount — see CreateInvoiceFromOrder);
	// only fall back to amount+tax when no total was supplied, so derived
	// invoices don't silently under-bill freight.
	if invoice.TotalAmount == 0 {
		invoice.TotalAmount = invoice.Amount + invoice.TaxAmount
	}
	return s.repo.Create(ctx, invoice)
}

// UpdateInvoice saves changes to an existing invoice.
func (s *InvoiceService) UpdateInvoice(ctx context.Context, invoice *modelsOrder.Invoice) error {
	// H9: same guard as CreateInvoice — don't clobber a total that already
	// carries a freight component with a shipping-free recomputation.
	if invoice.TotalAmount == 0 {
		invoice.TotalAmount = invoice.Amount + invoice.TaxAmount
	}
	return s.repo.Update(ctx, invoice)
}

// SendInvoice marks an invoice as sent.
func (s *InvoiceService) SendInvoice(ctx context.Context, id string) (*modelsOrder.Invoice, error) {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if invoice.Status == modelsOrder.InvoiceStatusVoided {
		return nil, errors.New("cannot send a voided invoice")
	}
	now := time.Now()
	invoice.Status = modelsOrder.InvoiceStatusSent
	invoice.SentAt = &now
	if err := s.repo.Update(ctx, invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

// DeleteInvoice removes an invoice (only drafts may be deleted).
func (s *InvoiceService) DeleteInvoice(ctx context.Context, id string) error {
	invoice, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if strings.ToLower(invoice.Status) != modelsOrder.InvoiceStatusDraft {
		return errors.New("only draft invoices can be deleted")
	}
	return s.repo.Delete(ctx, id)
}

// GetStats returns counts grouped by status.
func (s *InvoiceService) GetStats(ctx context.Context) (map[string]int64, error) {
	return s.repo.Stats(ctx)
}

// SumByStatus returns total amount for invoices of a given status.
func (s *InvoiceService) SumByStatus(ctx context.Context, status string) (float64, error) {
	return s.repo.SumByStatus(ctx, status)
}

// GetOverdueCount counts overdue invoices.
func (s *InvoiceService) GetOverdueCount(ctx context.Context) (int64, error) {
	return s.repo.OverdueCount(ctx)
}

// GetOverdueTotal returns total amount of overdue invoices.
func (s *InvoiceService) GetOverdueTotal(ctx context.Context) (float64, error) {
	return s.repo.OverdueTotal(ctx)
}

// GetInvoicesByStatus returns invoices filtered by status with pagination.
func (s *InvoiceService) GetInvoicesByStatus(ctx context.Context, status string, page, limit int) ([]modelsOrder.Invoice, int64, error) {
	return s.repo.FindByStatus(ctx, status, page, limit)
}

// SumOverdue returns total amount of overdue invoices (alias for handler compatibility).
func (s *InvoiceService) SumOverdue(ctx context.Context) (float64, error) {
	return s.repo.OverdueTotal(ctx)
}

// FindByStatus returns invoices filtered by status with pagination (alias for handler compatibility).
func (s *InvoiceService) FindByStatus(ctx context.Context, status string, page, limit int) ([]modelsOrder.Invoice, int64, error) {
	return s.repo.FindByStatus(ctx, status, page, limit)
}
