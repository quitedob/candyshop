package user

import (
	"context"
	"testing"

	modelsUser "candypro/api/internal/models/user"
)

// stubUserRepo embeds the userRepository interface so only the methods under
// test need to be provided. It records the role filter passed to
// FindByRoleNames so the staff-list scoping can be asserted without a DB.
type stubUserRepo struct {
	userRepository
	roleNames  []string
	staffUsers []modelsUser.User
	allUsers   []modelsUser.User
	allTotal   int64
	lastPage   int
	lastLimit  int
}

func (s *stubUserRepo) FindAll(ctx context.Context, page, limit int) ([]modelsUser.User, int64, error) {
	s.lastPage = page
	s.lastLimit = limit
	return s.allUsers, s.allTotal, nil
}

func (s *stubUserRepo) FindByRoleNames(ctx context.Context, names []string) ([]modelsUser.User, error) {
	s.roleNames = names
	return s.staffUsers, nil
}

func userFixture(id, role string) modelsUser.User {
	return modelsUser.User{
		ID:     id,
		Email:  id + "@example.com",
		RoleID: role,
		Role:   &modelsUser.RoleSnapshot{ID: role, Name: role},
	}
}

// TestGetStaffUsers_FiltersCustomers is the G23 regression: the /admin/staff
// list must be sourced from a role-filtered lookup (admin + superadmin) so
// customer accounts and their PII never enter the response, while pagination
// still applies on top.
func TestGetStaffUsers_FiltersCustomers(t *testing.T) {
	repo := &stubUserRepo{
		staffUsers: []modelsUser.User{
			userFixture("u-admin-1", "admin"),
			userFixture("u-admin-2", "admin"),
			userFixture("u-super-1", "superadmin"),
		},
	}
	svc := NewUserService(repo)

	users, total, err := svc.GetStaffUsers(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("GetStaffUsers: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(users) != 2 {
		t.Fatalf("page 1 returned %d users, want 2", len(users))
	}
	if len(repo.roleNames) != 2 || repo.roleNames[0] != "admin" || repo.roleNames[1] != "superadmin" {
		t.Fatalf("FindByRoleNames called with %v, want [admin superadmin]", repo.roleNames)
	}

	page2, total2, err := svc.GetStaffUsers(context.Background(), 2, 2)
	if err != nil {
		t.Fatalf("GetStaffUsers page 2: %v", err)
	}
	if total2 != 3 {
		t.Fatalf("page 2 total = %d, want 3", total2)
	}
	if len(page2) != 1 {
		t.Fatalf("page 2 returned %d users, want 1", len(page2))
	}
}

// TestGetStaffUsers_EmptyStaff ensures an empty staff set returns an empty
// (non-nil) slice so the API serializes [] instead of null.
func TestGetStaffUsers_EmptyStaff(t *testing.T) {
	svc := NewUserService(&stubUserRepo{})
	users, total, err := svc.GetStaffUsers(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("GetStaffUsers: %v", err)
	}
	if total != 0 {
		t.Fatalf("total = %d, want 0", total)
	}
	if users == nil {
		t.Fatal("expected a non-nil empty slice")
	}
	if len(users) != 0 {
		t.Fatalf("got %d users, want 0", len(users))
	}
}

// TestGetUsers_StillReturnsAll is the no-regression guard for the non-affected
// path: the generic admin users list (AdminGetUsers) and the customer XLSX
// export both rely on GetUsers returning every account, so the staff filter
// must NOT be folded into it.
func TestGetUsers_StillReturnsAll(t *testing.T) {
	repo := &stubUserRepo{
		allUsers: []modelsUser.User{
			userFixture("u-admin-1", "admin"),
			userFixture("u-cust-1", "customer"),
			userFixture("u-super-1", "superadmin"),
		},
		allTotal: 3,
	}
	svc := NewUserService(repo)

	users, total, err := svc.GetUsers(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}
	if total != 3 {
		t.Fatalf("total = %d, want 3", total)
	}
	if len(users) != 3 {
		t.Fatalf("GetUsers returned %d users, want all 3 including customers", len(users))
	}
}
