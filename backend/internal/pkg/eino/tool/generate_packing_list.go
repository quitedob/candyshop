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

type GeneratePackingListRequest struct {
	PIReference      string  `json:"pi_reference" jsonschema_description:"Reference number of the accepted Proforma Invoice or Sales Contract"`
	TotalCartons     int     `json:"total_cartons" jsonschema_description:"Total number of cartons/boxes packed"`
	TotalPallets     int     `json:"total_pallets" jsonschema_description:"Total number of pallets used for shipping"`
	TotalGrossWeight float64 `json:"total_gross_weight" jsonschema_description:"Total gross weight in KG"`
	TotalNetWeight   float64 `json:"total_net_weight" jsonschema_description:"Total net weight in KG"`
	TotalVolume      float64 `json:"total_volume" jsonschema_description:"Total volume in CBM"`
	MarksAndNumbers  string  `json:"marks_and_numbers" jsonschema_description:"Shipping marks and numbers printed on cartons"`
	TradeID          uint    `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GeneratePackingListResponse struct {
	Status   string `json:"status"`
	PLNumber string `json:"pl_number"`
	Content  string `json:"content"`
	SavedTo  string `json:"saved_to,omitempty"`
}

func NewGeneratePackingListTool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_packing_list",
		"Generate a Packing List from a confirmed Proforma Invoice based on packaging logic.",
		func(ctx context.Context, req *GeneratePackingListRequest) (*GeneratePackingListResponse, error) {
			plNumber := fmt.Sprintf("PL-%d", time.Now().Unix())

			contentStr := fmt.Sprintf(`{
	"reference": %q,
	"cartons": %d,
	"pallets": %d,
	"gross_weight": %.2f,
	"net_weight": %.2f,
	"volume_cbm": %.2f,
	"marks": %q
}`, req.PIReference, req.TotalCartons, req.TotalPallets, req.TotalGrossWeight, req.TotalNetWeight, req.TotalVolume, req.MarksAndNumbers)

			resp := &GeneratePackingListResponse{
				Status:   "draft_pl_created",
				PLNumber: plNumber,
				Content:  contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          tradeModels.DocTypePackingList,
					DocNumber:     plNumber,
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
