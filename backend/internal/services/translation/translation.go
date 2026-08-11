package translation

import (
	"context"

	"candypro/api/internal/models/common"
	"candypro/api/internal/pkg/i18n"
	"log/slog"
	"strconv"
)

type translationRepository interface {
	FindAll(ctx context.Context, group, locale, search string, page, limit int) ([]common.Translation, int64, error)
	FindByID(ctx context.Context, id uint) (*common.Translation, error)
	Create(ctx context.Context, t *common.Translation) error
	Update(ctx context.Context, t *common.Translation) error
	Delete(ctx context.Context, id uint) error
	CountByGroup(ctx context.Context) ([]struct {
		Group string
		Count int64
	}, error)
	FindAllActive(ctx context.Context) ([]common.Translation, error)
	BulkUpsert(ctx context.Context, translations []common.Translation) error
}

type TranslationService struct {
	repo translationRepository
}

func NewService(repo translationRepository) *TranslationService {
	return &TranslationService{repo: repo}
}

func (s *TranslationService) List(ctx context.Context, group, locale, search string, page, limit int) ([]common.Translation, int64, error) {
	return s.repo.FindAll(ctx, group, locale, search, page, limit)
}

func (s *TranslationService) GetByID(ctx context.Context, id uint) (*common.Translation, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *TranslationService) Create(ctx context.Context, t *common.Translation) error {
	if err := s.repo.Create(ctx, t); err != nil {
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

func (s *TranslationService) Update(ctx context.Context, t *common.Translation) error {
	if err := s.repo.Update(ctx, t); err != nil {
		return err
	}
	s.applyDelta(*t)
	return nil
}

func (s *TranslationService) Delete(ctx context.Context, id uint) error {
	// Capture the row before deletion so we can invalidate the matching cache
	// entry without falling back to a full reload.
	var stale *common.Translation
	if existing, ferr := s.repo.FindByID(ctx, id); ferr == nil {
		stale = existing
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if stale != nil {
		// Remove the key from the in-memory cache entirely. Leaving a blank
		// entry behind would make Translate() return "" and skip both the
		// default-locale and raw-key fallbacks.
		i18n.DeleteFromCache(stale.Locale, stale.Group+"."+stale.Key)
	}
	return nil
}

func (s *TranslationService) Groups(ctx context.Context) ([]struct {
	Group string
	Count int64
}, error) {
	return s.repo.CountByGroup(ctx)
}

func (s *TranslationService) Import(ctx context.Context, translations []common.Translation) error {
	if err := s.repo.BulkUpsert(ctx, translations); err != nil {
		return err
	}
	// Bulk import is the one path where a full reload is justified — many
	// rows change in one shot and the marginal cost of FindAllActive is
	// dominated by the BulkUpsert work itself.
	s.refreshCache(ctx)
	return nil
}

func (s *TranslationService) Export(ctx context.Context) ([]common.Translation, error) {
	return s.repo.FindAllActive(ctx)
}

// applyDelta merges a single translation into the in-memory cache without
// re-reading the entire table. A deactivated translation behaves like a delete:
// it is removed from the cache so it stops resolving.
func (s *TranslationService) applyDelta(t common.Translation) {
	if t.Locale == "" || t.Key == "" {
		return
	}
	if !t.IsActive {
		i18n.DeleteFromCache(t.Locale, t.Group+"."+t.Key)
		return
	}
	i18n.WarmCache([]struct{ Locale, Key, Value string }{{
		Locale: t.Locale,
		Key:    t.Group + "." + t.Key,
		Value:  t.Value,
	}})
}

func (s *TranslationService) refreshCache(ctx context.Context) {
	records, err := s.repo.FindAllActive(ctx)
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
	// Full replace, not merge: a batch import may have deactivated or removed
	// keys that must stop resolving instead of lingering as stale entries.
	i18n.ReplaceCache(entries)
}

// ParseID parses a string ID to uint.
func ParseID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
