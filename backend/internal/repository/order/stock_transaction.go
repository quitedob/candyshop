package order

import (
	modelsOrder "candypro/api/internal/models/order"
	"context"
	"time"

	"gorm.io/gorm"
)

// StockTransactionRepository records inventory audit trail entries.
type StockTransactionRepository struct {
	db *gorm.DB
}

// NewStockTransactionRepository creates a new StockTransactionRepository.
func NewStockTransactionRepository(db *gorm.DB) *StockTransactionRepository {
	return &StockTransactionRepository{db: db}
}

// Record writes a single stock transaction audit record.
func (r *StockTransactionRepository) Record(ctx context.Context, tx *modelsOrder.StockTransaction) error {
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}
	return r.db.WithContext(ctx).Create(tx).Error
}

// RecordMany writes multiple stock transaction audit records in a batch.
func (r *StockTransactionRepository) RecordMany(ctx context.Context, txs []*modelsOrder.StockTransaction) error {
	if len(txs) == 0 {
		return nil
	}
	now := time.Now()
	for _, tx := range txs {
		if tx.CreatedAt.IsZero() {
			tx.CreatedAt = now
		}
	}
	return r.db.WithContext(ctx).Create(txs).Error
}

// FindByProductID returns paginated stock transactions for a product.
func (r *StockTransactionRepository) FindByProductID(ctx context.Context, productID string, page, limit int) ([]modelsOrder.StockTransaction, int64, error) {
	var txs []modelsOrder.StockTransaction
	var total int64
	query := r.db.WithContext(ctx).Model(&modelsOrder.StockTransaction{}).Where("product_id = ?", productID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&txs).Error; err != nil {
		return nil, 0, err
	}
	return txs, total, nil
}

// FindByReferenceID returns all stock transactions for a given reference (e.g. order ID).
func (r *StockTransactionRepository) FindByReferenceID(ctx context.Context, referenceID string) ([]modelsOrder.StockTransaction, error) {
	var txs []modelsOrder.StockTransaction
	if err := r.db.WithContext(ctx).Where("reference_id = ?", referenceID).Order("created_at DESC").Find(&txs).Error; err != nil {
		return nil, err
	}
	return txs, nil
}
