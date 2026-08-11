package translation

import (
	"context"
	"testing"

	modelsCommon "candypro/api/internal/models/common"
	"candypro/api/internal/pkg/i18n"
	translationrepo "candypro/api/internal/repository/translation"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupServiceTestDB creates an in-memory SQLite DB with the translations table.
// It mirrors the repository test schema and adds the UNIQUE(key, locale)
// constraint so the bulk upsert (ON CONFLICT) used by Import resolves.
func setupServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.Exec(`CREATE TABLE translations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		locale TEXT NOT NULL,
		value TEXT,
		"group" TEXT,
		is_active INTEGER DEFAULT 1,
		updated_by TEXT,
		created_at DATETIME,
		updated_at DATETIME,
		deleted_at DATETIME,
		UNIQUE(key, locale)
	)`).Error; err != nil {
		t.Fatalf("create translations: %v", err)
	}
	return db
}

// TestDeleteInvalidatesCache is the G17 regression test: deleting a translation
// must remove it from the in-memory i18n cache so it stops resolving (falling
// back to the default locale / raw key) without a process restart. Pre-fix, the
// delete path wrote an empty value into the cache, so Translate() returned ""
// and skipped both fallbacks.
func TestDeleteInvalidatesCache(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := translationrepo.New(db)
	svc := NewService(repo)
	ctx := context.Background()
	t.Cleanup(i18n.ClearCache)

	// Insert the row directly through the repo so only the Delete path touches
	// the cache invalidation hook.
	tr := &modelsCommon.Translation{
		Key: "greeting", Locale: "en", Value: "Hello", Group: "test", IsActive: true,
	}
	if err := repo.Create(ctx, tr); err != nil {
		t.Fatalf("seed: %v", err)
	}

	// Load the cache as the server does at startup.
	if err := i18n.Init(db); err != nil {
		t.Fatalf("i18n.Init: %v", err)
	}
	if got := i18n.Translate("en", "test.greeting"); got != "Hello" {
		t.Fatalf("sanity: expected 'Hello', got %q", got)
	}

	if err := svc.Delete(ctx, tr.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Pre-fix the key resolved to "" (blank); post-fix it must fall back to the
	// raw key (no zh entry exists for it).
	if got := i18n.Translate("en", "test.greeting"); got != "test.greeting" {
		t.Fatalf("deleted key must stop resolving; got %q", got)
	}
}

// TestImportRebuildsCache is the G17 regression test for the batch path: an
// Import must rebuild (replace) the in-memory cache so keys that are no longer
// active stop resolving. Pre-fix, refreshCache only merged via WarmCache,
// leaving a deactivated key resolving with its stale value.
//
// The row is deactivated out-of-band (directly in the DB, bypassing the cache
// hooks) to simulate the cache going stale before a batch import. Note that
// deactivating via a Create batch would not work: GORM coerces an explicit
// IsActive:false to true because of the `default:true` tag on the field.
func TestImportRebuildsCache(t *testing.T) {
	db := setupServiceTestDB(t)
	repo := translationrepo.New(db)
	svc := NewService(repo)
	ctx := context.Background()
	t.Cleanup(i18n.ClearCache)

	// Seed an active key and load the cache as the server does at startup.
	greeting := &modelsCommon.Translation{
		Key: "greeting", Locale: "en", Value: "Hello", Group: "test", IsActive: true,
	}
	if err := repo.Create(ctx, greeting); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := i18n.Init(db); err != nil {
		t.Fatalf("i18n.Init: %v", err)
	}
	if got := i18n.Translate("en", "test.greeting"); got != "Hello" {
		t.Fatalf("sanity: expected 'Hello', got %q", got)
	}

	// Simulate the cache going stale: deactivate the row directly in the DB
	// without touching the in-memory cache (as an out-of-band process would).
	row, err := repo.FindByID(ctx, greeting.ID)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	row.IsActive = false
	if err := repo.Update(ctx, row); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	// Batch import adds a brand-new key. The full reload triggered by Import
	// must drop the deactivated key from the cache.
	batch := []modelsCommon.Translation{
		{Key: "welcome", Locale: "en", Value: "Welcome", Group: "test", IsActive: true},
	}
	if err := svc.Import(ctx, batch); err != nil {
		t.Fatalf("Import: %v", err)
	}

	// The deactivated key must stop resolving (raw-key fallback), not keep the
	// stale value.
	if got := i18n.Translate("en", "test.greeting"); got != "test.greeting" {
		t.Fatalf("deactivated key must stop resolving; got %q", got)
	}
	// The newly imported key must resolve.
	if got := i18n.Translate("en", "test.welcome"); got != "Welcome" {
		t.Fatalf("imported key must resolve; got %q", got)
	}
}
