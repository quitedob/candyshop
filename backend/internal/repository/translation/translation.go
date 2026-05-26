package translation

import (
	"candypro/api/internal/models/common"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TranslationRepository struct {
	db *gorm.DB
}

func New(db *gorm.DB) *TranslationRepository {
	return &TranslationRepository{db: db}
}

func (r *TranslationRepository) FindAll(group, locale, search string, page, limit int) ([]common.Translation, int64, error) {
	var translations []common.Translation
	var total int64

	base := func() *gorm.DB {
		q := r.db.Session(&gorm.Session{}).Model(&common.Translation{}).Where("is_active = ?", true)
		if group != "" {
			q = q.Where("\"group\" = ?", group)
		}
		if locale != "" {
			q = q.Where("locale = ?", locale)
		}
		if search != "" {
			q = q.Where("key ILIKE ? OR value ILIKE ?", "%"+search+"%", "%"+search+"%")
		}
		return q
	}

	if err := base().Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := base().Order("\"group\", key, locale").Offset(offset).Limit(limit).Find(&translations).Error; err != nil {
		return nil, 0, err
	}

	return translations, total, nil
}

func (r *TranslationRepository) FindByID(id uint) (*common.Translation, error) {
	var t common.Translation
	if err := r.db.Session(&gorm.Session{}).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TranslationRepository) Create(t *common.Translation) error {
	return r.db.Create(t).Error
}

func (r *TranslationRepository) Update(t *common.Translation) error {
	return r.db.Save(t).Error
}

func (r *TranslationRepository) Delete(id uint) error {
	return r.db.Delete(&common.Translation{}, id).Error
}

func (r *TranslationRepository) CountByGroup() ([]struct {
	Group string
	Count int64
}, error) {
	var counts []struct {
		Group string
		Count int64
	}
	if err := r.db.Session(&gorm.Session{}).Model(&common.Translation{}).Select("\"group\", count(*) as count").Where("is_active = ?", true).Group("\"group\"").Find(&counts).Error; err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *TranslationRepository) FindAllActive() ([]common.Translation, error) {
	var translations []common.Translation
	if err := r.db.Session(&gorm.Session{}).Where("is_active = ?", true).Find(&translations).Error; err != nil {
		return nil, err
	}
	return translations, nil
}

func (r *TranslationRepository) BulkUpsert(translations []common.Translation) error {
	if len(translations) == 0 {
		return nil
	}
	// R2 E-10: previously this was a per-row FirstOrCreate inside a transaction —
	// 500 imported translations meant 500 round-trips. Replace with a single
	// ON CONFLICT (key, locale) DO UPDATE batch insert. The unique index
	// idx_trans_key_locale on the Translation model is the conflict target.
	// Batch size of 200 keeps PostgreSQL's bind-parameter count safely under
	// the default 65535 limit (~7 columns × 200 rows = 1.4k params).
	//
	// Note: we list the update columns explicitly rather than using
	// AssignmentColumns(["group", ...]) because "group" is a PostgreSQL
	// reserved word and AssignmentColumns doesn't quote identifiers — manual
	// expressions referencing EXCLUDED.* sidestep the quoting issue.
	return r.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}, {Name: "locale"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":      gorm.Expr(`EXCLUDED.value`),
			`"group"`:    gorm.Expr(`EXCLUDED."group"`),
			"is_active":  gorm.Expr(`EXCLUDED.is_active`),
			"updated_by": gorm.Expr(`EXCLUDED.updated_by`),
			"updated_at": gorm.Expr(`EXCLUDED.updated_at`),
		}),
	}).CreateInBatches(translations, 200).Error
}
