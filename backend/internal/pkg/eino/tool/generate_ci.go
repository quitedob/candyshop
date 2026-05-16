package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	tradeModels "candypro/api/internal/models/trade"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateCIRequest struct {
	PIReference     string `json:"pi_reference" jsonschema_description:"Reference number of the accepted Proforma Invoice"`
	FinalQuantity   int    `json:"final_quantity" jsonschema_description:"Final shipped quantity"`
	FinalAmount     float64 `json:"final_amount" jsonschema_description:"Final invoice amount"`
	ShippingVessel  string `json:"shipping_vessel" jsonschema_description:"Name of the vessel or flight number"`
	PortOfLoading   string `json:"port_of_loading" jsonschema_description:"Port where goods were loaded"`
	PortOfDischarge string `json:"port_of_discharge" jsonschema_description:"Port where goods will be discharged"`
	TradeID         uint   `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GenerateCIResponse struct {
	Status  string `json:"status"`
	DocNo   string `json:"doc_number"`
	Content string `json:"content"`
	SavedTo string `json:"saved_to,omitempty"`
}

func NewGenerateCITool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_commercial_invoice",
		"Generate a Commercial Invoice from a confirmed Proforma Invoice and shipping details.",
		func(ctx context.Context, req *GenerateCIRequest) (*GenerateCIResponse, error) {
			docNo := fmt.Sprintf("CI-%d", ctx.Value("timestamp"))

			contentStr := fmt.Sprintf(`{
	"pi_ref": %q,
	"quantity": %d,
	"amount": %.2f,
	"vessel": %q,
	"pol": %q,
	"pod": %q
}`, req.PIReference, req.FinalQuantity, req.FinalAmount, req.ShippingVessel, req.PortOfLoading, req.PortOfDischarge)

			resp := &GenerateCIResponse{
				Status:  "draft_ci_created",
				DocNo:   docNo,
				Content: contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          tradeModels.DocTypeCommercialInvoice,
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
