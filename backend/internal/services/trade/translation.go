package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

var translationLocaleNames = map[string]string{
	"en": "English",
	"zh": "Chinese (Simplified)",
	"ar": "Arabic",
	"es": "Spanish",
	"fr": "French",
	"de": "German",
	"ja": "Japanese",
	"ko": "Korean",
	"zh-tw": "Traditional Chinese",
	"th":    "Thai",
	"vi":    "Vietnamese",
	"id":    "Indonesian",
	"ms":    "Malay",
}

const (
	maxConcurrentTranslations = 3
	translationTimeout        = 30 * time.Second
)

// TranslationResult holds the translation output and any per-locale warnings.
type TranslationResult struct {
	Fields   map[string]map[string]string
	Warnings []string
}

// BatchTranslateFields translates a set of named text fields into each target locale.
// Runs translations concurrently (up to maxConcurrentTranslations) with per-locale timeout.
// Failed locales are reported as warnings; partial results are always returned.
func (s *AIService) BatchTranslateFields(ctx context.Context, sourceData map[string]string, targetLocales []string) (*TranslationResult, error) {
	result := &TranslationResult{
		Fields: make(map[string]map[string]string),
	}

	if len(targetLocales) == 0 {
		return result, nil
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		sem      = make(chan struct{}, maxConcurrentTranslations)
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
	prompt := buildTranslationPrompt(sourceData, locale)
	raw, err := s.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, err
	}
	var fields map[string]string
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		raw = extractJSONBlock(raw)
		if err := json.Unmarshal([]byte(raw), &fields); err != nil {
			return nil, fmt.Errorf("parse translation JSON for %s: %w", locale, err)
		}
	}
	return fields, nil
}

// TranslationLocaleName returns a human-readable name for a locale code.
func TranslationLocaleName(locale string) string {
	if name, ok := translationLocaleNames[locale]; ok {
		return name
	}
	lower := strings.ToLower(locale)
	if name, ok := translationLocaleNames[lower]; ok {
		return name
	}
	return locale
}

func buildTranslationPrompt(data map[string]string, targetLocale string) string {
	var sb strings.Builder
	sb.WriteString("Translate the following fields to ")
	sb.WriteString(TranslationLocaleName(targetLocale))
	sb.WriteString(". Return ONLY a JSON object with the translated values. Do not include explanations or markdown.\n\nFields:\n")
	for key, value := range data {
		if strings.TrimSpace(value) == "" {
			continue
		}
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(value)
		sb.WriteString("\n")
	}
	return sb.String()
}

// TranslateSingleText translates a single text string to the target language.
func (s *AIService) TranslateSingleText(ctx context.Context, text, sourceLang, targetLang string) (string, error) {
	if sourceLang == "" {
		sourceLang = "auto"
	}
	prompt := "Translate the following text from " + sourceLang + " to " + targetLang +
		". Return only the translated content without explanations:\n\n" + text
	return s.Generate(ctx, prompt)
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
