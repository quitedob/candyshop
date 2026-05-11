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

type GenerateSLIRequest struct {
	ShipperName     string `json:"shipper_name" jsonschema_description:"Name of the exporting company"`
	ConsigneeName   string `json:"consignee_name" jsonschema_description:"Name of the importing company"`
	PortOfLoading   string `json:"port_of_loading" jsonschema_description:"Port of departure"`
	PortOfDischarge string `json:"port_of_discharge" jsonschema_description:"Port of destination"`
	CargoDetails    string `json:"cargo_details" jsonschema_description:"Brief description of cargo, volume, and weight"`
	IsFCL           bool   `json:"is_fcl" jsonschema_description:"True if Full Container Load, False if LCL"`
	TradeID         uint   `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GenerateSLIResponse struct {
	Status  string `json:"status"`
	DocNo   string `json:"doc_no"`
	Content string `json:"content"`
	SavedTo string `json:"saved_to,omitempty"`
}

func NewGenerateSLITool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_shipper_letter_of_instruction",
		"Generate a Shipper's Letter of Instruction (SLI) for the freight forwarder.",
		func(ctx context.Context, req *GenerateSLIRequest) (*GenerateSLIResponse, error) {
			docNo := fmt.Sprintf("SLI-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
	"shipper": %q,
	"consignee": %q,
	"origin_port": %q,
	"destination_port": %q,
	"cargo": %q,
	"fcl_status": %t
}`, req.ShipperName, req.ConsigneeName, req.PortOfLoading, req.PortOfDischarge, req.CargoDetails, req.IsFCL)

			resp := &GenerateSLIResponse{
				Status:  "sli_drafted",
				DocNo:   docNo,
				Content: contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          "SHIPPER_LETTER_OF_INSTRUCTION",
					DocNumber:     docNo,
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
