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

type GenerateInsuranceRequest struct {
	InsuredParty      string  `json:"insured_party" jsonschema_description:"Name of the company receiving the insurance"`
	CoveragePercent   float64 `json:"coverage_percent" jsonschema_description:"Percentage of invoice value covered (e.g., 110)"`
	TotalInvoiceValue float64 `json:"total_invoice_value" jsonschema_description:"Total CIF/CIP invoice value in USD"`
	VesselDetails     string  `json:"vessel_details" jsonschema_description:"Vessel name or Flight number"`
	PortOfLoading     string  `json:"port_of_loading" jsonschema_description:"Port of departure"`
	PortOfDischarge   string  `json:"port_of_discharge" jsonschema_description:"Port of destination"`
	TradeID           uint    `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GenerateInsuranceResponse struct {
	Status     string  `json:"status"`
	RequestNo  string  `json:"request_no"`
	Content    string  `json:"content"`
	InsuredAmt float64 `json:"insured_amount"`
	SavedTo    string  `json:"saved_to,omitempty"`
}

func NewGenerateInsuranceCertTool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_insurance_certificate_request",
		"Generate a request draft for a Marine/Cargo Insurance Certificate under CIF/CIP terms.",
		func(ctx context.Context, req *GenerateInsuranceRequest) (*GenerateInsuranceResponse, error) {
			reqNo := fmt.Sprintf("INS-REQ-%d", time.Now().Unix())
			insuredAmt := req.TotalInvoiceValue * (req.CoveragePercent / 100.0)

			contentStr := fmt.Sprintf(`{
	"insured": %q,
	"vessel": %q,
	"route": "%s to %s",
	"invoice_value": %.2f,
	"coverage": "%.1f%%",
	"insured_amount": %.2f
}`, req.InsuredParty, req.VesselDetails, req.PortOfLoading, req.PortOfDischarge,
				req.TotalInvoiceValue, req.CoveragePercent, insuredAmt)

			resp := &GenerateInsuranceResponse{
				Status:     "insurance_request_drafted",
				RequestNo:  reqNo,
				Content:    contentStr,
				InsuredAmt: insuredAmt,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          "INSURANCE_CERTIFICATE",
					DocNumber:     reqNo,
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
