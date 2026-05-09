package i18n

import (
	"embed"
	"fmt"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	// embedFS holds locale YAML files compiled into the binary.
	embedFS embed.FS

	// mu protects the overrides map.
	mu sync.RWMutex

	// overrides stores DB-managed translations keyed by "locale|key".
	overrides map[string]string

	// locales holds the embedded baseline translations keyed by locale → group → key → value.
	locales map[string]map[string]string
)

func init() {
	overrides = make(map[string]string)
	locales = make(map[string]map[string]string)
}

// LoadEmbedded reads all YAML files from the embedded FS into memory.
func LoadEmbedded(fs embed.FS) error {
	for _, lang := range []string{"en", "zh"} {
		data, err := fs.ReadFile("locales/" + lang + ".yaml")
		if err != nil {
			return fmt.Errorf("i18n: read locale %s: %w", lang, err)
		}

		var raw map[string]map[string]string
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("i18n: parse locale %s: %w", lang, err)
		}

		flattened := make(map[string]string)
		for group, entries := range raw {
			for key, val := range entries {
				flattened[group+"."+key] = val
			}
		}
		locales[lang] = flattened
	}
	return nil
}

// Translate returns the translated string for the given locale and key.
// Lookup order: DB override → embedded baseline for locale → embedded baseline for "en" → key itself.
// The key format is "group.code", e.g. "errors.not_found".
func Translate(locale, key string) string {
	cacheKey := locale + "|" + key

	// Check DB override cache
	mu.RLock()
	if v, ok := overrides[cacheKey]; ok {
		mu.RUnlock()
		return v
	}
	mu.RUnlock()

	// Check embedded baseline for requested locale
	if m, ok := locales[locale]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}

	// Fallback to English
	if m, ok := locales["en"]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}

	return key
}

// TranslateWithVars translates and substitutes {{.varName}} placeholders.
func TranslateWithVars(locale, key string, vars map[string]string) string {
	s := Translate(locale, key)
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{{."+k+"}}", v)
	}
	return s
}

// SupportedLocales returns the list of loaded embedded locales.
func SupportedLocales() []string {
	locs := make([]string, 0, len(locales))
	for k := range locales {
		locs = append(locs, k)
	}
	return locs
}

// WarmCache loads all active DB translations into the in-memory cache.
func WarmCache(translationRecords []struct{ Locale, Key, Value string }) {
	mu.Lock()
	defer mu.Unlock()
	overrides = make(map[string]string, len(translationRecords))
	for _, r := range translationRecords {
		overrides[r.Locale+"|"+r.Key] = r.Value
	}
}

// ClearOverrides removes all cached DB overrides.
func ClearOverrides() {
	mu.Lock()
	defer mu.Unlock()
	overrides = make(map[string]string)
}

// DefaultLocale returns the default locale code.
func DefaultLocale() string {
	return "zh"
}
