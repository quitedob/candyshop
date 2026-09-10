package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"time"

	"gorm.io/gorm"
)

// TradeRepository interface defines the methods for managing trade data
type TradeRepository interface {
	// Transactions
	CreateTransaction(ctx context.Context, transaction *modelsTrade.TradeTransaction) error
	GetTransactionByID(ctx context.Context, id uint) (*modelsTrade.TradeTransaction, error)
	GetTransactionByReference(ctx context.Context, ref string) (*modelsTrade.TradeTransaction, error)
	ListTransactionsByUserID(ctx context.Context, userID string, page, pageSize int) ([]modelsTrade.TradeTransaction, int64, error)
	ListAllTransactions(ctx context.Context, page, pageSize int, status string) ([]modelsTrade.TradeTransaction, int64, error)
	UpdateTransaction(ctx context.Context, transaction *modelsTrade.TradeTransaction) error
	UpdateTransactionStatus(ctx context.Context, id uint, fromStatus, toStatus string) error
	CountTransactionsByOrderID(ctx context.Context, orderID string) (int64, error)
	GetFirstTransactionByOrderID(ctx context.Context, orderID string) (*modelsTrade.TradeTransaction, error)

	// Documents
	CreateDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error
	GetDocumentByID(ctx context.Context, id uint) (*modelsTrade.TradeDocument, error)
	GetDocumentByNumber(ctx context.Context, docNumber string) (*modelsTrade.TradeDocument, error)
	ListDocumentsByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.TradeDocument, error)
	UpdateDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error
	UpdateDocumentStatus(ctx context.Context, id uint, fromStatus, toStatus string) error
	DeleteDocument(ctx context.Context, id uint) error

	// Compliance
	CreateCompliance(ctx context.Context, comp *modelsTrade.ComplianceRequirement) error
	GetComplianceByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ComplianceRequirement, error)
	UpdateCompliance(ctx context.Context, comp *modelsTrade.ComplianceRequirement) error
	UpdateComplianceStatus(ctx context.Context, id uint, status string) error
}

type tradeRepository struct {
	db *gorm.DB
}

// NewTradeRepository creates a new TradeRepository instance
func NewTradeRepository(db *gorm.DB) TradeRepository {
	return &tradeRepository{db: db}
}

// CreateTransaction creates a new trade transaction
func (r *tradeRepository) CreateTransaction(ctx context.Context, transaction *modelsTrade.TradeTransaction) error {
	return r.db.WithContext(ctx).Create(transaction).Error
}

