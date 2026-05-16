package trade

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// BuildColumnMapping uses AI to map XLSX column headers to CandyPro product fields.
// Returns a map of field name → column header name, or nil on failure.
func (s *AIService) BuildColumnMapping(ctx context.Context, headers []string, sampleRows [][]string) map[string]string {
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

	raw, err := s.GenerateJSON(ctx, sb.String())
	if err != nil || raw == "" {
		return nil
	}

	var aiMapping map[string]string
	if json.Unmarshal([]byte(raw), &aiMapping) != nil {
		return nil
	}
	return aiMapping
}
