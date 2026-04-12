package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateHealthCertRequest struct {
	ProductName        string `json:"product_name" jsonschema_description:"Name of the candy or food product"`
	DestinationCountry string `json:"destination_country" jsonschema_description:"Target country importing the product"`
	ProducerName       string `json:"producer_name" jsonschema_description:"Name of the manufacturing facility (e.g., CandyPro)"`
	HealthAuthority    string `json:"health_authority" jsonschema_description:"Destination authority requiring the cert (e.g., FDA, SFDA)"`
	IssueDate          string `json:"issue_date" jsonschema_description:"Requested issuance date (YYYY-MM-DD)"`
}

type GenerateHealthCertResponse struct {
	Status    string `json:"status"`
	RequestNo string `json:"request_no"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func NewGenerateHealthCertificateTool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_health_certificate_request", "Generate a request for health/sanitary certificate targeting a specific country's regulatory body.",
		func(ctx context.Context, req *GenerateHealthCertRequest) (*GenerateHealthCertResponse, error) {

			reqNo := fmt.Sprintf("HC-REQ-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
				"product": "%s",
				"destination": "%s",
				"producer": "%s",
				"authority": "%s",
				"issue_date": "%s"
			}`, req.ProductName, req.DestinationCountry, req.ProducerName, req.HealthAuthority, req.IssueDate)

			return &GenerateHealthCertResponse{
				Status:    "health_cert_request_drafted",
				RequestNo: reqNo,
				Content:   contentStr,
				Timestamp: time.Now().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
