package trade

import (
	"context"

	"candypro/api/internal/pkg/eino"
)

// ContentGenResult is re-exported from eino for backward compatibility.
type ContentGenResult = eino.ContentGenResult

// GenerateContent uses AI to generate blog/case-study content.
// Delegates to the core eino.Client for prompt engineering and JSON parsing.
func (s *AIService) GenerateContent(ctx context.Context, topic, contentType, language string) (*ContentGenResult, string, error) {
	return s.Client.GenerateContent(ctx, topic, contentType, language)
}
