package trade

import (
	modelsOrder "candypro/api/internal/models/order"
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

type tradeRepository interface {
	CreateTransaction(ctx context.Context, transaction *modelsTrade.TradeTransaction) error
	GetTransactionByID(ctx context.Context, id uint) (*modelsTrade.TradeTransaction, error)
	ListTransactionsByUserID(ctx context.Context, userID string, page, pageSize int) ([]modelsTrade.TradeTransaction, int64, error)
	ListAllTransactions(ctx context.Context, page, pageSize int, status string) ([]modelsTrade.TradeTransaction, int64, error)
	UpdateTransaction(ctx context.Context, transaction *modelsTrade.TradeTransaction) error
	CountTransactionsByOrderID(ctx context.Context, orderID string) (int64, error)
	GetFirstTransactionByOrderID(ctx context.Context, orderID string) (*modelsTrade.TradeTransaction, error)
	CreateDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error
	GetDocumentByID(ctx context.Context, id uint) (*modelsTrade.TradeDocument, error)
	ListDocumentsByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.TradeDocument, error)
	UpdateDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error
	DeleteDocument(ctx context.Context, id uint) error
	GetComplianceByTransactionID(ctx context.Context, transactionID uint) ([]modelsTrade.ComplianceRequirement, error)
	UpdateCompliance(ctx context.Context, comp *modelsTrade.ComplianceRequirement) error
	UpdateComplianceStatus(ctx context.Context, id uint, status string) error
}

type TradeService struct {
	repo tradeRepository
}

func NewTradeService(repo tradeRepository) *TradeService {
	return &TradeService{repo: repo}
}

// CreateTransaction starts a new transaction
func (s *TradeService) CreateTransaction(ctx context.Context, trans *modelsTrade.TradeTransaction) error {
	// Set initial status to draft if not provided
	if trans.Status == "" {
		trans.Status = string(modelsTrade.TradeStatusDraft)
	}
	if strings.TrimSpace(trans.Currency) == "" {
		trans.Currency = "USD"
	}
	if strings.TrimSpace(trans.Reference) == "" {
		trans.Reference = generateTradeReference()
	}
	if strings.TrimSpace(trans.Terms) == "" {
		trans.Terms = "FOB"
	}
	return s.repo.CreateTransaction(ctx, trans)
}

// GetTransaction retrieves a transaction and its docs
func (s *TradeService) GetTransaction(ctx context.Context, id uint) (*modelsTrade.TradeTransaction, error) {
	return s.repo.GetTransactionByID(ctx, id)
}

// ListTransactions retrieves paginated transactions for a user
func (s *TradeService) ListTransactions(ctx context.Context, userID string, page, limit int) ([]modelsTrade.TradeTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.repo.ListTransactionsByUserID(ctx, userID, page, limit)
}

// UpdateTransaction status
func (s *TradeService) UpdateTransaction(ctx context.Context, trans *modelsTrade.TradeTransaction) error {
	return s.repo.UpdateTransaction(ctx, trans)
}

// HasTransactionForOrder reports whether a trade record already exists for the order.
func (s *TradeService) HasTransactionForOrder(ctx context.Context, orderID string) (bool, error) {
	n, err := s.repo.CountTransactionsByOrderID(ctx, orderID)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SyncTradeTotalFromOrder 将贸易主单总金额/币种与订单对齐（订单财务变更后调用）
func (s *TradeService) SyncTradeTotalFromOrder(ctx context.Context, order *modelsOrder.Order) error {
	if order == nil || strings.TrimSpace(order.ID) == "" {
		return errors.New("invalid order")
	}
	tx, err := s.repo.GetFirstTransactionByOrderID(ctx, order.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	tx.TotalAmount = order.TotalAmount
	if c := strings.TrimSpace(order.Currency); c != "" {
		tx.Currency = c
	}
	return s.repo.UpdateTransaction(ctx, tx)
}

// AddDocument attaches a new document (PI, CI, etc) to a transaction
func (s *TradeService) AddDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error {
	doc.Status = string(modelsTrade.TradeStatusDraft)
	return s.repo.CreateDocument(ctx, doc)
}

// GetDocument retrieves a document
func (s *TradeService) GetDocument(ctx context.Context, id uint) (*modelsTrade.TradeDocument, error) {
	return s.repo.GetDocumentByID(ctx, id)
}

// UpdateDocument updates an existing document
func (s *TradeService) UpdateDocument(ctx context.Context, doc *modelsTrade.TradeDocument) error {
	return s.repo.UpdateDocument(ctx, doc)
}

func generateTradeReference() string {
	return fmt.Sprintf("TRD-%s-%d", time.Now().UTC().Format("20060102"), time.Now().UTC().UnixNano()%1000000)
}

// ListAllTransactions retrieves all transactions for admin (paginated, filterable).
func (s *TradeService) ListAllTransactions(ctx context.Context, page, limit int, status string) ([]modelsTrade.TradeTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 50 {
		limit = 10
	}
	return s.repo.ListAllTransactions(ctx, page, limit, status)
}

// UpdateTransactionStatus updates a transaction's status with state machine validation.
func (s *TradeService) UpdateTransactionStatus(ctx context.Context, id uint, status string) error {
	trans, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return err
	}
	prev := strings.ToUpper(strings.TrimSpace(trans.Status))
	next := strings.ToUpper(strings.TrimSpace(status))
	if allowed, known := modelsTrade.ValidTradeStatusTransitions[prev]; known {
		if !allowed[next] {
			return fmt.Errorf("cannot transition trade from '%s' to '%s'", prev, next)
		}
	}
	trans.Status = next
	return s.repo.UpdateTransaction(ctx, trans)
}

// ListDocuments retrieves all documents for a transaction.
func (s *TradeService) ListDocuments(ctx context.Context, transactionID uint) ([]modelsTrade.TradeDocument, error) {
	return s.repo.ListDocumentsByTransactionID(ctx, transactionID)
}

// DeleteDocument removes a trade document by ID.
func (s *TradeService) DeleteDocument(ctx context.Context, id uint) error {
	return s.repo.DeleteDocument(ctx, id)
}

// ListCompliance returns compliance requirements for a transaction.
func (s *TradeService) ListCompliance(ctx context.Context, transactionID uint) ([]modelsTrade.ComplianceRequirement, error) {
	return s.repo.GetComplianceByTransactionID(ctx, transactionID)
}

// UpdateComplianceStatus updates the status of a compliance requirement by ID.
func (s *TradeService) UpdateComplianceStatus(ctx context.Context, id uint, status string) error {
	return s.repo.UpdateComplianceStatus(ctx, id, status)
}
