package tool

import (
	"context"
	"fmt"
	"time"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type GenerateSLIRequest struct {
	ShipperName     string `json:"shipper_name" jsonschema_description:"Name of the exporting company"`
	ConsigneeName   string `json:"consignee_name" jsonschema_description:"Name of the importing company"`
	PortOfLoading   string `json:"port_of_loading" jsonschema_description:"Port of departure"`
	PortOfDischarge string `json:"port_of_discharge" jsonschema_description:"Port of destination"`
	CargoDetails    string `json:"cargo_details" jsonschema_description:"Brief description of cargo, volume, and weight"`
	IsFCL           bool   `json:"is_fcl" jsonschema_description:"True if Full Container Load, False if LCL"`
}

type GenerateSLIResponse struct {
	Status    string `json:"status"`
	DocNo     string `json:"doc_no"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func NewGenerateSLITool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("generate_shipper_letter_of_instruction", "Generate a Shipper's Letter of Instruction (SLI) for the freight forwarder.",
		func(ctx context.Context, req *GenerateSLIRequest) (*GenerateSLIResponse, error) {

			docNo := fmt.Sprintf("SLI-%d", time.Now().Unix())
			contentStr := fmt.Sprintf(`{
				"shipper": "%s",
				"consignee": "%s",
				"origin_port": "%s",
				"destination_port": "%s",
				"cargo": "%s",
				"fcl_status": %t
			}`, req.ShipperName, req.ConsigneeName, req.PortOfLoading, req.PortOfDischarge, req.CargoDetails, req.IsFCL)

			return &GenerateSLIResponse{
				Status:    "sli_drafted",
				DocNo:     docNo,
				Content:   contentStr,
				Timestamp: time.Now().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
