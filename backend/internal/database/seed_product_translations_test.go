package database

import (
	"testing"

	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupProductTranslationsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsProduct.Product{}); err != nil {
		t.Fatalf("AutoMigrate products: %v", err)
	}
	return db
}

// TestSeedProductTranslations_KeepsAdminEdits proves re-seeding does not clobber
// admin-edited product translations or scalar columns (M4 / G27-c). Before the
// fix, seedProductTranslations unconditionally overwrote zh translation values
// and re-mapped them onto Name/Summary via syncProductScalarsFromZh, reverting
// the admin's edits on every seed run.
func TestSeedProductTranslations_KeepsAdminEdits(t *testing.T) {
	db := setupProductTranslationsTestDB(t)

	adminName := "管理员定制软糖"
	adminZhSummary := "管理员定制的摘要"
	if err := db.Create(&modelsProduct.Product{
		ID:       "p-admin-edited",
		Slug:     "4d-fruit-gummy",
		Name:     adminName,
		Summary:  adminZhSummary,
		Category: "软糖",
		Translations: modelsCommon.JSONMap{
			"zh": {"name": adminName, "summary": adminZhSummary},
			"en": {"name": "Admin Brand Gummy"},
		},
	}).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := seedProductTranslations(db); err != nil {
		t.Fatalf("seedProductTranslations: %v", err)
	}

	var got modelsProduct.Product
	if err := db.Where("slug = ?", "4d-fruit-gummy").First(&got).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}

	if got.Name != adminName {
		t.Fatalf("admin-edited Name scalar overwritten by seed: got %q want %q", got.Name, adminName)
	}
	zh := got.Translations["zh"]
	if zh == nil || zh["name"] != adminName {
		t.Fatalf("admin-edited zh name translation overwritten: got %v want %q", zh, adminName)
	}
	if zh == nil || zh["summary"] != adminZhSummary {
		t.Fatalf("admin-edited zh summary translation overwritten: got %v want %q", zh, adminZhSummary)
	}
}

// TestSeedProductTranslations_FreshSeedFillsLocalizedFields pins that the first
// seed run still fills the zh translations and maps them onto the scalar columns
// (zh is the primary display locale), so the guard does not regress fresh seeds.
func TestSeedProductTranslations_FreshSeedFillsLocalizedFields(t *testing.T) {
	db := setupProductTranslationsTestDB(t)

	if err := db.Create(&modelsProduct.Product{
		ID:      "p-fresh",
		Slug:    "4d-fruit-gummy",
		Name:    "4D Fruit Gummy",
		Summary: "Real fruit juice gummy",
		Flavors: modelsCommon.StringArray{"Strawberry", "Orange"},
	}).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := seedProductTranslations(db); err != nil {
		t.Fatalf("seedProductTranslations: %v", err)
	}

	var got modelsProduct.Product
	if err := db.Where("slug = ?", "4d-fruit-gummy").First(&got).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}

	const wantZhName = "四维水果软糖"
	if got.Name != wantZhName {
		t.Fatalf("fresh seed did not map zh name onto scalar: got %q want %q", got.Name, wantZhName)
	}
	zh := got.Translations["zh"]
	if zh == nil || zh["name"] != wantZhName {
		t.Fatalf("fresh seed did not fill zh translation: got %v want name=%q", zh, wantZhName)
	}
}

// TestSeedProductTranslationsWrapper_KeepsDivergentScalar drives the exported
// SeedProductTranslations wrapper — the function cmd/api/main.go actually calls
// on the AUTO_SEED_DATA=true startup path — not just the unexported inner helper.
// A scalar Name that diverges from zh.name (e.g. written by the inventory XLSX
// importer UpdateProductColumns without touching translations) must survive a
// re-seed. Pre-fix, the wrapper routed through seedProductMissingLocales, which
// unconditionally re-mapped zh -> scalar and reverted the admin's Name.
func TestSeedProductTranslationsWrapper_KeepsDivergentScalar(t *testing.T) {
	db := setupProductTranslationsTestDB(t)

	divergentName := "管理员定制软糖"
	if err := db.Create(&modelsProduct.Product{
		ID:       "p-divergent",
		Slug:     "4d-fruit-gummy",
		Name:     divergentName,
		Category: "软糖",
		Translations: modelsCommon.JSONMap{
			"zh": {"name": "四维水果软糖", "summary": "真果汁软糖"},
			"en": {"name": "4D Fruit Gummy"},
		},
	}).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := SeedProductTranslations(db); err != nil {
		t.Fatalf("SeedProductTranslations: %v", err)
	}

	var got modelsProduct.Product
	if err := db.Where("slug = ?", "4d-fruit-gummy").First(&got).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if got.Name != divergentName {
		t.Fatalf("wrapper reverted divergent scalar Name: got %q want %q", got.Name, divergentName)
	}
	zh := got.Translations["zh"]
	if zh == nil || zh["name"] == "" {
		t.Fatalf("wrapper dropped zh translation: got %v", zh)
	}
}

