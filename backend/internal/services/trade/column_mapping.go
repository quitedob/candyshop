package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
)

// BuildColumnMapping 使用 Trade Agent 将 XLSX 表头映射到 CandyPro 产品字段。
// 返回 field → column header 名称；失败时返回 nil（调用方仍可用启发式映射）。
func (s *AIService) BuildColumnMapping(ctx context.Context, headers []string, sampleRows [][]string) map[string]string {
	raw, err := s.generateStructuredJSON(ctx, buildColumnMappingPrompt(headers, sampleRows))
	if err != nil || raw == "" {
		if err != nil {
			slog.Warn("column mapping AI generation failed", "error", err)
		}
		return nil
	}

	var aiMapping map[string]string
	text := extractJSONBlock(strings.TrimSpace(raw))
	if err := json.Unmarshal([]byte(text), &aiMapping); err != nil {
		slog.Warn("column mapping JSON parse failed", "error", err)
		return nil
	}
	return aiMapping
}

func buildColumnMappingPrompt(headers []string, sampleRows [][]string) string {
	var sb strings.Builder
	sb.WriteString("You are a product data importer. Map the following XLSX columns to CandyPro product fields.\n\n")
	sb.WriteString("Available fields: id, name, category, categorySlug, basePrice, moq, stockQuantity, leadTime, halalCertified, oemAvailable, hsCode, shelfLife, storage, status, thumbnail, images, flavors, shapes, ingredients, allergens, certifications, description, summary\n\n")
	sb.WriteString("Column headers found in the file:\n")
	for i, header := range headers {
		sb.WriteString(fmt.Sprintf("  %d: \"%s\"\n", i, header))
	}
	if len(sampleRows) > 0 && len(sampleRows[0]) > 0 {
		sb.WriteString("\nSample data (first row):\n")
		for i, cell := range sampleRows[0] {
			if i < len(headers) {
				sb.WriteString(fmt.Sprintf("  \"%s\": \"%s\"\n", headers[i], cell))
			}
		}
	}
	sb.WriteString("\nReturn a JSON object mapping field names to column header names. Only include columns that have a clear match. Example: {\"name\": \"Product Name\", \"basePrice\": \"Unit Price\"}\n")
	sb.WriteString("Return ONLY the JSON object, no other text.")
	return sb.String()
}
