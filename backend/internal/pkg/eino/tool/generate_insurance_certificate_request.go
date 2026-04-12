package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateInsuranceRequest struct {
	InsuredParty      string  `json:"insured_party" jsonschema_description:"Name of the company receiving the insurance (Conignee or Shipper)"`
	CoveragePercent   float64 `json:"coverage_percent" jsonschema_description:"Percentage of invoice value covered (e.g., 110%)"`
	TotalInvoiceValue float64 `json:"total_invoice_value" jsonschema_description:"Total CIF/CIP invoice value in USD"`
	VesselDetails     string  `json:"vessel_details" jsonschema_description:"Vessel name or Flight number"`
	PortOfLoading     string  `json:"port_of_loading" jsonschema_description:"Port of departure"`
	PortOfDischarge   string  `json:"port_of_discharge" jsonschema_description:"Port of destination"`
}

type GenerateInsuranceResponse struct {
	Status     string  `json:"status"`
	RequestNo  string  `json:"request_no"`
	Content    string  `json:"content"`
	InsuredAmt float64 `json:"insured_amount"`
	Timestamp  string  `json:"timestamp"`
}

func NewGenerateInsuranceCertTool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_insurance_certificate_request", "Generate a request draft for a Marine/Cargo Insurance Certificate under CIF/CIP terms.",
		func(ctx context.Context, req *GenerateInsuranceRequest) (*GenerateInsuranceResponse, error) {

			reqNo := fmt.Sprintf("INS-REQ-%d", time.Now().Unix())
			insuredAmt := req.TotalInvoiceValue * (req.CoveragePercent / 100.0)

			contentStr := fmt.Sprintf(`{
				"insured": "%s",
				"vessel": "%s",
				"route": "%s to %s",
				"invoice_value": %.2f,
				"coverage": "%.1f%%",
				"insured_amount": %.2f
			}`, req.InsuredParty, req.VesselDetails, req.PortOfLoading, req.PortOfDischarge, req.TotalInvoiceValue, req.CoveragePercent, insuredAmt)

			return &GenerateInsuranceResponse{
				Status:     "insurance_request_drafted",
				RequestNo:  reqNo,
				Content:    contentStr,
				InsuredAmt: insuredAmt,
				Timestamp:  time.Now().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
