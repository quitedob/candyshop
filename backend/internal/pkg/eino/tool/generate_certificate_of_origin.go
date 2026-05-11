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

type GenerateCOORequest struct {
	CountryOfOrigin    string `json:"country_of_origin" jsonschema_description:"The country where the goods were manufactured"`
	DestinationCountry string `json:"destination_country" jsonschema_description:"The destination country of the shipment"`
	FTAType            string `json:"fta_type" jsonschema_description:"Type of Free Trade Agreement e.g., FORM E, FORM A, RCEP"`
	HSCode             string `json:"hs_code" jsonschema_description:"Harmonized System (HS) code classification for the candy/food product"`
	TradeID            uint   `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GenerateCOOResponse struct {
	Status        string `json:"status"`
	CertificateNo string `json:"certificate_no"`
	Content       string `json:"content"`
	SavedTo       string `json:"saved_to,omitempty"`
}

func NewGenerateCertificateOfOriginTool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_certificate_of_origin",
		"Generate a Certificate of Origin (COO) draft based on production location, destination, and Free Trade Agreements.",
		func(ctx context.Context, req *GenerateCOORequest) (*GenerateCOOResponse, error) {
			certNo := fmt.Sprintf("COO-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
	"origin": %q,
	"destination": %q,
	"fta_type": %q,
	"hs_code": %q
}`, req.CountryOfOrigin, req.DestinationCountry, req.FTAType, req.HSCode)

			resp := &GenerateCOOResponse{
				Status:        "draft_coo_created",
				CertificateNo: certNo,
				Content:       contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          tradeModels.DocTypeOriginCertificate,
					DocNumber:     certNo,
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
