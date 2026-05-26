package trade

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	eino "candypro/api/internal/pkg/eino"
	einotool "candypro/api/internal/pkg/eino/tool"
)

const (
	maxConcurrentTranslations = 3
	translationTimeout        = 600 * time.Second
	xlsxTranslateBatchSize    = 30
)

// TranslationResult holds the translation output and any per-locale warnings.
type TranslationResult struct {
	Fields   map[string]map[string]string
	Warnings []string
}

// BatchTranslateFields 批量翻译字段，与 Trade Agent translate_content 工具同源。
// 外部 API 走 Trade Agent Generate；工具内部走直连 JSON，避免递归。
func (s *AIService) BatchTranslateFields(ctx context.Context, sourceData map[string]string, targetLocales []string) (*TranslationResult, error) {
	result := &TranslationResult{
		Fields: make(map[string]map[string]string),
	}

	if len(targetLocales) == 0 {
		return result, nil
	}

	var (
		wg        sync.WaitGroup
		mu        sync.Mutex
		sem       = make(chan struct{}, maxConcurrentTranslations)
		parentErr error
	)

	for _, locale := range targetLocales {
		wg.Add(1)
		go func(loc string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			localeCtx, cancel := context.WithTimeout(ctx, translationTimeout)
			defer cancel()

			fields, err := s.translateOneLocale(localeCtx, sourceData, loc)
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				if ctx.Err() != nil {
					parentErr = ctx.Err()
				}
				warning := fmt.Sprintf("translate to %s failed: %v", loc, err)
				result.Warnings = append(result.Warnings, warning)
				log.Printf("[WARN] BatchTranslateFields: %s", warning)
				return
			}
			result.Fields[loc] = fields
		}(locale)
	}

	wg.Wait()

	if parentErr != nil {
		return result, parentErr
	}
	return result, nil
}

func (s *AIService) translateOneLocale(ctx context.Context, sourceData map[string]string, locale string) (map[string]string, error) {
	if einotool.IsDirectTranslate(ctx) || !s.HasAgent() {
		return s.translateOneLocaleJSON(ctx, sourceData, locale)
	}
	return s.translateOneLocaleViaAgent(ctx, sourceData, locale)
}

// translateOneLocaleViaAgent 经 Trade Agent 生成 JSON 翻译结果。
func (s *AIService) translateOneLocaleViaAgent(ctx context.Context, sourceData map[string]string, locale string) (map[string]string, error) {
	raw, err := s.generateStructuredJSON(ctx, eino.BuildTranslationPrompt(sourceData, locale))
	if err != nil {
		return s.translateOneLocaleJSON(ctx, sourceData, locale)
	}
	fields, parseErr := eino.ParseTranslationFieldsJSON(raw)
	if parseErr != nil {
		return s.translateOneLocaleJSON(ctx, sourceData, locale)
	}
	return fields, nil
}

// translateOneLocaleJSON 工具内部直连 JSON 模式（translate_content 工具回调路径）。
func (s *AIService) translateOneLocaleJSON(ctx context.Context, sourceData map[string]string, locale string) (map[string]string, error) {
	return s.TranslateFieldsJSON(ctx, sourceData, locale)
}

// TranslateSingleText 单条翻译，经 Trade Agent Generate。
func (s *AIService) TranslateSingleText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if sourceLang == "" {
		sourceLang = "auto"
	}
	prompt := "Translate the following text from " + sourceLang + " to " + eino.TranslationLocaleName(targetLang) +
		". Return only the translated content without explanations:\n\n" + text
	if s.HasAgent() {
		return s.Generate(ctx, prompt)
	}
	return s.GenerateDirect(ctx, prompt)
}

// BatchTranslateTexts XLSX 等场景批量翻译，复用 translate_content 同款 BatchTranslateFields。
func (s *AIService) BatchTranslateTexts(ctx context.Context, items map[string]string, sourceLang, targetLang string) (map[string]string, []string, error) {
	result := make(map[string]string, len(items))
	warnings := make([]string, 0)

	if len(items) == 0 {
		return result, warnings, nil
	}

	ctx = einotool.WithDirectTranslate(ctx)

	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := 0; i < len(keys); i += xlsxTranslateBatchSize {
		end := i + xlsxTranslateBatchSize
		if end > len(keys) {
			end = len(keys)
		}
		chunk := make(map[string]string, end-i)
		for _, k := range keys[i:end] {
			chunk[k] = items[k]
		}

		transResult, err := s.BatchTranslateFields(ctx, chunk, []string{targetLang})
		if err != nil {
			warnings = append(warnings, err.Error())
			continue
		}
		if transResult != nil {
			warnings = append(warnings, transResult.Warnings...)
			if localeFields, ok := transResult.Fields[targetLang]; ok {
				for k, v := range localeFields {
					result[k] = v
				}
			}
		}
	}
	return result, warnings, nil
}

func extractJSONBlock(raw string) string {
	start := strings.Index(raw, "```json")
	if start == -1 {
		start = strings.Index(raw, "```")
	}
	if start != -1 {
		start = strings.Index(raw[start:], "\n")
		if start != -1 {
			raw = raw[start:]
		}
		end := strings.LastIndex(raw, "```")
		if end != -1 {
			raw = raw[:end]
		}
	}
	return strings.TrimSpace(raw)
}
