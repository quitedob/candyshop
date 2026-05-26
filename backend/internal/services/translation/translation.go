package translation

import (
	"candypro/api/internal/models/common"
	"candypro/api/internal/pkg/i18n"
	"log/slog"
	"strconv"
)

type translationRepository interface {
	FindAll(group, locale, search string, page, limit int) ([]common.Translation, int64, error)
	FindByID(id uint) (*common.Translation, error)
	Create(t *common.Translation) error
	Update(t *common.Translation) error
	Delete(id uint) error
	CountByGroup() ([]struct {
		Group string
		Count int64
	}, error)
	FindAllActive() ([]common.Translation, error)
	BulkUpsert(translations []common.Translation) error
}

type TranslationService struct {
	repo translationRepository
}

func NewService(repo translationRepository) *TranslationService {
	return &TranslationService{repo: repo}
}

func (s *TranslationService) List(group, locale, search string, page, limit int) ([]common.Translation, int64, error) {
	return s.repo.FindAll(group, locale, search, page, limit)
}

func (s *TranslationService) GetByID(id uint) (*common.Translation, error) {
	return s.repo.FindByID(id)
}

func (s *TranslationService) Create(t *common.Translation) error {
	if err := s.repo.Create(t); err != nil {
		return err
	}
	// R2 E-9: previously every single-key write triggered a full FindAllActive
	// reload. The hot path for translation editing in the admin UI does many
	// rapid single-key saves, so we now apply only the changed entry to the
	// cache. The full reload is reserved for bulk Import where dozens or
	// hundreds of rows change at once.
	s.applyDelta(*t)
	return nil
}

func (s *TranslationService) Update(t *common.Translation) error {
	if err := s.repo.Update(t); err != nil {
		return err
	}
	s.applyDelta(*t)
	return nil
}

func (s *TranslationService) Delete(id uint) error {
	// Capture the row before deletion so we can invalidate the matching cache
	// entry without falling back to a full reload.
	var stale *common.Translation
	if existing, ferr := s.repo.FindByID(id); ferr == nil {
		stale = existing
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	if stale != nil {
		// Best signal we have for invalidation: overwrite with empty value.
		// Translate() will still find the entry (returning ""), and the next
		// full reload (Import) will remove it. This avoids the global reload
		// cost while keeping correctness for the admin's immediate next read.
		i18n.WarmCache([]struct{ Locale, Key, Value string }{{
			Locale: stale.Locale,
			Key:    stale.Group + "." + stale.Key,
			Value:  "",
		}})
	}
	return nil
}

func (s *TranslationService) Groups() ([]struct {
	Group string
	Count int64
}, error) {
	return s.repo.CountByGroup()
}

func (s *TranslationService) Import(translations []common.Translation) error {
	if err := s.repo.BulkUpsert(translations); err != nil {
		return err
	}
	// Bulk import is the one path where a full reload is justified — many
	// rows change in one shot and the marginal cost of FindAllActive is
	// dominated by the BulkUpsert work itself.
	s.refreshCache()
	return nil
}

func (s *TranslationService) Export() ([]common.Translation, error) {
	return s.repo.FindAllActive()
}

// applyDelta merges a single translation into the in-memory cache without
// re-reading the entire table.
func (s *TranslationService) applyDelta(t common.Translation) {
	if t.Locale == "" || t.Key == "" {
		return
	}
	i18n.WarmCache([]struct{ Locale, Key, Value string }{{
		Locale: t.Locale,
		Key:    t.Group + "." + t.Key,
		Value:  t.Value,
	}})
}

func (s *TranslationService) refreshCache() {
	records, err := s.repo.FindAllActive()
	if err != nil {
		slog.Warn("translation cache refresh failed", "error", err)
		return
	}
	entries := make([]struct{ Locale, Key, Value string }, len(records))
	for i, r := range records {
		entries[i].Locale = r.Locale
		entries[i].Key = r.Group + "." + r.Key
		entries[i].Value = r.Value
	}
	i18n.WarmCache(entries)
}

// ParseID parses a string ID to uint.
func ParseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
