package auth

import (
	"context"
	"errors"
	"testing"

	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
	authrepo "candypro/api/internal/repository/auth"
	userRepo "candypro/api/internal/repository/user"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupRoleTestDB seeds an in-memory sqlite DB with the three system roles and
// one user per role, so CanGrantRole can resolve callers from the roles table.
func setupRoleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsAuth.Role{}, &modelsUser.User{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	for _, r := range []modelsAuth.Role{
		{ID: "r-customer", Name: modelsAuth.User, IsSystem: true},
		{ID: "r-admin", Name: modelsAuth.Admin, IsSystem: true},
		{ID: "r-super", Name: modelsAuth.SuperAdmin, IsSystem: true},
	} {
		if err := db.Create(&r).Error; err != nil {
			t.Fatalf("seed role %s: %v", r.Name, err)
		}
	}
	for _, u := range []modelsUser.User{
		{ID: "u-super", Email: "super@example.com", PasswordHash: "x", Status: "active", RoleID: "r-super"},
		{ID: "u-admin", Email: "admin@example.com", PasswordHash: "x", Status: "active", RoleID: "r-admin"},
		{ID: "u-cust", Email: "cust@example.com", PasswordHash: "x", Status: "active", RoleID: "r-customer"},
	} {
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("seed user %s: %v", u.Email, err)
		}
	}
	return db
}

func newTestAuthService(t *testing.T) *AuthService {
	t.Helper()
	db := setupRoleTestDB(t)
	return NewAuthService(userRepo.NewUserRepository(db), authrepo.NewRoleRepository(db), nil, nil, nil)
}

func TestIsPrivilegedRole(t *testing.T) {
	svc := newTestAuthService(t)
	cases := []struct {
		name string
		role *modelsAuth.Role
		want bool
	}{
		{"nil role", nil, false},
		{"customer", &modelsAuth.Role{Name: modelsAuth.User}, false},
		{"admin", &modelsAuth.Role{Name: modelsAuth.Admin}, true},
		{"superadmin", &modelsAuth.Role{Name: modelsAuth.SuperAdmin}, true},
	}
	for _, tc := range cases {
		if got := svc.IsPrivilegedRole(tc.role); got != tc.want {
			t.Fatalf("%s: IsPrivilegedRole=%v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestCanGrantRole_SuperadminCanGrantPrivileged(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()
	if err := svc.CanGrantRole(ctx, "u-super", &modelsAuth.Role{ID: "r-admin", Name: modelsAuth.Admin}); err != nil {
		t.Fatalf("superadmin granting admin: %v", err)
	}
	if err := svc.CanGrantRole(ctx, "u-super", &modelsAuth.Role{ID: "r-super", Name: modelsAuth.SuperAdmin}); err != nil {
		t.Fatalf("superadmin granting superadmin: %v", err)
	}
}

func TestCanGrantRole_AdminCannotGrantPrivileged(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()
	admin := &modelsAuth.Role{ID: "r-admin", Name: modelsAuth.Admin}
	sup := &modelsAuth.Role{ID: "r-super", Name: modelsAuth.SuperAdmin}
	if err := svc.CanGrantRole(ctx, "u-admin", admin); !errors.Is(err, ErrForbiddenRoleGrant) {
		t.Fatalf("admin granting admin: err=%v, want ErrForbiddenRoleGrant", err)
	}
	if err := svc.CanGrantRole(ctx, "u-admin", sup); !errors.Is(err, ErrForbiddenRoleGrant) {
		t.Fatalf("admin granting superadmin: err=%v, want ErrForbiddenRoleGrant", err)
	}
}

func TestCanGrantRole_AdminCanGrantLowerRoles(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()
	cust := &modelsAuth.Role{ID: "r-customer", Name: modelsAuth.User}
	if err := svc.CanGrantRole(ctx, "u-admin", cust); err != nil {
		t.Fatalf("admin granting customer: %v", err)
	}
	if err := svc.CanGrantRole(ctx, "u-admin", nil); err != nil {
		t.Fatalf("nil target role: %v", err)
	}
}

func TestCanGrantRole_MissingOrUnknownCallerDenied(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()
	admin := &modelsAuth.Role{ID: "r-admin", Name: modelsAuth.Admin}
	if err := svc.CanGrantRole(ctx, "", admin); !errors.Is(err, ErrForbiddenRoleGrant) {
		t.Fatalf("empty caller: err=%v, want ErrForbiddenRoleGrant", err)
	}
	if err := svc.CanGrantRole(ctx, "u-nobody", admin); !errors.Is(err, ErrForbiddenRoleGrant) {
		t.Fatalf("unknown caller: err=%v, want ErrForbiddenRoleGrant", err)
	}
}

func TestResolveRole_ByNameAndByID(t *testing.T) {
	svc := newTestAuthService(t)
	ctx := context.Background()

	byName, err := svc.ResolveRole(ctx, "", modelsAuth.SuperAdmin)
	if err != nil {
		t.Fatalf("resolve by name: %v", err)
	}
	if byName.Name != modelsAuth.SuperAdmin {
		t.Fatalf("name=%q, want %q", byName.Name, modelsAuth.SuperAdmin)
	}

	byID, err := svc.ResolveRole(ctx, "r-admin", "")
	if err != nil {
		t.Fatalf("resolve by id: %v", err)
	}
	if byID.Name != modelsAuth.Admin {
		t.Fatalf("name=%q, want %q", byID.Name, modelsAuth.Admin)
	}

	if _, err := svc.ResolveRole(ctx, "", "does-not-exist"); err == nil {
		t.Fatal("expected error for unknown role name")
	}
	if _, err := svc.ResolveRole(ctx, "", ""); err == nil {
		t.Fatal("expected error when both roleId and roleName are empty")
	}
}