// GetTransactionByID retrieved a transaction by its ID with all related documents
func (r *tradeRepository) GetTransactionByID(ctx context.Context, id uint) (*modelsTrade.TradeTransaction, error) {
	var transaction modelsTrade.TradeTransaction
	err := r.db.WithContext(ctx).
		Preload("Documents").
		First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetTransactionByReference retrieves a transaction by reference string
func (r *tradeRepository) GetTransactionByReference(ctx context.Context, ref string) (*modelsTrade.TradeTransaction, error) {
	var transaction modelsTrade.TradeTransaction
	err := r.db.WithContext(ctx).
		Preload("Documents").
		First(&transaction, "reference = ?", ref).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// ListTransactionsByUserID retrieves list of transactions for a user
func (r *tradeRepository) ListTransactionsByUserID(ctx context.Context, userID string, page, pageSize int) ([]modelsTrade.TradeTransaction, int64, error) {
	var transactions []modelsTrade.TradeTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsTrade.TradeTransaction{}).Where("user_id = ?", userID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.
		Preload("Documents").
		Offset(offset).
		Limit(pageSize).
		Order("created_at desc").
		Find(&transactions).Error

	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// UpdateTransaction updates a transaction
func (r *tradeRepository) UpdateTransaction(ctx context.Context, transaction *modelsTrade.TradeTransaction) error {
	return r.db.WithContext(ctx).Save(transaction).Error
}

// UpdateTransactionStatus transitions a transaction status with a guarded
// conditional UPDATE keyed on the currently-loaded status (M3).
func (r *tradeRepository) UpdateTransactionStatus(ctx context.Context, id uint, fromStatus, toStatus string) error {
	res := r.db.WithContext(ctx).Model(&modelsTrade.TradeTransaction{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(map[string]interface{}{
			"status":     toStatus,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrTradeStateMismatch
	}
	return nil
}

// CountTransactionsByOrderID returns how many trade rows reference the order (idempotency / dedupe).
func (r *tradeRepository) CountTransactionsByOrderID(ctx context.Context, orderID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&modelsTrade.TradeTransaction{}).
		Where("order_id = ?", orderID).
		Count(&n).Error
	return n, err
}

// GetFirstTransactionByOrderID 返回关联订单的第一条贸易主单（用于金额与订单对齐）
func (r *tradeRepository) GetFirstTransactionByOrderID(ctx context.Context, orderID string) (*modelsTrade.TradeTransaction, error) {
	var transaction modelsTrade.TradeTransaction
	err := r.db.WithContext(ctx).
		Where("order_id = ?", orderID).
		Order("id ASC").
		First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// CreateDocument creates a new trade document
func (r *tradeRepository) CreateDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error {
	return r.db.WithContext(ctx).Create(doc).Error
}

// GetDocumentByID gets document by ID
func (r *tradeRepository) GetDocumentByID(ctx context.Context, id uint) (*modelsTrade.TradeDocument, error) {
	var doc modelsTrade.TradeDocument
	err := r.db.WithContext(ctx).First(&doc, id).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// GetDocumentByNumber gets document by doc sequence string
func (r *tradeRepository) GetDocumentByNumber(ctx context.Context, docNumber string) (*modelsTrade.TradeDocument, error) {
	var doc modelsTrade.TradeDocument
	err := r.db.WithContext(ctx).First(&doc, "doc_number = ?", docNumber).Error
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocumentsByTransactionID gets all documents belonging to a trade transaction
func (r *tradeRepository) ListDocumentsByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.TradeDocument, error) {
	var docs []modelsTrade.TradeDocument
	err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Order("created_at asc").Find(&docs).Error
	return docs, err
}

// UpdateDocument limits update to a specific document block
func (r *tradeRepository) UpdateDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error {
	return r.db.WithContext(ctx).Save(doc).Error
}

// UpdateDocumentStatus transitions a document status with a guarded conditional
// UPDATE keyed on the currently-loaded status (M3).
func (r *tradeRepository) UpdateDocumentStatus(ctx context.Context, id uint, fromStatus, toStatus string) error {
	res := r.db.WithContext(ctx).Model(&modelsTrade.TradeDocument{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(map[string]interface{}{
			"status":     toStatus,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrTradeDocumentStateMismatch
	}
	return nil
}

// DeleteDocument removes a document by ID.
func (r *tradeRepository) DeleteDocument(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsTrade.TradeDocument{}, id).Error
}

// CreateCompliance creates a compliance req
func (r *tradeRepository) CreateCompliance(ctx context.Context, comp *modelsTrade.ComplianceRequirement) error {
	return r.db.WithContext(ctx).Create(comp).Error
}

// GetComplianceByTransactionID lists compliance checklists for a transaction
func (r *tradeRepository) GetComplianceByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ComplianceRequirement, error) {
	var comps []modelsTrade.ComplianceRequirement
	err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Find(&comps).Error
	return comps, err
}

// UpdateCompliance sets completion status for a compliance document
func (r *tradeRepository) UpdateCompliance(ctx context.Context, comp *modelsTrade.ComplianceRequirement) error {
	return r.db.WithContext(ctx).Save(comp).Error
}

// UpdateComplianceStatus updates only the status field to avoid zeroing other fields.
func (r *tradeRepository) UpdateComplianceStatus(ctx context.Context, id uint, status string) error {
	return r.db.WithContext(ctx).
		Model(&modelsTrade.ComplianceRequirement{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// ListAllTransactions retrieves all transactions with optional status filter (admin).
func (r *tradeRepository) ListAllTransactions(ctx context.Context, page, pageSize int, status string) ([]modelsTrade.TradeTransaction, int64, error) {
	var transactions []modelsTrade.TradeTransaction
	var total int64

	query := r.db.WithContext(ctx).Model(&modelsTrade.TradeTransaction{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.
		Preload("Documents").
		Offset(offset).
		Limit(pageSize).
		Order("created_at desc").
		Find(&transactions).Error
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}
