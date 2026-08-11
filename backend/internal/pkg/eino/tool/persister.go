package tool

import (
	"context"

	tradeModels "candypro/api/internal/models/trade"
)

// DocumentPersister abstracts trade document persistence so tools can save
// AI-generated documents without knowing about the underlying storage.
// TradeService satisfies this interface.
type DocumentPersister interface {
	AddDocument(ctx context.Context, doc *tradeModels.TradeDocument) error

	// ResolveProformaInvoiceRef returns the doc_number of the trade's existing
	// Proforma Invoice, or "" when none exists. Errors also yield "" so callers
	// never block document generation on a reference lookup.
	ResolveProformaInvoiceRef(ctx context.Context, transactionID uint) (string, error)
}
