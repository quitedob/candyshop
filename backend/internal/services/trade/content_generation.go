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

// ReviseContent applies targeted edits to an existing draft based on admin feedback.
func (s *AIService) ReviseContent(ctx context.Context, contentType, language, instruction, selectedText string, focusFields []string, current map[string]string) (*eino.ContentReviseResult, string, error) {
	return s.Client.ReviseContent(ctx, contentType, language, instruction, selectedText, focusFields, current)
}

// InlineEditContent rewrites only the selected excerpt (Cursor-style inline edit).
func (s *AIService) InlineEditContent(ctx context.Context, language, instruction, selectedText, selectedHTML, contextBefore, contextAfter, fieldType string) (*eino.InlineEditResult, string, error) {
	return s.Client.InlineEditContent(ctx, language, instruction, selectedText, selectedHTML, contextBefore, contextAfter, fieldType)
}
