package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"time"

	"gorm.io/gorm"
)

// TradeDocumentDetailRepository handles CRUD for the rich trade document models.
type TradeDocumentDetailRepository struct {
	db *gorm.DB
}

// NewTradeDocumentDetailRepository creates a new TradeDocumentDetailRepository.
func NewTradeDocumentDetailRepository(db *gorm.DB) *TradeDocumentDetailRepository {
	return &TradeDocumentDetailRepository{db: db}
}

// --- SalesContract ---

func (r *TradeDocumentDetailRepository) GetSalesContract(ctx context.Context, transactionID uint) (*modelsTrade.SalesContract, error) {
	var sc modelsTrade.SalesContract
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&sc).Error; err != nil {
		return nil, err
	}
	return &sc, nil
}

func (r *TradeDocumentDetailRepository) CreateSalesContract(ctx context.Context, sc *modelsTrade.SalesContract) error {
	return r.db.WithContext(ctx).Create(sc).Error
}

func (r *TradeDocumentDetailRepository) UpdateSalesContract(ctx context.Context, sc *modelsTrade.SalesContract) error {
	return r.db.WithContext(ctx).Save(sc).Error
}

// --- PackingList ---

func (r *TradeDocumentDetailRepository) GetPackingList(ctx context.Context, transactionID uint) (*modelsTrade.PackingList, error) {
	var pl modelsTrade.PackingList
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&pl).Error; err != nil {
		return nil, err
	}
	return &pl, nil
}

func (r *TradeDocumentDetailRepository) CreatePackingList(ctx context.Context, pl *modelsTrade.PackingList) error {
	return r.db.WithContext(ctx).Create(pl).Error
}

func (r *TradeDocumentDetailRepository) UpdatePackingList(ctx context.Context, pl *modelsTrade.PackingList) error {
	return r.db.WithContext(ctx).Save(pl).Error
}

// --- CertificateOfOrigin ---

func (r *TradeDocumentDetailRepository) GetCertificateOfOrigin(ctx context.Context, transactionID uint) (*modelsTrade.CertificateOfOrigin, error) {
	var coo modelsTrade.CertificateOfOrigin
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&coo).Error; err != nil {
		return nil, err
	}
	return &coo, nil
}

func (r *TradeDocumentDetailRepository) CreateCertificateOfOrigin(ctx context.Context, coo *modelsTrade.CertificateOfOrigin) error {
	return r.db.WithContext(ctx).Create(coo).Error
}

func (r *TradeDocumentDetailRepository) UpdateCertificateOfOrigin(ctx context.Context, coo *modelsTrade.CertificateOfOrigin) error {
	return r.db.WithContext(ctx).Save(coo).Error
}

// --- HealthCertificate ---

func (r *TradeDocumentDetailRepository) GetHealthCertificate(ctx context.Context, transactionID uint) (*modelsTrade.HealthCertificate, error) {
	var hc modelsTrade.HealthCertificate
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&hc).Error; err != nil {
		return nil, err
	}
	return &hc, nil
}

func (r *TradeDocumentDetailRepository) CreateHealthCertificate(ctx context.Context, hc *modelsTrade.HealthCertificate) error {
	return r.db.WithContext(ctx).Create(hc).Error
}

func (r *TradeDocumentDetailRepository) UpdateHealthCertificate(ctx context.Context, hc *modelsTrade.HealthCertificate) error {
	return r.db.WithContext(ctx).Save(hc).Error
}

// --- SettlementRecord ---

func (r *TradeDocumentDetailRepository) ListSettlements(ctx context.Context, transactionID uint) ([]modelsTrade.SettlementRecord, error) {
	var records []modelsTrade.SettlementRecord
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).Order("due_date asc").Find(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *TradeDocumentDetailRepository) GetSettlement(ctx context.Context, id uint) (*modelsTrade.SettlementRecord, error) {
	var record modelsTrade.SettlementRecord
	if err := r.db.WithContext(ctx).First(&record, id).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

func (r *TradeDocumentDetailRepository) CreateSettlement(ctx context.Context, record *modelsTrade.SettlementRecord) error {
	return r.db.WithContext(ctx).Create(record).Error
}

func (r *TradeDocumentDetailRepository) UpdateSettlement(ctx context.Context, record *modelsTrade.SettlementRecord) error {
	return r.db.WithContext(ctx).Save(record).Error
}

// UpdateSettlementStatus transitions a settlement status with a guarded conditional
// UPDATE keyed on the currently-loaded status (M3). extra holds the non-status
// fields (amount_paid, payment_date, lc_reference) edited alongside the status.
func (r *TradeDocumentDetailRepository) UpdateSettlementStatus(ctx context.Context, id uint, fromStatus, toStatus string, extra map[string]interface{}) error {
	updates := map[string]interface{}{
		"status":     toStatus,
		"updated_at": time.Now(),
	}
	for k, v := range extra {
		updates[k] = v
	}
	res := r.db.WithContext(ctx).Model(&modelsTrade.SettlementRecord{}).
		Where("id = ? AND status = ?", id, fromStatus).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrSettlementStateMismatch
	}
	return nil
}

func (r *TradeDocumentDetailRepository) DeleteSettlement(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsTrade.SettlementRecord{}, id).Error
}

// --- ProformaInvoice ---

func (r *TradeDocumentDetailRepository) GetProformaInvoice(ctx context.Context, transactionID uint) (*modelsTrade.ProformaInvoice, error) {
	var pi modelsTrade.ProformaInvoice
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&pi).Error; err != nil {
		return nil, err
	}
	return &pi, nil
}

func (r *TradeDocumentDetailRepository) CreateProformaInvoice(ctx context.Context, pi *modelsTrade.ProformaInvoice) error {
	return r.db.WithContext(ctx).Create(pi).Error
}

func (r *TradeDocumentDetailRepository) UpdateProformaInvoice(ctx context.Context, pi *modelsTrade.ProformaInvoice) error {
	return r.db.WithContext(ctx).Save(pi).Error
}

// --- CommercialInvoice ---

func (r *TradeDocumentDetailRepository) GetCommercialInvoice(ctx context.Context, transactionID uint) (*modelsTrade.CommercialInvoice, error) {
	var ci modelsTrade.CommercialInvoice
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&ci).Error; err != nil {
		return nil, err
	}
	return &ci, nil
}

func (r *TradeDocumentDetailRepository) CreateCommercialInvoice(ctx context.Context, ci *modelsTrade.CommercialInvoice) error {
	return r.db.WithContext(ctx).Create(ci).Error
}

func (r *TradeDocumentDetailRepository) UpdateCommercialInvoice(ctx context.Context, ci *modelsTrade.CommercialInvoice) error {
	return r.db.WithContext(ctx).Save(ci).Error
}

// --- BillOfLading ---

func (r *TradeDocumentDetailRepository) GetBillOfLading(ctx context.Context, transactionID uint) (*modelsTrade.BillOfLading, error) {
	var bl modelsTrade.BillOfLading
	if err := r.db.WithContext(ctx).Where("transaction_id = ?", transactionID).First(&bl).Error; err != nil {
		return nil, err
	}
	return &bl, nil
}

func (r *TradeDocumentDetailRepository) CreateBillOfLading(ctx context.Context, bl *modelsTrade.BillOfLading) error {
	return r.db.WithContext(ctx).Create(bl).Error
}

func (r *TradeDocumentDetailRepository) UpdateBillOfLading(ctx context.Context, bl *modelsTrade.BillOfLading) error {
	return r.db.WithContext(ctx).Save(bl).Error
}
