package trade

import (
	modelsTrade "candypro/api/internal/models/trade"
	"context"
	"fmt"
	"strings"
	"time"
)

type tradeDocDetailRepository interface {
	GetSalesContract(ctx context.Context, transactionID uint) (*modelsTrade.SalesContract, error)
	CreateSalesContract(ctx context.Context, sc *modelsTrade.SalesContract) error
	UpdateSalesContract(ctx context.Context, sc *modelsTrade.SalesContract) error

	GetPackingList(ctx context.Context, transactionID uint) (*modelsTrade.PackingList, error)
	CreatePackingList(ctx context.Context, pl *modelsTrade.PackingList) error
	UpdatePackingList(ctx context.Context, pl *modelsTrade.PackingList) error

	GetCertificateOfOrigin(ctx context.Context, transactionID uint) (*modelsTrade.CertificateOfOrigin, error)
	CreateCertificateOfOrigin(ctx context.Context, coo *modelsTrade.CertificateOfOrigin) error
	UpdateCertificateOfOrigin(ctx context.Context, coo *modelsTrade.CertificateOfOrigin) error

	GetHealthCertificate(ctx context.Context, transactionID uint) (*modelsTrade.HealthCertificate, error)
	CreateHealthCertificate(ctx context.Context, hc *modelsTrade.HealthCertificate) error
	UpdateHealthCertificate(ctx context.Context, hc *modelsTrade.HealthCertificate) error

	ListSettlements(ctx context.Context, transactionID uint) ([]modelsTrade.SettlementRecord, error)
	GetSettlement(ctx context.Context, id uint) (*modelsTrade.SettlementRecord, error)
	CreateSettlement(ctx context.Context, record *modelsTrade.SettlementRecord) error
	UpdateSettlement(ctx context.Context, record *modelsTrade.SettlementRecord) error
	DeleteSettlement(ctx context.Context, id uint) error

	GetProformaInvoice(ctx context.Context, transactionID uint) (*modelsTrade.ProformaInvoice, error)
	CreateProformaInvoice(ctx context.Context, pi *modelsTrade.ProformaInvoice) error
	UpdateProformaInvoice(ctx context.Context, pi *modelsTrade.ProformaInvoice) error

	GetCommercialInvoice(ctx context.Context, transactionID uint) (*modelsTrade.CommercialInvoice, error)
	CreateCommercialInvoice(ctx context.Context, ci *modelsTrade.CommercialInvoice) error
	UpdateCommercialInvoice(ctx context.Context, ci *modelsTrade.CommercialInvoice) error

	GetBillOfLading(ctx context.Context, transactionID uint) (*modelsTrade.BillOfLading, error)
	CreateBillOfLading(ctx context.Context, bl *modelsTrade.BillOfLading) error
	UpdateBillOfLading(ctx context.Context, bl *modelsTrade.BillOfLading) error
}

// TradeDocumentDetailService provides business logic for rich trade document models.
type TradeDocumentDetailService struct {
	repo tradeDocDetailRepository
}

// NewTradeDocumentDetailService creates a TradeDocumentDetailService.
func NewTradeDocumentDetailService(repo tradeDocDetailRepository) *TradeDocumentDetailService {
	return &TradeDocumentDetailService{repo: repo}
}

// --- SalesContract ---

func (s *TradeDocumentDetailService) GetSalesContract(ctx context.Context, transactionID uint) (*modelsTrade.SalesContract, error) {
	return s.repo.GetSalesContract(ctx, transactionID)
}

func (s *TradeDocumentDetailService) CreateSalesContract(ctx context.Context, sc *modelsTrade.SalesContract) error {
	if sc.ContractNo == "" {
		sc.ContractNo = s.GenerateDocNo("SC")
	}
	if sc.Currency == "" {
		sc.Currency = "USD"
	}
	return s.repo.CreateSalesContract(ctx, sc)
}

func (s *TradeDocumentDetailService) UpdateSalesContract(ctx context.Context, sc *modelsTrade.SalesContract) error {
	return s.repo.UpdateSalesContract(ctx, sc)
}

// --- PackingList ---

func (s *TradeDocumentDetailService) GetPackingList(ctx context.Context, transactionID uint) (*modelsTrade.PackingList, error) {
	return s.repo.GetPackingList(ctx, transactionID)
}

func (s *TradeDocumentDetailService) CreatePackingList(ctx context.Context, pl *modelsTrade.PackingList) error {
	if pl.PLNumber == "" {
		pl.PLNumber = s.GenerateDocNo("PL")
	}
	return s.repo.CreatePackingList(ctx, pl)
}

func (s *TradeDocumentDetailService) UpdatePackingList(ctx context.Context, pl *modelsTrade.PackingList) error {
	return s.repo.UpdatePackingList(ctx, pl)
}

// --- CertificateOfOrigin ---

func (s *TradeDocumentDetailService) GetCertificateOfOrigin(ctx context.Context, transactionID uint) (*modelsTrade.CertificateOfOrigin, error) {
	return s.repo.GetCertificateOfOrigin(ctx, transactionID)
}

func (s *TradeDocumentDetailService) CreateCertificateOfOrigin(ctx context.Context, coo *modelsTrade.CertificateOfOrigin) error {
	if coo.CertificateNo == "" {
		coo.CertificateNo = s.GenerateDocNo("COO")
	}
	return s.repo.CreateCertificateOfOrigin(ctx, coo)
}

func (s *TradeDocumentDetailService) UpdateCertificateOfOrigin(ctx context.Context, coo *modelsTrade.CertificateOfOrigin) error {
	return s.repo.UpdateCertificateOfOrigin(ctx, coo)
}

// --- HealthCertificate ---

