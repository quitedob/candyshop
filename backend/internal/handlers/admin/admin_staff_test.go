package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
	userRepo "candypro/api/internal/repository/user"
	servicesCommon "candypro/api/internal/services/common"
	userSvc "candypro/api/internal/services/user"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// G23 regression tests (follow-up round). The /admin/staff screen must be fed by
// the role-filtered GetStaffUsers (admin + superadmin) and never by the generic
// GetUsers/FindAll path, which returns every account including customers with
// PII (email, firstName, lastName, phone, company). Before the wiring fix the
// handler called h.services.User.GetUsers and the customer record — with all of
// its PII — landed in the staff payload.

func setupStaffListTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&modelsAuth.Role{}, &modelsUser.User{}); err != nil {
		t.Fatalf("migrate roles+users: %v", err)
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

	// Staff carry the usual staff attributes; the customer carries rich PII that
	// must NOT appear in the staff list.
	for _, u := range []modelsUser.User{
		{ID: "u-admin-1", Email: "alice@candypro.com", PasswordHash: "x", FirstName: "Alice", LastName: "Admin", Phone: "111", Company: "CandyPro", Status: "active", RoleID: "r-admin"},
		{ID: "u-super-1", Email: "bob@candypro.com", PasswordHash: "x", FirstName: "Bob", LastName: "Super", Phone: "222", Company: "CandyPro", Status: "active", RoleID: "r-super"},
		{ID: "u-cust-1", Email: "carol@acme.com", PasswordHash: "x", FirstName: "Carol", LastName: "Customer", Phone: "333", Company: "ACME Inc", Status: "active", RoleID: "r-customer"},
	} {
		if err := db.Create(&u).Error; err != nil {
			t.Fatalf("seed user %s: %v", u.Email, err)
		}
	}
	return db
}

func newStaffListTestHandler(t *testing.T, db *gorm.DB) *Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	ur := userRepo.NewUserRepository(db)
	svcs := &servicesCommon.AdminPortalServices{
		User: userSvc.NewUserService(ur),
	}
	return &Handler{services: svcs}
}

func staffListTestCtx() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "u-admin-1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/staff", nil)
	return c, w
}

// The staff list must contain ONLY admin/superadmin accounts: the customer row
// and its PII are absent, and the pagination total reflects the staff set.
func TestGetStaffList_ReturnsOnlyStaffRoles(t *testing.T) {
	db := setupStaffListTestDB(t)
	h := newStaffListTestHandler(t, db)

	c, w := staffListTestCtx()
	h.GetStaffList(c)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /admin/staff: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Data       []modelsUser.User `json:"data"`
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse staff response %q: %v", w.Body.String(), err)
	}

	if resp.Pagination.Total != 2 {
		t.Fatalf("total = %d, want 2 (staff only, customer excluded)", resp.Pagination.Total)
	}
	if len(resp.Data) != 2 {
		t.Fatalf("returned %d staff, want 2", len(resp.Data))
	}

	for _, u := range resp.Data {
		if u.Role == nil {
			t.Fatalf("staff user %s has no role loaded", u.ID)
		}
		if u.Role.Name != modelsAuth.Admin && u.Role.Name != modelsAuth.SuperAdmin {
			t.Fatalf("staff list leaked non-staff role %q on user %s", u.Role.Name, u.ID)
		}
		// No customer PII may appear in the staff payload.
		if u.Email == "carol@acme.com" || u.FirstName == "Carol" || u.LastName == "Customer" ||
			u.Phone == "333" || u.Company == "ACME Inc" {
			t.Fatalf("customer PII leaked into /admin/staff: %+v", u)
		}
	}
}

// Pagination must still apply on the staff set: page 1 at limit 1 returns one
// staff row and a total of 2 (not the full user table).
func TestGetStaffList_PaginationAppliesOnStaffSet(t *testing.T) {
	db := setupStaffListTestDB(t)
	h := newStaffListTestHandler(t, db)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", "u-admin-1")
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/staff?page=1&limit=1", nil)
	h.GetStaffList(c)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /admin/staff?page=1&limit=1: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Data []modelsUser.User `json:"data"`
		Pagination struct {
			Total      int `json:"total"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse staff response %q: %v", w.Body.String(), err)
	}
	if resp.Pagination.Total != 2 {
		t.Fatalf("total = %d, want 2", resp.Pagination.Total)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("limit=1 returned %d rows, want 1", len(resp.Data))
	}
	if resp.Pagination.TotalPages != 2 {
		t.Fatalf("totalPages = %d, want 2", resp.Pagination.TotalPages)
	}
	// The single row must be a staff member, never the customer.
	if strings.Contains(w.Body.String(), "carol@acme.com") {
		t.Fatalf("customer PII leaked into paginated staff response: %s", w.Body.String())
	}
}
