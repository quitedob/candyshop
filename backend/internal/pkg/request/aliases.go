package request

import "strings"

// FirstNonEmpty returns the first non-empty trimmed string.
func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}

// FirstOptionalNonEmpty returns the first non-nil pointer whose trimmed value is non-empty.
func FirstOptionalNonEmpty(values ...*string) *string {
	for _, v := range values {
		if v == nil {
			continue
		}
		if t := strings.TrimSpace(*v); t != "" {
			return &t
		}
	}
	return nil
}
