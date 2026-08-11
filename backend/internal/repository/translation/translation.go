package translation

import (
	"context"

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

func (r *TranslationRepository) FindAll(ctx context.Context, group, locale, search string, page, limit int) ([]common.Translation, int64, error) {
	var translations []common.Translation
	var total int64

	base := func() *gorm.DB {
		q := r.db.WithContext(ctx).Model(&common.Translation{}).Where("is_active = ?", true)
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

func (r *TranslationRepository) FindByID(ctx context.Context, id uint) (*common.Translation, error) {
	var t common.Translation
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TranslationRepository) Create(ctx context.Context, t *common.Translation) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *TranslationRepository) Update(ctx context.Context, t *common.Translation) error {
	return r.db.WithContext(ctx).Save(t).Error
}

func (r *TranslationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&common.Translation{}, id).Error
}

func (r *TranslationRepository) CountByGroup(ctx context.Context) ([]struct {
	Group string
	Count int64
}, error) {
	var counts []struct {
		Group string
		Count int64
	}
	if err := r.db.WithContext(ctx).Model(&common.Translation{}).Select("\"group\", count(*) as count").Where("is_active = ?", true).Group("\"group\"").Find(&counts).Error; err != nil {
		return nil, err
	}
	return counts, nil
}

func (r *TranslationRepository) FindAllActive(ctx context.Context) ([]common.Translation, error) {
	var translations []common.Translation
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&translations).Error; err != nil {
		return nil, err
	}
	return translations, nil
}

func (r *TranslationRepository) BulkUpsert(ctx context.Context, translations []common.Translation) error {
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
	// Note: the SET column names are left unquoted so each driver quotes them
	// itself (PostgreSQL: "group", SQLite: `group`). Pre-quoting the key as
	// "group" only worked on PostgreSQL (whose QuoteTo collapses self-quoted
	// identifiers) and generated invalid SQL on SQLite. The EXCLUDED.*
	// expressions are written verbatim, which is accepted by both drivers.
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "key"}, {Name: "locale"}},
		DoUpdates: clause.Assignments(map[string]any{
			"value":      gorm.Expr(`EXCLUDED.value`),
			"group":      gorm.Expr(`EXCLUDED."group"`),
			"is_active":  gorm.Expr(`EXCLUDED.is_active`),
			"updated_by": gorm.Expr(`EXCLUDED.updated_by`),
			"updated_at": gorm.Expr(`EXCLUDED.updated_at`),
		}),
	}).CreateInBatches(translations, 200).Error
}
