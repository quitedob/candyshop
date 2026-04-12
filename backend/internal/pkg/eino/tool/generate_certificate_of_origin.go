package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GenerateCOORequest holds args for generating a Certificate of Origin
type GenerateCOORequest struct {
	CountryOfOrigin    string `json:"country_of_origin" jsonschema_description:"The country where the goods were manufactured"`
	DestinationCountry string `json:"destination_country" jsonschema_description:"The destination country of the shipment"`
	FTAType            string `json:"fta_type" jsonschema_description:"Type of Free Trade Agreement to apply under e.g., FORM E, FORM A, RCEP"`
	HSCode             string `json:"hs_code" jsonschema_description:"Harmonized System (HS) code classification for the candy/food product"`
}

type GenerateCOOResponse struct {
	Status        string `json:"status"`
	CertificateNo string `json:"certificate_no"`
	Content       string `json:"content"`
	Timestamp     string `json:"timestamp"`
}

// NewGenerateCertificateOfOriginTool creates a COO generation tool wrapper
func NewGenerateCertificateOfOriginTool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_certificate_of_origin", "Generate a Certificate of Origin (COO) draft based on production location, destination, and Free Trade Agreements.",
		func(ctx context.Context, req *GenerateCOORequest) (*GenerateCOOResponse, error) {

			certNo := fmt.Sprintf("COO-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
				"origin": "%s",
				"destination": "%s",
				"fta_type": "%s",
				"hs_code": "%s"
			}`, req.CountryOfOrigin, req.DestinationCountry, req.FTAType, req.HSCode)

			return &GenerateCOOResponse{
				Status:        "draft_coo_created",
				CertificateNo: certNo,
				Content:       contentStr,
				Timestamp:     time.Now().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
