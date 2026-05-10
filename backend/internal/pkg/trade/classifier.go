package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"candypro/api/internal/pkg/eino/prompts"

	"github.com/cloudwego/eino/components/model"
)

// ClassifyDocument classifies the given document
func ClassifyDocument(ctx context.Context, cm model.ChatModel, doc *RawDocument) (*ExtractedData, error) {
	msgs, err := prompts.DocumentClassificationPrompt.Format(ctx, map[string]any{
		"text": doc.Text,
	})
	if err != nil {
		return nil, err
	}

	resp, err := cm.Generate(ctx, msgs)
	if err != nil {
		return nil, err
	}

	// Parse the result
	content := strings.TrimSpace(resp.Content)
	// Basic cleanup for markdown
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}

	var result ClassifierResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, fmt.Errorf("failed to parse classification: %v, content: %s", err, content)
	}

	return &ExtractedData{
		DocumentType: result.DocumentType,
		Data: map[string]any{
			"metadata": doc.Metadata,
			"raw_text": doc.Text, // carry forward for the extractor
		},
	}, nil
}
