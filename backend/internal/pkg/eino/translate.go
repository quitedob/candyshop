package eino

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// translationLocaleNames 翻译目标语言的人类可读名称。
var translationLocaleNames = map[string]string{
	"en":    "English",
	"zh":    "Chinese (Simplified)",
	"ar":    "Arabic",
	"es":    "Spanish",
	"fr":    "French",
	"de":    "German",
	"ja":    "Japanese",
	"ko":    "Korean",
	"zh-tw": "Traditional Chinese",
	"th":    "Thai",
	"vi":    "Vietnamese",
	"id":    "Indonesian",
	"ms":    "Malay",
}

// TranslationLocaleName 返回 locale 代码对应的语言名称。
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

// BuildTranslationPrompt 构建批量字段翻译 prompt，含 JSON 样例以配合 response_format=json_object。
func BuildTranslationPrompt(data map[string]string, targetLocale string) string {
	keys := make([]string, 0, len(data))
	for k, v := range data {
		if strings.TrimSpace(v) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("Translate the following product/content fields into ")
	sb.WriteString(TranslationLocaleName(targetLocale))
	sb.WriteString(`.

Return ONLY a valid JSON object (no markdown fences, no explanations).
Every value MUST be a JSON string — never use bare arrays or objects as values.

If a field value is a JSON array string (e.g. flavors or shapes), translate each item and return the result as a JSON array string.

EXAMPLE INPUT:
name: Strawberry Gummies
flavors: ["Strawberry","Apple"]

EXAMPLE JSON OUTPUT:
{
  "name": "Translated product name",
  "flavors": "[\"Translated flavor 1\",\"Translated flavor 2\"]"
}

Fields to translate:
`)
	for _, key := range keys {
		sb.WriteString(key)
		sb.WriteString(": ")
		sb.WriteString(data[key])
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ParseTranslationFieldsJSON 解析模型返回的翻译 JSON，兼容数组/嵌套等非常规格式。
func ParseTranslationFieldsJSON(raw string) (map[string]string, error) {
	text := cleanJSONBlock(strings.TrimSpace(raw))
	if text == "" {
		return map[string]string{}, nil
	}

	// 顶层为数组时取第一个对象
	if strings.HasPrefix(text, "[") {
		var arr []map[string]json.RawMessage
		if err := json.Unmarshal([]byte(text), &arr); err == nil && len(arr) > 0 {
			return rawMessageMapToStrings(arr[0])
		}
	}

	var flexible map[string]json.RawMessage
	if err := json.Unmarshal([]byte(text), &flexible); err != nil {
		return nil, fmt.Errorf("parse translation JSON: %w", err)
	}

	// 兼容 {"translations": {...}} 包装
	if inner, ok := flexible["translations"]; ok && len(flexible) == 1 {
		var nested map[string]json.RawMessage
		if err := json.Unmarshal(inner, &nested); err == nil {
			return rawMessageMapToStrings(nested)
		}
	}

	return rawMessageMapToStrings(flexible)
}

// rawMessageMapToStrings 将 map[string]json.RawMessage 转为 map[string]string。
func rawMessageMapToStrings(src map[string]json.RawMessage) (map[string]string, error) {
	out := make(map[string]string, len(src))
	for k, v := range src {
		s, err := coerceJSONValueToString(v)
		if err != nil {
			return nil, fmt.Errorf("parse field %q: %w", k, err)
		}
		out[k] = s
	}
	return out, nil
}

// coerceJSONValueToString 将 JSON 值统一转为字符串；数组/对象序列化为 JSON 字符串。
func coerceJSONValueToString(raw json.RawMessage) (string, error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return "", nil
	}

	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, nil
	}

	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		strs := make([]string, 0, len(arr))
		for _, item := range arr {
			itemStr, err := coerceJSONValueToString(item)
			if err != nil {
				return "", err
			}
			strs = append(strs, itemStr)
		}
		b, err := json.Marshal(strs)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	var obj map[string]json.RawMessage
	if err := json.Unmarshal(raw, &obj); err == nil {
		b, err := json.Marshal(obj)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	var num json.Number
	if err := json.Unmarshal(raw, &num); err == nil {
		return num.String(), nil
	}

	var b bool
	if err := json.Unmarshal(raw, &b); err == nil {
		return fmt.Sprintf("%t", b), nil
	}

	return "", fmt.Errorf("unsupported JSON value type")
}

// TranslateFieldsJSON 直连 JSON 模式翻译一批字段到目标语言。
func (c *Client) TranslateFieldsJSON(ctx context.Context, sourceData map[string]string, targetLocale string) (map[string]string, error) {
	prompt := BuildTranslationPrompt(sourceData, targetLocale)
	raw, err := c.GenerateJSON(ctx, prompt)
	if err != nil {
		return nil, err
	}
	return ParseTranslationFieldsJSON(raw)
}
