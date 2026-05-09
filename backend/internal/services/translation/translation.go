package translation

import (
	"candypro/api/internal/i18n"
	"candypro/api/internal/models/common"
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
	s.refreshCache()
	return nil
}

func (s *TranslationService) Update(t *common.Translation) error {
	if err := s.repo.Update(t); err != nil {
		return err
	}
	s.refreshCache()
	return nil
}

func (s *TranslationService) Delete(id uint) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	s.refreshCache()
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
	s.refreshCache()
	return nil
}

func (s *TranslationService) Export() ([]common.Translation, error) {
	return s.repo.FindAllActive()
}

func (s *TranslationService) refreshCache() {
	records, err := s.repo.FindAllActive()
	if err != nil {
		return
	}
	entries := make([]struct{ Locale, Key, Value string }, len(records))
	for i, r := range records {
		entries[i].Locale = r.Locale
		entries[i].Key = "errors." + r.Key
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
