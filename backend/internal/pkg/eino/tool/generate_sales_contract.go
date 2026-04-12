package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GenerateContractRequest holds args for generating a Sales Contract
type GenerateContractRequest struct {
	PIReference     string  `json:"pi_reference" jsonschema_description:"Reference number of the accepted Proforma Invoice"`
	BuyerName       string  `json:"buyer_name" jsonschema_description:"Corporate name of the buying party"`
	Incoterms       string  `json:"incoterms" jsonschema_description:"Shipping terms, e.g., FOB, CIF, EXW"`
	TermsOfPayment  string  `json:"terms_of_payment" jsonschema_description:"Payment structure, e.g., 30% T/T Advance, 70% L/C at sight"`
	QualityStandard string  `json:"quality_standard" jsonschema_description:"Agreed upon quality and compliance standards for the product"`
	TotalAmount     float64 `json:"total_amount" jsonschema_description:"Final negotiated order value"`
	Currency        string  `json:"currency" jsonschema_description:"Currency of the transaction, e.g., USD, EUR"`
	ValidityDays    int     `json:"validity_days" jsonschema_description:"Number of days the contract remains valid for signature"`
}

type GenerateContractResponse struct {
	Status     string `json:"status"`
	ContractNo string `json:"contract_no"`
	Content    string `json:"content"`
	Timestamp  string `json:"timestamp"`
}

// NewGenerateSalesContractTool creates a Sales Contract generation tool wrapper
func NewGenerateSalesContractTool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_sales_contract", "Generate a formal Sales Contract from a confirmed Proforma Invoice with detailed terms and conditions.",
		func(ctx context.Context, req *GenerateContractRequest) (*GenerateContractResponse, error) {

			contractNo := fmt.Sprintf("SC-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
				"reference": "%s",
				"buyer": "%s",
				"incoterms": "%s",
				"payment_terms": "%s",
				"quality": "%s",
				"amount": %.2f,
				"currency": "%s"
			}`, req.PIReference, req.BuyerName, req.Incoterms, req.TermsOfPayment, req.QualityStandard, req.TotalAmount, req.Currency)

			return &GenerateContractResponse{
				Status:     "draft_contract_created",
				ContractNo: contractNo,
				Content:    contentStr,
				Timestamp:  time.Now().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
