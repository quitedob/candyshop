package middleware

import (
	"strings"

	"candypro/api/internal/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// LocaleMiddleware extracts the preferred locale and stores it in gin.Context.
// Priority: ?lang= query param > Accept-Language header > default "zh".
func LocaleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := ""

		// 1. Query parameter takes highest priority
		if q := c.Query("lang"); q != "" {
			locale = normalizeLocale(q)
		}

		// 2. Accept-Language header
		if locale == "" {
			locale = parseAcceptLanguage(c.GetHeader("Accept-Language"))
		}

		// 3. Default
		if locale == "" {
			locale = i18n.DefaultLocale()
		}

		c.Set("locale", locale)
		c.Next()
	}
}

// normalizeLocale returns a clean locale code or "" if invalid.
func normalizeLocale(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "zh", "zh-cn", "zh-tw", "zh-hans", "zh-hant", "cn":
		return "zh"
	case "en", "en-us", "en-gb", "us", "gb":
		return "en"
	case "ko", "ko-kr":
		return "ko"
	case "ar", "ar-sa", "ar-ae", "ar-eg":
		return "ar"
	case "ja", "ja-jp":
		return "ja"
	case "th", "th-th":
		return "th"
	case "vi", "vi-vn":
		return "vi"
	case "id", "id-id":
		return "id"
	case "ms", "ms-my":
		return "ms"
	default:
		if len(s) >= 2 {
			candidate := s[:2]
			for _, supported := range i18n.SupportedLocales() {
				if candidate == supported {
					return candidate
				}
			}
		}
		return ""
	}
}

// parseAcceptLanguage parses the Accept-Language header and returns the best match.
func parseAcceptLanguage(header string) string {
	if header == "" {
		return ""
	}

	// Simple parsing: take the first language tag, ignoring quality values
	parts := strings.Split(header, ",")
	for _, part := range parts {
		// Strip quality value (;q=...)
		tag := strings.TrimSpace(strings.Split(part, ";")[0])
		if loc := normalizeLocale(tag); loc != "" {
			return loc
		}
	}
	return ""
}
