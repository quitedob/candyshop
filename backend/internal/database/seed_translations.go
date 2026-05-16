package database

import (
	"log"
	"os"
	"path/filepath"
	"time"

	modelsCommon "candypro/api/internal/models/common"

	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

// SeedTranslations reads YAML locale files and inserts translations into the database.
// Skips if translations already exist (idempotent).
func SeedTranslations(db *gorm.DB, localesDir string) error {
	var count int64
	db.Session(&gorm.Session{}).Model(&modelsCommon.Translation{}).Count(&count)
	if count > 0 {
		return nil
	}

	for _, lang := range []string{"en", "zh", "ko", "ar", "ja", "th", "vi", "id", "ms"} {
		path := filepath.Join(localesDir, lang+".yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: cannot read translation seed file %s: %v", path, err)
			continue
		}

		var raw map[string]map[string]string
		if err := yaml.Unmarshal(data, &raw); err != nil {
			log.Printf("Warning: cannot parse translation seed file %s: %v", path, err)
			continue
		}

		var translations []modelsCommon.Translation
		seen := make(map[string]bool)
		now := time.Now()
		for group, entries := range raw {
			for key, value := range entries {
				dedup := lang + "|" + group + "." + key
				if seen[dedup] {
					log.Printf("Warning: skipping duplicate translation key %s.%s in %s", group, key, lang)
					continue
				}
				seen[dedup] = true
				translations = append(translations, modelsCommon.Translation{
					Key:       key,
					Locale:    lang,
					Value:     value,
					Group:     group,
					IsActive:  true,
					CreatedAt: now,
					UpdatedAt: now,
				})
			}
		}

		if len(translations) > 0 {
			if err := db.Create(&translations).Error; err != nil {
				return err
			}
			log.Printf("Seeded %d %s translations", len(translations), lang)
		}
	}

	return nil
}
