package tool

import (
	"context"
	"fmt"

	einotool "github.com/cloudwego/eino-examples/adk/common/tool"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// TranslateFunc 批量翻译回调，供 Trade Agent translate_content 工具注入。
// sourceData: 字段名 → 源文本；返回 locale → 字段名 → 译文。
type TranslateFunc func(ctx context.Context, sourceData map[string]string, targetLocales []string) (map[string]map[string]string, error)

// TranslateBatchResult 批量翻译完整结果（含 per-locale 警告）。
type TranslateBatchResult struct {
	Translations map[string]map[string]string
	Warnings     []string
}

// TranslateBatchFunc 带警告的批量翻译回调，供 HTTP handler 直连调用。
type TranslateBatchFunc func(ctx context.Context, sourceData map[string]string, targetLocales []string) (*TranslateBatchResult, error)

// BatchTranslateBackend 由 handler 注入的底层翻译实现（避免 tool 包依赖 services）。
type BatchTranslateBackend func(ctx context.Context, sourceData map[string]string, targetLocales []string) (fields map[string]map[string]string, warnings []string, err error)

type directTranslateKey struct{}

// WithDirectTranslate 标记 translate_content 工具内部直连 JSON 路径，避免 Agent 递归。
func WithDirectTranslate(ctx context.Context) context.Context {
	return context.WithValue(ctx, directTranslateKey{}, true)
}

// IsDirectTranslate 是否为 translate_content 直连翻译上下文。
func IsDirectTranslate(ctx context.Context) bool {
	v, _ := ctx.Value(directTranslateKey{}).(bool)
	return v
}

// TranslateContentRequest translate_content 工具入参。
type TranslateContentRequest struct {
	Fields        map[string]string `json:"fields" jsonschema_description:"Map of field names to source text to translate"`
	TargetLocales []string          `json:"target_locales" jsonschema_description:"List of target locale codes, e.g. [en, ar, ja]"`
}

// TranslateContentResponse translate_content 工具出参。
type TranslateContentResponse struct {
	Translations map[string]map[string]string `json:"translations"`
	LocalesDone  int                            `json:"locales_done"`
	Warnings     []string                       `json:"warnings,omitempty"`
}

// NewTranslateBatchFunc 将底层 BatchTranslate 包装为 eino 规范的批量翻译回调（自动标记直连上下文）。
func NewTranslateBatchFunc(backend BatchTranslateBackend) TranslateBatchFunc {
	return func(ctx context.Context, sourceData map[string]string, targetLocales []string) (*TranslateBatchResult, error) {
		if backend == nil {
			return nil, fmt.Errorf("translate backend is required")
		}
		ctx = WithDirectTranslate(ctx)
		fields, warnings, err := backend(ctx, sourceData, targetLocales)
		if err != nil {
			return nil, err
		}
		if fields == nil {
			fields = map[string]map[string]string{}
		}
		return &TranslateBatchResult{Translations: fields, Warnings: warnings}, nil
	}
}

// TranslateFuncFromBatch 将 TranslateBatchFunc 适配为 Agent 工具所需的 TranslateFunc。
func TranslateFuncFromBatch(batchFn TranslateBatchFunc) TranslateFunc {
	return func(ctx context.Context, sourceData map[string]string, targetLocales []string) (map[string]map[string]string, error) {
		result, err := batchFn(ctx, sourceData, targetLocales)
		if err != nil {
			return nil, err
		}
		return result.Translations, nil
	}
}

// ExecuteTranslateContent 直连调用 translate_content 同款逻辑（HTTP API / 生成后自动翻译）。
func ExecuteTranslateContent(ctx context.Context, batchFn TranslateBatchFunc, fields map[string]string, targetLocales []string) (*TranslateContentResponse, error) {
	if batchFn == nil {
		return nil, fmt.Errorf("translate batch func is required")
	}
	if len(fields) == 0 {
		return nil, fmt.Errorf("fields is required and must not be empty")
	}
	if len(targetLocales) == 0 {
		return nil, fmt.Errorf("target_locales is required and must not be empty")
	}

	result, err := batchFn(ctx, fields, targetLocales)
	if err != nil {
		return nil, fmt.Errorf("translation failed: %w", err)
	}
	if result == nil {
		result = &TranslateBatchResult{Translations: map[string]map[string]string{}}
	}

	return &TranslateContentResponse{
		Translations: result.Translations,
		LocalesDone:  len(result.Translations),
		Warnings:     result.Warnings,
	}, nil
}

// NewTranslateContentTool 创建 Eino translate_content InvokableTool。
func NewTranslateContentTool(ctx context.Context, translateFn TranslateFunc) (tool.BaseTool, error) {
	if translateFn == nil {
		return nil, fmt.Errorf("translateFn is required")
	}

	baseTool, err := utils.InferTool("translate_content",
		"Translate named text fields into multiple target languages. Use this when you need to localize product names, descriptions, summaries, or other content fields for international markets. Returns translations keyed by locale code.",
		func(ctx context.Context, req *TranslateContentRequest) (*TranslateContentResponse, error) {
			if req == nil || len(req.Fields) == 0 {
				return nil, fmt.Errorf("fields is required and must not be empty")
			}
			if len(req.TargetLocales) == 0 {
				return nil, fmt.Errorf("target_locales is required and must not be empty")
			}

			translations, err := translateFn(ctx, req.Fields, req.TargetLocales)
			if err != nil {
				return nil, fmt.Errorf("translation failed: %w", err)
			}

			return &TranslateContentResponse{
				Translations: translations,
				LocalesDone:  len(translations),
			}, nil
		})
	if err != nil {
		return nil, err
	}

	return &einotool.InvokableReviewEditTool{InvokableTool: baseTool}, nil
}
