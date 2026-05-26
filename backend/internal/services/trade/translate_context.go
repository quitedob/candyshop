package trade

import (
	"context"

	einotool "candypro/api/internal/pkg/eino/tool"
)

// WithDirectTranslateTool 标记 translate_content 工具内部直连 JSON 路径，避免 Agent 递归。
func WithDirectTranslateTool(ctx context.Context) context.Context {
	return einotool.WithDirectTranslate(ctx)
}

// IsDirectTranslateTool 是否为 translate_content 直连翻译上下文。
func IsDirectTranslateTool(ctx context.Context) bool {
	return einotool.IsDirectTranslate(ctx)
}
