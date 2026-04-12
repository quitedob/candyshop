package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

const (
	// ComplianceToolName is the Eino tool name exposed to agents.
	ComplianceToolName = "compliance_lookup"
)

// NewComplianceTool wraps ComplianceRetriever as an Eino invokable tool.
func NewComplianceTool(retriever *ComplianceRetriever) (tool.InvokableTool, error) {
	if retriever == nil {
		return nil, fmt.Errorf("compliance retriever is nil")
	}

	return utils.InferTool(
		ComplianceToolName,
		"Look up country and region compliance requirements from local markdown corpus. Use this tool for regulations, certifications, labeling, additives, and packaging compliance.",
		func(ctx context.Context, input *ComplianceQueryInput) (*ComplianceQueryOutput, error) {
			_ = ctx

			if input == nil {
				return nil, fmt.Errorf("input is required")
			}

			query := strings.TrimSpace(input.Query)
			if query == "" {
				query = "food safety authority certification labeling additives packaging import registration"
			}

			matches := retriever.Search(input.Country, query, input.TopK)
			officialDomains := officialDomainsForCountry(input.Country)
			output := &ComplianceQueryOutput{
				Matches:               matches,
				HasOfficialEvidence:   hasOfficialMatches(matches),
				OfficialSourceDomains: officialDomains,
			}

			if len(matches) == 0 {
				output.Answer = "No exact match found in local compliance markdown corpus. Verify with the official regulatory authority website for the target country."
				return output, nil
			}

			output.Answer = retriever.FormatAnswer(input.Country, query, matches)
			if !output.HasOfficialEvidence {
				output.Answer += "\nNo official-domain evidence matched the current retrieval. Please verify latest rules directly on authority websites."
			}
			return output, nil
		},
	)
}
