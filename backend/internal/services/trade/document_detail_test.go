package trade

import (
	"context"
	"testing"

	modelsTrade "candypro/api/internal/models/trade"
)

// fakeDocDetailRepo implements tradeDocDetailRepository for the doc-detail
// service tests. Only the PI/CI create+get paths used by CreateCommercialInvoice
// are exercised; the rest are no-ops.
type fakeDocDetailRepo struct {
	pi        *modelsTrade.ProformaInvoice
	createdCI *modelsTrade.CommercialInvoice
}

func (f *fakeDocDetailRepo) GetSalesContract(context.Context, uint) (*modelsTrade.SalesContract, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) CreateSalesContract(context.Context, *modelsTrade.SalesContract) error {
	return nil
}
func (f *fakeDocDetailRepo) UpdateSalesContract(context.Context, *modelsTrade.SalesContract) error {
	return nil
}
func (f *fakeDocDetailRepo) GetPackingList(context.Context, uint) (*modelsTrade.PackingList, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) CreatePackingList(context.Context, *modelsTrade.PackingList) error {
	return nil
}
func (f *fakeDocDetailRepo) UpdatePackingList(context.Context, *modelsTrade.PackingList) error {
	return nil
}
func (f *fakeDocDetailRepo) GetCertificateOfOrigin(context.Context, uint) (*modelsTrade.CertificateOfOrigin, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) CreateCertificateOfOrigin(context.Context, *modelsTrade.CertificateOfOrigin) error {
	return nil
}
func (f *fakeDocDetailRepo) UpdateCertificateOfOrigin(context.Context, *modelsTrade.CertificateOfOrigin) error {
	return nil
}
func (f *fakeDocDetailRepo) GetHealthCertificate(context.Context, uint) (*modelsTrade.HealthCertificate, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) CreateHealthCertificate(context.Context, *modelsTrade.HealthCertificate) error {
	return nil
}
func (f *fakeDocDetailRepo) UpdateHealthCertificate(context.Context, *modelsTrade.HealthCertificate) error {
	return nil
}
func (f *fakeDocDetailRepo) ListSettlements(context.Context, uint) ([]modelsTrade.SettlementRecord, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) GetSettlement(context.Context, uint) (*modelsTrade.SettlementRecord, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) CreateSettlement(context.Context, *modelsTrade.SettlementRecord) error {
	return nil
}
func (f *fakeDocDetailRepo) UpdateSettlement(context.Context, *modelsTrade.SettlementRecord) error {
	return nil
}
func (f *fakeDocDetailRepo) DeleteSettlement(context.Context, uint) error {
	return nil
}
func (f *fakeDocDetailRepo) GetProformaInvoice(context.Context, uint) (*modelsTrade.ProformaInvoice, error) {
	return f.pi, nil
}
func (f *fakeDocDetailRepo) CreateProformaInvoice(context.Context, *modelsTrade.ProformaInvoice) error {
	return nil
}
func (f *fakeDocDetailRepo) UpdateProformaInvoice(context.Context, *modelsTrade.ProformaInvoice) error {
	return nil
}
func (f *fakeDocDetailRepo) GetCommercialInvoice(context.Context, uint) (*modelsTrade.CommercialInvoice, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) CreateCommercialInvoice(_ context.Context, ci *modelsTrade.CommercialInvoice) error {
	f.createdCI = ci
	return nil
}
func (f *fakeDocDetailRepo) UpdateCommercialInvoice(context.Context, *modelsTrade.CommercialInvoice) error {
	return nil
}
func (f *fakeDocDetailRepo) GetBillOfLading(context.Context, uint) (*modelsTrade.BillOfLading, error) {
	return nil, nil
}
func (f *fakeDocDetailRepo) CreateBillOfLading(context.Context, *modelsTrade.BillOfLading) error {
	return nil
}
func (f *fakeDocDetailRepo) UpdateBillOfLading(context.Context, *modelsTrade.BillOfLading) error {
	return nil
}

func TestCreateCommercialInvoice_AutoFillsPINumberFromExistingPI(t *testing.T) {
	repo := &fakeDocDetailRepo{pi: &modelsTrade.ProformaInvoice{PINumber: "PI-2026-0001"}}
	s := NewTradeDocumentDetailService(repo)

	ci := &modelsTrade.CommercialInvoice{TransactionID: 42, CINumber: "CI-2026-0001"}
	if err := s.CreateCommercialInvoice(context.Background(), ci); err != nil {
		t.Fatalf("CreateCommercialInvoice: %v", err)
	}
	if repo.createdCI == nil {
		t.Fatal("repo.CreateCommercialInvoice not called")
	}
	if repo.createdCI.PINumber != "PI-2026-0001" {
		t.Fatalf("expected auto-filled PINumber PI-2026-0001, got %q", repo.createdCI.PINumber)
	}
}

func TestCreateCommercialInvoice_NoPILeavesEmptyRef(t *testing.T) {
	repo := &fakeDocDetailRepo{pi: nil}
	s := NewTradeDocumentDetailService(repo)

	ci := &modelsTrade.CommercialInvoice{TransactionID: 42, CINumber: "CI-2026-0002"}
	if err := s.CreateCommercialInvoice(context.Background(), ci); err != nil {
		t.Fatalf("CreateCommercialInvoice: %v", err)
	}
	if repo.createdCI.PINumber != "" {
		t.Fatalf("expected empty PINumber when no PI exists, got %q", repo.createdCI.PINumber)
	}
}

func TestCreateCommercialInvoice_PreservesProvidedPINumber(t *testing.T) {
	repo := &fakeDocDetailRepo{pi: &modelsTrade.ProformaInvoice{PINumber: "PI-2026-0001"}}
	s := NewTradeDocumentDetailService(repo)

	ci := &modelsTrade.CommercialInvoice{TransactionID: 42, CINumber: "CI-2026-0003", PINumber: "PI-CLIENT-SUPPLIED"}
	if err := s.CreateCommercialInvoice(context.Background(), ci); err != nil {
		t.Fatalf("CreateCommercialInvoice: %v", err)
	}
	if repo.createdCI.PINumber != "PI-CLIENT-SUPPLIED" {
		t.Fatalf("expected provided PINumber preserved, got %q", repo.createdCI.PINumber)
	}
}
