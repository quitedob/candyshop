package trade

import (
	"context"
	"encoding/json"
	"strings"
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

// BatchTranslateFields translates a set of named text fields into each target locale.
// sourceData is a map of field name → source text.
// Returns a map of locale → field name → translated text.
func (s *AIService) BatchTranslateFields(ctx context.Context, sourceData map[string]string, targetLocales []string) (map[string]map[string]string, error) {
	result := make(map[string]map[string]string)

	for _, locale := range targetLocales {
		prompt := buildTranslationPrompt(sourceData, locale)
		raw, err := s.GenerateJSON(ctx, prompt)
		if err != nil {
			return result, err
		}
		var fields map[string]string
		if err := json.Unmarshal([]byte(raw), &fields); err != nil {
			raw = extractJSONBlock(raw)
			if err := json.Unmarshal([]byte(raw), &fields); err != nil {
				continue
			}
		}
		result[locale] = fields
	}
	return result, nil
}

// TranslationLocaleName returns a human-readable name for a locale code.
func TranslationLocaleName(locale string) string {
	if name, ok := translationLocaleNames[locale]; ok {
		return name
	}
	// Try lowercased
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
// sourceLang may be empty (auto-detect).
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
