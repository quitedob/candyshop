package tool

import (
	"context"
	"fmt"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GenerateCIRequest holds args for generating a Commercial Invoice
type GenerateCIRequest struct {
	PIReference     string  `json:"pi_reference" jsonschema_description:"Reference number of the accepted Proforma Invoice"`
	FinalQuantity   int     `json:"final_quantity" jsonschema_description:"Final shipped quantity"`
	FinalAmount     float64 `json:"final_amount" jsonschema_description:"Final invoice amount"`
	ShippingVessel  string  `json:"shipping_vessel" jsonschema_description:"Name of the vessel or flight number"`
	PortOfLoading   string  `json:"port_of_loading" jsonschema_description:"Port where goods were loaded"`
	PortOfDischarge string  `json:"port_of_discharge" jsonschema_description:"Port where goods will be discharged"`
}

type GenerateCIResponse struct {
	Status  string `json:"status"`
	Content string `json:"content"`
}

// NewGenerateCITool creates a commercial invoice generation tool wrapper
func NewGenerateCITool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_commercial_invoice", "Generate a Commercial Invoice from a confirmed Proforma Invoice and shipping details.",
		func(ctx context.Context, req *GenerateCIRequest) (*GenerateCIResponse, error) {

			contentStr := fmt.Sprintf(`{
				"pi_ref": "%s",
				"quantity": %d,
				"amount": %.2f,
				"vessel": "%s",
				"pol": "%s",
				"pod": "%s"
			}`, req.PIReference, req.FinalQuantity, req.FinalAmount, req.ShippingVessel, req.PortOfLoading, req.PortOfDischarge)

			return &GenerateCIResponse{
				Status:  "draft_ci_created",
				Content: contentStr,
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
