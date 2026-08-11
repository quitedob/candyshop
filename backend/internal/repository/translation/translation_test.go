package translation

import (
	"context"
	"testing"

	modelsCommon "candypro/api/internal/models/common"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTranslationTestDB creates an in-memory SQLite DB with the translations
// table. Columns mirror the modelsCommon.Translation fields the repository
// reads/writes. "group" is a reserved word in SQL and must be quoted.
func setupTranslationTestDB(t *testing.T) *gorm.DB {
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
		deleted_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create translations: %v", err)
	}
	return db
}

func seedTranslation(t *testing.T, r *TranslationRepository, key string, id uint) {
	t.Helper()
	tr := &modelsCommon.Translation{
		Key:      key,
		Locale:   "en",
		Value:    "value-" + key,
		Group:    "test",
		IsActive: true,
	}
	if err := r.Create(context.Background(), tr); err != nil {
		t.Fatalf("seed %s: %v", key, err)
	}
	if tr.ID != id {
		t.Fatalf("seed %s: expected id %d, got %d", key, id, tr.ID)
	}
}

// TestDeleteDoesNotLeakWhereClause is the H13 regression test: a Delete must
// not leak its WHERE id=? clause into a subsequent Delete/Update on the same
// repository instance. Delete(1) followed by Delete(2) must actually delete
// row 2, and an Update after a Delete must persist.
func TestDeleteDoesNotLeakWhereClause(t *testing.T) {
	db := setupTranslationTestDB(t)
	r := New(db)
	ctx := context.Background()

	seedTranslation(t, r, "greeting", 1)
	seedTranslation(t, r, "farewell", 2)
	seedTranslation(t, r, "thanks", 3)

	// Delete row 1, then row 2 on the same repo instance.
	if err := r.Delete(ctx, 1); err != nil {
		t.Fatalf("Delete(1): %v", err)
	}
	if err := r.Delete(ctx, 2); err != nil {
		t.Fatalf("Delete(2): %v", err)
	}

	// Row 2 must now be soft-deleted (FindByID excludes it).
	if _, err := r.FindByID(ctx, 2); err != gorm.ErrRecordNotFound {
		t.Fatalf("after Delete(1)+Delete(2), row 2 should be gone, got err=%v", err)
	}
	// Row 3 must remain untouched.
	if tr, err := r.FindByID(ctx, 3); err != nil || tr.Key != "thanks" {
		t.Fatalf("row 3 should survive, got tr=%+v err=%v", tr, err)
	}

	// An Update after a Delete must persist.
	row3, err := r.FindByID(ctx, 3)
	if err != nil {
		t.Fatalf("reload row 3: %v", err)
	}
	row3.Value = "updated-value"
	if err := r.Update(ctx, row3); err != nil {
		t.Fatalf("Update after delete: %v", err)
	}
	row3b, err := r.FindByID(ctx, 3)
	if err != nil {
		t.Fatalf("reload row 3 after update: %v", err)
	}
	if row3b.Value != "updated-value" {
		t.Fatalf("expected updated value, got %q", row3b.Value)
	}
}
