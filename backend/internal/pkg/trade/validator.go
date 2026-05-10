package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"candypro/api/internal/pkg/eino/prompts"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// NewValidatorNode creates a lambda node that validates multiple extracted documents
func NewValidatorNode(cm model.ChatModel) *compose.Lambda {
	return compose.InvokableLambda(func(ctx context.Context, req *HitlRequest) (*HitlRequest, error) {

		docs := req.Data

		if len(docs) == 0 {
			req.Report = &VerificationReport{
				Passed:          false,
				Conflicts:       []string{"No documents provided"},
				ManualReviewReq: true,
			}
			return req, nil
		}

		// Convert to JSON string for the prompt
		docsJSON, _ := json.MarshalIndent(docs, "", "  ")

		msgs, err := prompts.ValidationPrompt.Format(ctx, map[string]any{
			"target_doc":     "All Documents",
			"reference_data": string(docsJSON),
		})
		if err != nil {
			return nil, err
		}

		resp, err := cm.Generate(ctx, msgs)
		if err != nil {
			return nil, err
		}

		content := strings.TrimSpace(resp.Content)
		if strings.HasPrefix(content, "```json") {
			content = strings.TrimPrefix(content, "```json")
			content = strings.TrimSuffix(content, "```")
		} else if strings.HasPrefix(content, "```") {
			content = strings.TrimPrefix(content, "```")
			content = strings.TrimSuffix(content, "```")
		}

		var report VerificationReport
		if err := json.Unmarshal([]byte(content), &report); err != nil {
			return nil, fmt.Errorf("failed to parse validation report: %v, content: %s", err, content)
		}

		// Always require manual review if it didn't pass
		if !report.Passed {
			report.ManualReviewReq = true
		}

		req.Report = &report
		return req, nil
	})
}
