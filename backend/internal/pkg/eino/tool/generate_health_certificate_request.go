package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	tradeModels "candypro/api/internal/models/trade"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateHealthCertRequest struct {
	ProductName        string `json:"product_name" jsonschema_description:"Name of the candy or food product"`
	DestinationCountry string `json:"destination_country" jsonschema_description:"Target country importing the product"`
	ProducerName       string `json:"producer_name" jsonschema_description:"Name of the manufacturing facility (e.g., CandyPro)"`
	HealthAuthority    string `json:"health_authority" jsonschema_description:"Destination authority requiring the cert (e.g., FDA, SFDA)"`
	IssueDate          string `json:"issue_date" jsonschema_description:"Requested issuance date (YYYY-MM-DD)"`
	TradeID            uint   `json:"trade_id" jsonschema_description:"Transaction ID to attach this document to"`
}

type GenerateHealthCertResponse struct {
	Status    string `json:"status"`
	RequestNo string `json:"request_no"`
	Content   string `json:"content"`
	SavedTo   string `json:"saved_to,omitempty"`
}

func NewGenerateHealthCertificateTool(ctx context.Context, persister DocumentPersister) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_health_certificate_request",
		"Generate a request for health/sanitary certificate targeting a specific country's regulatory body.",
		func(ctx context.Context, req *GenerateHealthCertRequest) (*GenerateHealthCertResponse, error) {
			reqNo := fmt.Sprintf("HC-REQ-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
	"product": %q,
	"destination": %q,
	"producer": %q,
	"authority": %q,
	"issue_date": %q
}`, req.ProductName, req.DestinationCountry, req.ProducerName, req.HealthAuthority, req.IssueDate)

			resp := &GenerateHealthCertResponse{
				Status:    "health_cert_request_drafted",
				RequestNo: reqNo,
				Content:   contentStr,
			}

			if persister != nil && req.TradeID > 0 {
				doc := &tradeModels.TradeDocument{
					TransactionID: req.TradeID,
					Type:          tradeModels.DocTypeHealthCertificate,
					DocNumber:     reqNo,
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
