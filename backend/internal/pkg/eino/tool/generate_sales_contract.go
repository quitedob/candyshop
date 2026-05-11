package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tradeModels "candypro/api/internal/models/trade"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateContractRequest struct {
	PIReference     string  `json:"pi_reference" jsonschema_description:"Reference number of the accepted Proforma Invoice"`
	BuyerName       string  `json:"buyer_name" jsonschema_description:"Corporate name of the buying party"`
	Incoterms       string  `json:"incoterms" jsonschema_description:"Shipping terms, e.g., FOB, CIF, EXW"`
	TermsOfPayment  string  `json:"terms_of_payment" jsonschema_description:"Payment structure, e.g., 30% T/T Advance, 70% L/C at sight"`
	QualityStandard string  `json:"quality_standard" jsonschema_description:"Agreed quality and compliance standards for the product"`
	TotalAmount     float64 `json:"total_amount" jsonschema_description:"Final negotiated order value"`
	Currency        string  `json:"currency" jsonschema_description:"Currency, e.g., USD, EUR"`
	ValidityDays    int     `json:"validity_days" jsonschema_description:"Days the contract remains valid for signature"`
	TradeID         uint    `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GenerateContractResponse struct {
	Status     string `json:"status"`
	ContractNo string `json:"contract_no"`
	Content    string `json:"content"`
	SavedTo    string `json:"saved_to,omitempty"`
}

func NewGenerateSalesContractTool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_sales_contract",
		"Generate a formal Sales Contract from a confirmed Proforma Invoice with detailed terms and conditions.",
		func(ctx context.Context, req *GenerateContractRequest) (*GenerateContractResponse, error) {
			contractNo := fmt.Sprintf("SC-%d", time.Now().Unix())

			contentStr := fmt.Sprintf(`{
	"reference": %q,
	"buyer": %q,
	"incoterms": %q,
	"payment_terms": %q,
	"quality": %q,
	"amount": %.2f,
	"currency": %q
}`, req.PIReference, req.BuyerName, req.Incoterms, req.TermsOfPayment, req.QualityStandard, req.TotalAmount, req.Currency)

			resp := &GenerateContractResponse{
				Status:     "draft_contract_created",
				ContractNo: contractNo,
				Content:    contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          tradeModels.DocTypeSalesContract,
					DocNumber:     contractNo,
					Status:        tradeModels.TradeStatusDraft,
					IsAIGenerated: true,
					LineageSource: "ai_draft",
				}
				contentJSON, _ := json.Marshal(contentStr)
				_ = json.Unmarshal(contentJSON, &doc.Content)
				if err := persister.AddDocument(ctx, doc); err != nil {
					return resp, fmt.Errorf("persist failed: %w", err)
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
