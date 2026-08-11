package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type ValidateLCRequest struct {
	LCNumber     string `json:"lc_number" jsonschema_description:"Letter of Credit reference number"`
	LCConditions string `json:"lc_conditions" jsonschema_description:"Key conditions (e.g., latest shipment date, docs required)"`
	DocsJSON     string `json:"docs_json" jsonschema_description:"JSON string of generated documents (PI, CI, PL, COO) to cross-check"`
}

type ValidateLCResponse struct {
	Status        string   `json:"status"`
	Discrepancies []string `json:"discrepancies"`
	IsClean       bool     `json:"is_clean"`
	Timestamp     string   `json:"timestamp"`
}

var lcDocumentRules = map[string]string{
	"invoice":               "Commercial Invoice",
	"packing list":          "Packing List",
	"bill of lading":        "Bill of Lading",
	"certificate of origin": "Certificate of Origin",
	"health certificate":    "Health Certificate",
	"insurance":             "Insurance Certificate",
}

func compactJSONString(raw string) string {
	var decoded interface{}
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return strings.ToLower(strings.TrimSpace(raw))
	}

	encoded, err := json.Marshal(decoded)
	if err != nil {
		return strings.ToLower(strings.TrimSpace(raw))
	}
	return strings.ToLower(string(encoded))
}

func NewValidateLCTool(ctx context.Context) (tool.BaseTool, error) {
	baseTool, err := utils.InferTool("validate_lc_documents", "Cross-check generated trade documents against Letter of Credit (L/C) conditions to prevent discrepancies.",
		func(ctx context.Context, req *ValidateLCRequest) (*ValidateLCResponse, error) {
			if req == nil {
				return nil, fmt.Errorf("request is required")
			}
			if strings.TrimSpace(req.LCNumber) == "" {
				return nil, fmt.Errorf("lc_number is required")
			}

			conditionText := strings.ToLower(strings.TrimSpace(req.LCConditions))
			documentText := compactJSONString(req.DocsJSON)
			discrepancies := []string{}

			if documentText == "" || documentText == "{}" || documentText == "[]" {
				discrepancies = append(discrepancies, "Document payload is empty.")
			}

			for keyword, label := range lcDocumentRules {
				if strings.Contains(conditionText, keyword) && !strings.Contains(documentText, keyword) {
					discrepancies = append(discrepancies, fmt.Sprintf("%s required by L/C conditions but missing in submitted documents.", label))
				}
			}

			if strings.Contains(conditionText, "latest shipment date") && !strings.Contains(documentText, "shipment") {
				discrepancies = append(discrepancies, "L/C includes latest shipment date condition, but shipment evidence is not present in submitted documents.")
			}

			isClean := len(discrepancies) == 0
			status := "discrepancy_detected"
			if isClean {
				status = "lc_validation_clean"
			}

			return &ValidateLCResponse{
				Status:        status,
				Discrepancies: discrepancies,
				IsClean:       isClean,
				Timestamp:     time.Now().UTC().Format(time.RFC3339),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
