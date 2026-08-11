package user

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupUserModelTestDB migrates the users table in-memory. FK constraints are
// disabled because User's CompanyRef/Role associations reference companies and
// roles tables that are not part of this isolated test.
func setupUserModelTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	return db
}

// TestSoftDeletedEmailCanBeReused pins the G24a fix: the email unique index must
// not span soft-deleted rows, otherwise a deleted account's address can never be
// re-registered (the full-table unique index makes the second INSERT fail with a
// constraint error).
func TestSoftDeletedEmailCanBeReused(t *testing.T) {
	db := setupUserModelTestDB(t)

	first := User{ID: "u1", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("create first user: %v", err)
	}
	if err := db.Delete(&first).Error; err != nil {
		t.Fatalf("soft delete first user: %v", err)
	}

	second := User{ID: "u2", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("re-registering a soft-deleted email must succeed, got: %v", err)
	}

	// Both rows exist: the first is soft-deleted, the second is the live account.
	var live User
	if err := db.Where("email = ?", "reuse@example.com").First(&live).Error; err != nil {
		t.Fatalf("load live user: %v", err)
	}
	if live.ID != "u2" {
		t.Fatalf("live user id = %q, want u2 (soft-deleted row must stay out of default scope)", live.ID)
	}
}

// TestDropLegacyEmailUniqueIndex_UnblocksReuse pins the G24a deployed-DB fix: on
// an already-migrated database the pre-G24a full-table unique index
// (idx_users_email) survives AutoMigrate and blocks re-registering a soft-deleted
// email with a unique-constraint 5xx. DropLegacyEmailUniqueIndex removes exactly
// that index, after which reuse succeeds on the partial index.
func TestDropLegacyEmailUniqueIndex_UnblocksReuse(t *testing.T) {
	db := setupUserModelTestDB(t)

	// Simulate an already-migrated deployment: the legacy full-table unique index
	// is still present alongside the new partial index.
	if err := db.Exec("CREATE UNIQUE INDEX idx_users_email ON users(email)").Error; err != nil {
		t.Fatalf("create legacy index: %v", err)
	}
	if !db.Migrator().HasIndex(&User{}, LegacyUniqueIndexEmail) {
		t.Fatal("legacy index should be present on the migrated DB")
	}

	first := User{ID: "u1", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&first).Error; err != nil {
		t.Fatalf("create first user: %v", err)
	}
	if err := db.Delete(&first).Error; err != nil {
		t.Fatalf("soft delete first user: %v", err)
	}

	second := User{ID: "u2", Email: "reuse@example.com", PasswordHash: "x"}
	if err := db.Create(&second).Error; err == nil {
		t.Fatal("expected legacy full-table unique index to block email reuse (pre-fix production symptom)")
	}

	if err := DropLegacyEmailUniqueIndex(db); err != nil {
		t.Fatalf("drop legacy index: %v", err)
	}
	if db.Migrator().HasIndex(&User{}, LegacyUniqueIndexEmail) {
		t.Fatal("legacy index should be gone after drop")
	}

	if err := db.Create(&second).Error; err != nil {
		t.Fatalf("reuse after dropping the legacy index must succeed: %v", err)
	}
}
