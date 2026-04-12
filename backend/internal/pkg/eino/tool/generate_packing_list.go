package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GeneratePackingListRequest holds args for generating a Packing List
type GeneratePackingListRequest struct {
	PIReference      string  `json:"pi_reference" jsonschema_description:"Reference number of the accepted Proforma Invoice or Sales Contract"`
	TotalCartons     int     `json:"total_cartons" jsonschema_description:"Total number of cartons/boxes packed"`
	TotalPallets     int     `json:"total_pallets" jsonschema_description:"Total number of pallets used for shipping"`
	TotalGrossWeight float64 `json:"total_gross_weight" jsonschema_description:"Total gross weight of the shipment in KG"`
	TotalNetWeight   float64 `json:"total_net_weight" jsonschema_description:"Total net weight of the shipment in KG"`
	TotalVolume      float64 `json:"total_volume" jsonschema_description:"Total volume of the shipment in CBM (Cubic Meters)"`
	MarksAndNumbers  string  `json:"marks_and_numbers" jsonschema_description:"Shipping marks and numbers printed on cartons"`
}

type GeneratePackingListResponse struct {
	Status    string `json:"status"`
	PLNumber  string `json:"pl_number"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// NewGeneratePackingListTool creates a packing list generation tool wrapper
func NewGeneratePackingListTool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_packing_list", "Generate a Packing List (PL) from a confirmed Proforma Invoice based on packaging logic.",
		func(ctx context.Context, req *GeneratePackingListRequest) (*GeneratePackingListResponse, error) {

			plNumber := fmt.Sprintf("PL-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
				"reference": "%s",
				"cartons": %d,
				"pallets": %d,
				"gross_weight": %.2f,
				"net_weight": %.2f,
				"volume_cbm": %.2f,
				"marks": "%s"
			}`, req.PIReference, req.TotalCartons, req.TotalPallets, req.TotalGrossWeight, req.TotalNetWeight, req.TotalVolume, req.MarksAndNumbers)

			return &GeneratePackingListResponse{
				Status:    "draft_pl_created",
				PLNumber:  plNumber,
				Content:   contentStr,
				Timestamp: time.Now().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
