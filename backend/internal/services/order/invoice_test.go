package order

import (
	"context"
	"testing"

	modelsOrder "candypro/api/internal/models/order"
	orderRepo "candypro/api/internal/repository/order"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// invoiceOrderReaderStub returns a canned order for the derived-invoice path.
type invoiceOrderReaderStub struct {
	order *modelsOrder.Order
}

func (s *invoiceOrderReaderStub) FindByID(ctx context.Context, id string) (*modelsOrder.Order, error) {
	if s.order == nil {
		return nil, nil
	}
	return s.order, nil
}

func setupInvoiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsOrder.Invoice{}); err != nil {
		t.Fatalf("migrate invoices: %v", err)
	}
	return db
}

func newInvoiceTestService(db *gorm.DB, ord *modelsOrder.Order) *InvoiceService {
	return NewInvoiceService(orderRepo.NewInvoiceRepository(db), &invoiceOrderReaderStub{order: ord}, nil)
}

// TestCreateInvoiceFromOrder_TotalIncludesShipping is the H9 regression: a
// derived invoice must keep the freight inside TotalAmount instead of having
// CreateInvoice overwrite it with a shipping-free amount+tax recomputation.
func TestCreateInvoiceFromOrder_TotalIncludesShipping(t *testing.T) {
	db := setupInvoiceTestDB(t)
	ord := &modelsOrder.Order{
		ID:             "order-h9",
		Currency:       "USD",
		Subtotal:       1000,
		TaxAmount:      50,
		ShippingAmount: 120,
		Items:          modelsOrder.OrderItemArray{},
	}
	svc := newInvoiceTestService(db, ord)

	inv, err := svc.CreateInvoiceFromOrder(context.Background(), ord.ID, "")
	if err != nil {
		t.Fatalf("CreateInvoiceFromOrder: %v", err)
	}
	want := ord.Subtotal + ord.TaxAmount + ord.ShippingAmount // 1170
	if inv.TotalAmount != want {
		t.Fatalf("derived invoice TotalAmount = %v, want %v (subtotal+tax+shipping)", inv.TotalAmount, want)
	}
	got, err := svc.GetInvoice(context.Background(), inv.ID)
	if err != nil {
		t.Fatalf("GetInvoice: %v", err)
	}
	if got.TotalAmount != want {
		t.Fatalf("persisted derived invoice TotalAmount = %v, want %v", got.TotalAmount, want)
	}
}

// TestCreateInvoice_FallbackTotalWhenUnset ensures the amount+tax fallback is
// still applied when the caller does not supply a total (manual create path).
func TestCreateInvoice_FallbackTotalWhenUnset(t *testing.T) {
	db := setupInvoiceTestDB(t)
	svc := newInvoiceTestService(db, nil)

	inv := &modelsOrder.Invoice{Amount: 100, TaxAmount: 10}
	if err := svc.CreateInvoice(context.Background(), inv); err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	if inv.TotalAmount != 110 {
		t.Fatalf("manual invoice TotalAmount = %v, want 110 (amount+tax fallback)", inv.TotalAmount)
	}
}

// TestCreateInvoice_PreservesExplicitTotal ensures a caller-supplied total is
// not clobbered by CreateInvoice (manual invoices that already carry freight).
func TestCreateInvoice_PreservesExplicitTotal(t *testing.T) {
	db := setupInvoiceTestDB(t)
	svc := newInvoiceTestService(db, nil)

	inv := &modelsOrder.Invoice{Amount: 100, TaxAmount: 10, TotalAmount: 130}
	if err := svc.CreateInvoice(context.Background(), inv); err != nil {
		t.Fatalf("CreateInvoice: %v", err)
	}
	if inv.TotalAmount != 130 {
		t.Fatalf("manual invoice TotalAmount = %v, want 130 (caller-supplied total preserved)", inv.TotalAmount)
	}
}

// TestValidateAndPersistInvoiceUpdate_PreservesFreight is the H9 companion for
// the manual-edit path: editing tax on a derived invoice must keep the freight
// component carried in the stored total instead of dropping it.
func TestValidateAndPersistInvoiceUpdate_PreservesFreight(t *testing.T) {
	db := setupInvoiceTestDB(t)
	ord := &modelsOrder.Order{
		ID:             "order-h9b",
		Currency:       "USD",
		Subtotal:       1000,
		TaxAmount:      50,
		ShippingAmount: 120,
		Items:          modelsOrder.OrderItemArray{},
	}
	svc := newInvoiceTestService(db, ord)

	stored, err := svc.CreateInvoiceFromOrder(context.Background(), ord.ID, "")
	if err != nil {
		t.Fatalf("CreateInvoiceFromOrder: %v", err)
	}
	before := *stored
	after := *stored
	after.TaxAmount = 70 // edit tax, keep freight

	if err := svc.ValidateAndPersistInvoiceUpdate(context.Background(), &before, &after, "", ""); err != nil {
		t.Fatalf("ValidateAndPersistInvoiceUpdate: %v", err)
	}
	want := after.Amount + after.TaxAmount + ord.ShippingAmount // 1000 + 70 + 120 = 1190
	if after.TotalAmount != want {
		t.Fatalf("updated TotalAmount = %v, want %v (amount+newTax+freight)", after.TotalAmount, want)
	}
	got, err := svc.GetInvoice(context.Background(), stored.ID)
	if err != nil {
		t.Fatalf("GetInvoice: %v", err)
	}
	if got.TotalAmount != want {
		t.Fatalf("persisted updated TotalAmount = %v, want %v", got.TotalAmount, want)
	}
}