func (s *TradeDocumentDetailService) GetHealthCertificate(ctx context.Context, transactionID uint) (*modelsTrade.HealthCertificate, error) {
	return s.repo.GetHealthCertificate(ctx, transactionID)
}

func (s *TradeDocumentDetailService) CreateHealthCertificate(ctx context.Context, hc *modelsTrade.HealthCertificate) error {
	if hc.CertificateNo == "" {
		hc.CertificateNo = s.GenerateDocNo("HC")
	}
	return s.repo.CreateHealthCertificate(ctx, hc)
}

func (s *TradeDocumentDetailService) UpdateHealthCertificate(ctx context.Context, hc *modelsTrade.HealthCertificate) error {
	return s.repo.UpdateHealthCertificate(ctx, hc)
}

// --- SettlementRecord ---

func (s *TradeDocumentDetailService) ListSettlements(ctx context.Context, transactionID uint) ([]modelsTrade.SettlementRecord, error) {
	return s.repo.ListSettlements(ctx, transactionID)
}

func (s *TradeDocumentDetailService) GetSettlement(ctx context.Context, id uint) (*modelsTrade.SettlementRecord, error) {
	return s.repo.GetSettlement(ctx, id)
}

func (s *TradeDocumentDetailService) CreateSettlement(ctx context.Context, record *modelsTrade.SettlementRecord) error {
	if record.Status == "" {
		record.Status = "UNPAID"
	}
	if record.Currency == "" {
		record.Currency = "USD"
	}
	return s.repo.CreateSettlement(ctx, record)
}

func (s *TradeDocumentDetailService) UpdateSettlement(ctx context.Context, record *modelsTrade.SettlementRecord) error {
	// M-13: enforce settlement status transitions before persisting. Previously
	// this passed the record straight through, allowing direct PAID→UNPAID flips
	// or arbitrary string values to land in the DB.
	current, err := s.repo.GetSettlement(ctx, record.ID)
	if err == nil && current != nil {
		if err := modelsTrade.ValidateSettlementStatusTransition(current.Status, record.Status); err != nil {
			return err
		}
	}
	return s.repo.UpdateSettlement(ctx, record)
}

func (s *TradeDocumentDetailService) DeleteSettlement(ctx context.Context, id uint) error {
	return s.repo.DeleteSettlement(ctx, id)
}

// GenerateDocNo generates a document number with a given prefix. Exposed so the
// admin CRUD handler produces numbers identically to the service layer.
func (s *TradeDocumentDetailService) GenerateDocNo(prefix string) string {
	return fmt.Sprintf("%s-%s-%06d", prefix, time.Now().UTC().Format("200601"), time.Now().UnixNano()%1000000)
}

// --- ProformaInvoice ---

func (s *TradeDocumentDetailService) GetProformaInvoice(ctx context.Context, transactionID uint) (*modelsTrade.ProformaInvoice, error) {
	return s.repo.GetProformaInvoice(ctx, transactionID)
}

func (s *TradeDocumentDetailService) CreateProformaInvoice(ctx context.Context, pi *modelsTrade.ProformaInvoice) error {
	if pi.PINumber == "" {
		pi.PINumber = s.GenerateDocNo("PI")
	}
	if pi.Currency == "" {
		pi.Currency = "USD"
	}
	if pi.Status == "" {
		pi.Status = "DRAFT"
	}
	return s.repo.CreateProformaInvoice(ctx, pi)
}

func (s *TradeDocumentDetailService) UpdateProformaInvoice(ctx context.Context, pi *modelsTrade.ProformaInvoice) error {
	return s.repo.UpdateProformaInvoice(ctx, pi)
}

// --- CommercialInvoice ---

func (s *TradeDocumentDetailService) GetCommercialInvoice(ctx context.Context, transactionID uint) (*modelsTrade.CommercialInvoice, error) {
	return s.repo.GetCommercialInvoice(ctx, transactionID)
}

func (s *TradeDocumentDetailService) CreateCommercialInvoice(ctx context.Context, ci *modelsTrade.CommercialInvoice) error {
	if ci.CINumber == "" {
		ci.CINumber = s.GenerateDocNo("CI")
	}
	// Auto-populate the CI→PI cross-reference from the trade's existing PI so a
	// CI created without an explicit pi_number never ships with an empty ref.
	if strings.TrimSpace(ci.PINumber) == "" {
		if pi, err := s.repo.GetProformaInvoice(ctx, ci.TransactionID); err == nil && pi != nil {
			ci.PINumber = pi.PINumber
		}
	}
	if ci.Currency == "" {
		ci.Currency = "USD"
	}
	if ci.Status == "" {
		ci.Status = "DRAFT"
	}
	return s.repo.CreateCommercialInvoice(ctx, ci)
}

func (s *TradeDocumentDetailService) UpdateCommercialInvoice(ctx context.Context, ci *modelsTrade.CommercialInvoice) error {
	return s.repo.UpdateCommercialInvoice(ctx, ci)
}

// --- BillOfLading ---

func (s *TradeDocumentDetailService) GetBillOfLading(ctx context.Context, transactionID uint) (*modelsTrade.BillOfLading, error) {
	return s.repo.GetBillOfLading(ctx, transactionID)
}

func (s *TradeDocumentDetailService) CreateBillOfLading(ctx context.Context, bl *modelsTrade.BillOfLading) error {
	if bl.BLNumber == "" {
		bl.BLNumber = s.GenerateDocNo("BL")
	}
	if bl.Status == "" {
		bl.Status = "DRAFT"
	}
	return s.repo.CreateBillOfLading(ctx, bl)
}

func (s *TradeDocumentDetailService) UpdateBillOfLading(ctx context.Context, bl *modelsTrade.BillOfLading) error {
	return s.repo.UpdateBillOfLading(ctx, bl)
}
