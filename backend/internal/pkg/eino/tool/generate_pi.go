package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	tradeModels "candypro/api/internal/models/trade"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GeneratePIRequest holds args for generating a Proforma Invoice
type GeneratePIRequest struct {
	BuyerName       string  `json:"buyer_name" jsonschema_description:"Name of the buying company"`
	SellerName      string  `json:"seller_name" jsonschema_description:"Name of the selling company (CandyPro)"`
	ItemsJSON       string  `json:"items_json" jsonschema_description:"JSON string of items, quantities, and unit prices"`
	Incoterms       string  `json:"incoterms" jsonschema_description:"Incoterms, e.g., FOB Shanghai, CIF New York"`
	PaymentTerms    string  `json:"payment_terms" jsonschema_description:"Payment terms, e.g., 30% TT advance, 70% LC"`
	EstimatedAmount float64 `json:"estimated_amount" jsonschema_description:"Estimated total amount"`
	TradeID         uint    `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

// GeneratePIResponse is the result of the PI generation
type GeneratePIResponse struct {
	Status   string `json:"status"`
	DocNo    string `json:"doc_number"`
	Content  string `json:"content"`
	SavedTo  string `json:"saved_to,omitempty"`
}

// NewGeneratePITool creates an Eino tool that generates and persists a Proforma Invoice.
func NewGeneratePITool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_proforma_invoice",
		"Generate a Proforma Invoice based on buyer/seller details and items. Requires user review.",
		func(ctx context.Context, req *GeneratePIRequest) (*GeneratePIResponse, error) {
			docNo := fmt.Sprintf("PI-%d", time.Now().Unix())

			contentStr := fmt.Sprintf(`{
	"buyer": %q,
	"seller": %q,
	"items": %s,
	"incoterms": %q,
	"payment_terms": %q,
	"total_amount": %.2f,
	"valid_until": %q
}`, req.BuyerName, req.SellerName, req.ItemsJSON, req.Incoterms, req.PaymentTerms, req.EstimatedAmount,
				time.Now().AddDate(0, 1, 0).Format("2006-01-02"))

			resp := &GeneratePIResponse{
				Status:  "draft_created",
				DocNo:   docNo,
				Content: contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          tradeModels.DocTypeProformaInvoice,
					DocNumber:     docNo,
					Status:        tradeModels.TradeStatusDraft,
					IsAIGenerated: true,
					LineageSource: "ai_draft",
				}
				contentJSON, _ := json.Marshal(contentStr)
				if ue := json.Unmarshal(contentJSON, &doc.Content); ue != nil {
				log.Printf("eino tool: unmarshal doc content failed: %v", ue)
			}
				if err := persister.AddDocument(ctx, doc); err != nil {
					return resp, fmt.Errorf("saved to DB but persist failed: %w", err)
				}
				resp.SavedTo = fmt.Sprintf("DB(trade=%d)", req.TradeID)
			}

			return resp, nil
		})
	if err != nil {
		return nil, err
	}

	return baseTool, nil
}
