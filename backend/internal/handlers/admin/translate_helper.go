package admin

import (
	"context"

	einotool "candypro/api/internal/pkg/eino/tool"
	tradeSvc "candypro/api/internal/services/trade"
)

// translateBatchFunc 返回 translate_content 工具同源的批量翻译回调。
func (h *Handler) translateBatchFunc() einotool.TranslateBatchFunc {
	if h.aiService == nil {
		return nil
	}
	return einotool.NewTranslateBatchFunc(func(ctx context.Context, sourceData map[string]string, targetLocales []string) (map[string]map[string]string, []string, error) {
		result, err := h.aiService.BatchTranslateFields(ctx, sourceData, targetLocales)
		if err != nil {
			return nil, nil, err
		}
		if result == nil {
			return map[string]map[string]string{}, nil, nil
		}
		return result.Fields, result.Warnings, nil
	})
}

// translateFunc 供 Trade Agent translate_content 工具注入。
func (h *Handler) translateFunc() einotool.TranslateFunc {
	return einotool.TranslateFuncFromBatch(h.translateBatchFunc())
}

// batchTranslateContent 通过 eino translate_content 同款路径执行批量翻译。
func (h *Handler) batchTranslateContent(ctx context.Context, fields map[string]string, targetLocales []string) (*tradeSvc.TranslationResult, error) {
	resp, err := einotool.ExecuteTranslateContent(ctx, h.translateBatchFunc(), fields, targetLocales)
	if err != nil {
		return nil, err
	}
	return &tradeSvc.TranslationResult{
		Fields:   resp.Translations,
		Warnings: resp.Warnings,
	}, nil
}
