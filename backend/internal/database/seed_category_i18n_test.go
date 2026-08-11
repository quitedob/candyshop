package database

import (
	"testing"

	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupCategoryTranslationsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsProduct.Category{}); err != nil {
		t.Fatalf("AutoMigrate categories: %v", err)
	}
	return db
}

// TestApplyCategoryI18nSeedGuarded_KeepsAdminEdits proves the drop-in guarded
// category merge (applyCategoryI18nSeedGuarded, seed.go) preserves
// admin-authored translations and does not re-map zh onto the scalar Name for an
// already-localized category, while still backfilling genuinely missing keys
// (M3 / G27-c). seed_catalog_i18n.go's seedAllCategoryTranslations should call
// it in place of its unconditional writes so the every-boot category seed stops
// reverting admin edits.
func TestApplyCategoryI18nSeedGuarded_KeepsAdminEdits(t *testing.T) {
	cat := &modelsProduct.Category{
		Slug:  "gummy-candy",
		Name:  "手工软糖",
		Alias: "手工QQ糖",
		Translations: modelsCommon.JSONMap{
			"zh": {"name": "手工软糖", "alias": "手工QQ糖", "description": "管理员撰写的中文描述"},
			"ko": {"name": "관리자 구미", "description": "관리자 설명"},
		},
	}

	applyCategoryI18nSeedGuarded(cat, categoryI18nSeed["gummy-candy"])

	if cat.Name != "手工软糖" {
		t.Fatalf("guarded merge re-mapped scalar Name from zh: got %q want %q", cat.Name, "手工软糖")
	}
	zh := cat.Translations["zh"]
	if zh == nil || zh["name"] != "手工软糖" {
		t.Fatalf("guarded merge reverted admin zh name: got %v", zh)
	}
	if zh["description"] != "管理员撰写的中文描述" {
		t.Fatalf("guarded merge reverted admin zh description: got %v", zh)
	}
	if ko := cat.Translations["ko"]; ko == nil || ko["name"] != "관리자 구미" {
		t.Fatalf("guarded merge reverted admin ko translation: got %v", ko)
	}
	// genuinely missing keys are still backfilled (fill-only merge)
	if en := cat.Translations["en"]; en == nil || en["name"] != "Gummy Candy" {
		t.Fatalf("guarded merge did not backfill missing en keys: got %v", cat.Translations["en"])
	}
}

// TestApplyCategoryI18nSeedGuarded_FreshSeedMapsZh pins that the first seed run
// still maps the zh name onto the scalar Name (zh is the primary display locale)
// and backfills the empty en alias fallback, so the alreadyLocalized gate does not
// regress fresh seeds. Mirrors the original seedAllCategoryTranslations, which only
// maps zh name onto Name (never zh alias onto Alias).
func TestApplyCategoryI18nSeedGuarded_FreshSeedMapsZh(t *testing.T) {
	cat := &modelsProduct.Category{
		Slug:        "hard-candy",
		Name:        "Hard Candy",
		Description: "Classic boiled candies with fruit fillings.",
	}

	applyCategoryI18nSeedGuarded(cat, categoryI18nSeed["hard-candy"])

	if cat.Name != "硬糖" {
		t.Fatalf("fresh guarded merge did not map zh name onto scalar Name: got %q want %q", cat.Name, "硬糖")
	}
	zh := cat.Translations["zh"]
	if zh == nil || zh["name"] != "硬糖" {
		t.Fatalf("fresh guarded merge did not fill zh translation: got %v", zh)
	}
	if cat.Alias != "Boiled Candy" {
		t.Fatalf("fresh guarded merge did not backfill the empty en alias fallback: got %q", cat.Alias)
	}
}

// TestSeedAllCategoryTranslations_KeepsAdminZhName proves the boot-path category
// seed (cmd/api/main.go:85 -> SeedCategoryTranslations -> seedAllCategoryTranslations,
// seed_catalog_i18n.go) no longer reverts an admin-authored zh name after a
// restart (M3 / G27-c). seedAllCategoryTranslations routes every category through
// the guarded fill-only merge applyCategoryI18nSeedGuarded (seed.go), so a
// non-empty scalar Name and existing zh translation survive while genuinely
// missing locales/keys are still backfilled. This test FAILED on the pre-fix path
// (the unguarded merge reverted "手工软糖" to the seed default "软糖"); it now pins
// the admin edit surviving the every-boot seed.
func TestSeedAllCategoryTranslations_KeepsAdminZhName(t *testing.T) {
	db := setupCategoryTranslationsTestDB(t)

	const adminZhName = "手工软糖"
	if err := db.Create(&modelsProduct.Category{
		Slug:  "gummy-candy",
		Name:  adminZhName,
		Alias: "手工QQ糖",
		Translations: modelsCommon.JSONMap{
			"zh": {"name": adminZhName, "alias": "手工QQ糖", "description": "管理员撰写的中文描述"},
		},
	}).Error; err != nil {
		t.Fatalf("insert category: %v", err)
	}

	if err := seedAllCategoryTranslations(db); err != nil {
		t.Fatalf("seedAllCategoryTranslations: %v", err)
	}

	var got modelsProduct.Category
	if err := db.Where("slug = ?", "gummy-candy").First(&got).Error; err != nil {
		t.Fatalf("reload category: %v", err)
	}
	zh := got.Translations["zh"]
	if zh == nil || zh["name"] != adminZhName {
		t.Fatalf("seedAllCategoryTranslations reverted an admin zh name edit: got %v want %q", zh, adminZhName)
	}
	if got.Name != adminZhName {
		t.Fatalf("seedAllCategoryTranslations re-mapped scalar Name to the zh seed default: got %q want %q", got.Name, adminZhName)
	}
	if zh["alias"] != "手工QQ糖" || zh["description"] != "管理员撰写的中文描述" {
		t.Fatalf("seedAllCategoryTranslations reverted admin zh alias/description: got %v", zh)
	}
	// genuinely missing locales/keys are still backfilled (fill-only merge)
	if en := got.Translations["en"]; en == nil || en["name"] != "Gummy Candy" {
		t.Fatalf("seedAllCategoryTranslations did not backfill missing en keys: got %v", got.Translations["en"])
	}
	if ja := got.Translations["ja"]; ja == nil || ja["name"] == "" {
		t.Fatalf("seedAllCategoryTranslations did not backfill missing ja locale: got %v", got.Translations["ja"])
	}
}
