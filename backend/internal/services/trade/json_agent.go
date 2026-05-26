package trade

import (
	"context"
	"strings"

	einotool "candypro/api/internal/pkg/eino/tool"
)

// generateStructuredJSON 优先经 Trade Agent 返回 JSON；失败或无 Agent 时降级 GenerateJSON。
func (s *AIService) generateStructuredJSON(ctx context.Context, prompt string) (string, error) {
	if s != nil && s.HasAgent() && !einotool.IsDirectTranslate(ctx) {
		agentPrompt := strings.TrimSpace(prompt) + "\n\nDo not call tools. Return ONLY valid JSON."
		if raw, err := s.Generate(ctx, agentPrompt); err == nil && strings.TrimSpace(raw) != "" {
			return raw, nil
		}
	}
	return s.GenerateJSON(ctx, prompt)
}
