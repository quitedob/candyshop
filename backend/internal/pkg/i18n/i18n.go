package i18n

import (
	"strings"
	"sync"

	modelsCommon "candypro/api/internal/models/common"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	mu      sync.RWMutex
	cache   map[string]string // "locale|group.key" → value
	locales = []string{"en", "zh", "ko", "ar", "ja", "th", "vi", "id", "ms"}
)

// Init loads all active translations from the database into the in-memory cache.
func Init(db *gorm.DB) error {
	var records []modelsCommon.Translation
	if err := db.Session(&gorm.Session{}).Where("is_active = ?", true).Find(&records).Error; err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	cache = make(map[string]string, len(records))
	localeSet := map[string]bool{"en": true, "zh": true, "ko": true, "ar": true, "ja": true, "th": true, "vi": true, "id": true, "ms": true}
	for _, r := range records {
		key := r.Group + "." + r.Key
		cache[r.Locale+"|"+key] = r.Value
		localeSet[r.Locale] = true
	}

	locales = make([]string, 0, len(localeSet))
	for l := range localeSet {
		locales = append(locales, l)
	}

	return nil
}

// Translate returns the translated string for the given locale and key.
// Lookup order: cache → en fallback → key itself.
func Translate(locale, key string) string {
	cacheKey := locale + "|" + key

	mu.RLock()
	if v, ok := cache[cacheKey]; ok {
		mu.RUnlock()
		return v
	}
	mu.RUnlock()

	if locale != "en" {
		return Translate("en", key)
	}
	return key
}

// TranslateWithVars substitutes {{.varName}} placeholders in the translated string.
func TranslateWithVars(locale, key string, vars map[string]string) string {
	s := Translate(locale, key)
	for varName, varValue := range vars {
		s = strings.ReplaceAll(s, "{{."+varName+"}}", varValue)
	}
	return s
}

// WarmCache merges records into the in-memory cache (called after DB changes).
func WarmCache(records []struct{ Locale, Key, Value string }) {
	mu.Lock()
	defer mu.Unlock()
	for _, r := range records {
		cache[r.Locale+"|"+r.Key] = r.Value
	}
}

// ClearCache clears the in-memory cache.
func ClearCache() {
	mu.Lock()
	defer mu.Unlock()
	cache = make(map[string]string)
}

// DefaultLocale returns the default locale code.
func DefaultLocale() string {
	return "zh"
}

// SupportedLocales returns the list of supported locale codes.
func SupportedLocales() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, len(locales))
	copy(out, locales)
	return out
}

// T returns the translated string using the locale stored in gin.Context.
func T(c *gin.Context, key string) string {
	locale, _ := c.Get("locale")
	if loc, ok := locale.(string); ok && loc != "" {
		return Translate(loc, key)
	}
	return Translate(DefaultLocale(), key)
}

// TWithVars translates with variable substitution using the locale in gin.Context.
func TWithVars(c *gin.Context, key string, vars map[string]string) string {
	locale, _ := c.Get("locale")
	if loc, ok := locale.(string); ok && loc != "" {
		return TranslateWithVars(loc, key, vars)
	}
	return TranslateWithVars(DefaultLocale(), key, vars)
}