// TestSeedProductTranslationsWrapper_FreshSeedLocalizes pins that the very first
// AUTO_SEED_DATA run still localizes a fresh product through the exported wrapper
// (fills zh, maps the zh name onto the scalar column), so the guard does not
// regress fresh seeds.
func TestSeedProductTranslationsWrapper_FreshSeedLocalizes(t *testing.T) {
	db := setupProductTranslationsTestDB(t)

	if err := db.Create(&modelsProduct.Product{
		ID:      "p-fresh",
		Slug:    "4d-fruit-gummy",
		Name:    "4D Fruit Gummy",
		Summary: "Real fruit juice gummy",
	}).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := SeedProductTranslations(db); err != nil {
		t.Fatalf("SeedProductTranslations: %v", err)
	}

	var got modelsProduct.Product
	if err := db.Where("slug = ?", "4d-fruit-gummy").First(&got).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	const wantZhName = "四维水果软糖"
	if got.Name != wantZhName {
		t.Fatalf("fresh wrapper seed did not map zh name onto scalar: got %q want %q", got.Name, wantZhName)
	}
	zh := got.Translations["zh"]
	if zh == nil || zh["name"] != wantZhName {
		t.Fatalf("fresh wrapper seed did not fill zh translation: got %v", zh)
	}
}

// TestSeedProductTranslationsWrapper_NoZhDataLossWhenNamesEqual pins the data-loss
// edge raised in review: if an admin sets zh.name == en.name (e.g. a brand name
// identical in both languages), the whole zh map — summary, description, etc. —
// must survive re-seeding. Pre-fix, the wrapper routed through
// seedProductMissingLocales -> repairProductZhCopiedFromEN, which deleted the zh
// map whenever en.name == zh.name.
func TestSeedProductTranslationsWrapper_NoZhDataLossWhenNamesEqual(t *testing.T) {
	db := setupProductTranslationsTestDB(t)

	const sharedName = "CandyPro Classic"
	if err := db.Create(&modelsProduct.Product{
		ID:      "p-equal",
		Slug:    "candypro-classic",
		Name:    sharedName,
		Summary: "En summary",
		Translations: modelsCommon.JSONMap{
			"zh": {"name": sharedName, "summary": "中文摘要", "description": "中文描述"},
			"en": {"name": sharedName, "summary": "En summary"},
		},
	}).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := SeedProductTranslations(db); err != nil {
		t.Fatalf("SeedProductTranslations: %v", err)
	}

	var got modelsProduct.Product
	if err := db.Where("slug = ?", "candypro-classic").First(&got).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	zh := got.Translations["zh"]
	if zh == nil || zh["summary"] != "中文摘要" {
		t.Fatalf("zh translation map was deleted or clobbered by re-seed: got %v", got.Translations["zh"])
	}
}

// TestSeedProductMissingLocalesLegacyStillRevertsScalar is a characterization
// test for the legacy unguarded seedProductMissingLocales (seed_catalog_i18n.go)
// that SeedProductTranslations used to route through: it unconditionally re-maps
// zh onto the scalar columns, reverting admin edits. The exported wrapper must
// NOT call it (see TestSeedProductTranslationsWrapper_KeepsDivergentScalar).
// Keeping this reference also keeps the legacy function reachable so the owner of
// seed_catalog_i18n.go can delete or fix it without a lint regression.
func TestSeedProductMissingLocalesLegacyStillRevertsScalar(t *testing.T) {
	db := setupProductTranslationsTestDB(t)

	if err := db.Create(&modelsProduct.Product{
		ID:       "p-legacy",
		Slug:     "divergent-sku",
		Name:     "管理员定制软糖",
		Category: "软糖",
		Translations: modelsCommon.JSONMap{
			"zh": {"name": "四维水果软糖", "summary": "真果汁软糖"},
			"en": {"name": "4D Fruit Gummy"},
		},
	}).Error; err != nil {
		t.Fatalf("insert product: %v", err)
	}

	if err := seedProductMissingLocales(db); err != nil {
		t.Fatalf("legacy seedProductMissingLocales: %v", err)
	}

	var got modelsProduct.Product
	if err := db.Where("slug = ?", "divergent-sku").First(&got).Error; err != nil {
		t.Fatalf("reload product: %v", err)
	}
	if got.Name != "四维水果软糖" {
		t.Fatalf("legacy seedProductMissingLocales should re-map zh->scalar (documenting why the wrapper avoids it): got %q", got.Name)
	}
}
