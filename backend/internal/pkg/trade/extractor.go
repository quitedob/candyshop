package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"candypro/api/internal/pkg/eino/prompts"

	"github.com/cloudwego/eino/components/model"
)

// ExtractDocument extracts structured data from the document
func ExtractDocument(ctx context.Context, cm model.ChatModel, input *ExtractedData) (*ExtractedData, error) {

	rawText, _ := input.Data["raw_text"].(string)

	msgs, err := prompts.ExtractionPrompt.Format(ctx, map[string]any{
		"doc_type": input.DocumentType,
		"content":  rawText,
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

	var extracted map[string]any
	if err := json.Unmarshal([]byte(content), &extracted); err != nil {
		return nil, fmt.Errorf("failed to parse extraction: %v, content: %s", err, content)
	}

	return &ExtractedData{
		DocumentType: input.DocumentType,
		Data:         extracted,
	}, nil
}
