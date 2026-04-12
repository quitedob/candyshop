package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
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
}

// GeneratePIResponse is the result of the PI generation
type GeneratePIResponse struct {
	status  string
	content string
}

// NewGeneratePITool creates an Eino tool that generates a Proforma Invoice.
// Wrapped in an InvokableReviewEditTool so the user can review/edit it before saving.
func NewGeneratePITool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_proforma_invoice", "Generate a Proforma Invoice based on buyer/seller details and items. Requires user review.",
		func(ctx context.Context, req *GeneratePIRequest) (*GeneratePIResponse, error) {

			// Build the PI Content JSON
			contentStr := fmt.Sprintf(`{
				"buyer": "%s",
				"seller": "%s",
				"items": %s,
				"incoterms": "%s",
				"payment_terms": "%s",
				"total_amount": %.2f,
				"valid_until": "%s"
			}`, req.BuyerName, req.SellerName, req.ItemsJSON, req.Incoterms, req.PaymentTerms, req.EstimatedAmount, time.Now().AddDate(0, 1, 0).Format("2006-01-02"))

			return &GeneratePIResponse{
				status:  "draft_created",
				content: contentStr,
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
