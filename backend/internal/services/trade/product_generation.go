package trade

import (
	"context"

	"candypro/api/internal/pkg/eino"
)

// GenerateProduct uses AI to generate structured product data from a free-text description.
// Delegates to the core eino.Client for prompt engineering and JSON parsing.
func (s *AIService) GenerateProduct(ctx context.Context, description, language string) (*eino.ProductGenResult, string, error) {
	return s.Client.GenerateProduct(ctx, description, language)
}
