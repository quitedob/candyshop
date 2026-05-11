package tool

import (
	"context"

	tradeModels "candypro/api/internal/models/trade"
)

// DocumentPersister abstracts trade document persistence so tools can save
// AI-generated documents without knowing about the underlying storage.
// TradeService.AddDocument satisfies this interface.
type DocumentPersister interface {
	AddDocument(ctx context.Context, doc *tradeModels.TradeDocument) error
}
