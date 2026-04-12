package common

import "strings"

// escapeLikePattern escapes special characters for SQL LIKE operator
func escapeLikePattern(pattern string) string {
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"%", "\\%",
		"_", "\\_",
	)
	return replacer.Replace(pattern)
}
