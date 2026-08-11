package admin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	modelsAuth "candypro/api/internal/models/auth"
	modelsUser "candypro/api/internal/models/user"
	authrepo "candypro/api/internal/repository/auth"
	userRepo "candypro/api/internal/repository/user"
	servicesAuth "candypro/api/internal/services/auth"
	servicesCommon "candypro/api/internal/services/common"
	userSvc "candypro/api/internal/services/user"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Regression tests for H1: only a superadmin may mint or promote admin /
// superadmin roles. A plain admin must be rejected with 403 by every role
// assignment path (AdminCreateUser, AdminUpdateUserRole, AdminUpdateUser),
// and must be able to continue granting non-privileged roles.

func setupRoleGrantTestDB(t *testing.T) *gorm.DB {
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

func newRoleGrantTestHandler(t *testing.T) *Handler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	db := setupRoleGrantTestDB(t)
	ur := userRepo.NewUserRepository(db)
	rr := authrepo.NewRoleRepository(db)
	authSvc := servicesAuth.NewAuthService(ur, rr, nil, nil, nil)
	svcs := &servicesCommon.AdminPortalServices{
		User: userSvc.NewUserService(ur),
		Auth: authSvc,
	}
	return &Handler{services: svcs}
}

func roleGrantTestCtx(method, callerID, body string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", callerID)
	if body != "" {
		c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")
	} else {
		c.Request = httptest.NewRequest(method, "/", nil)
	}
	return c, w
}

// Admin attempting to mint a superadmin must be rejected with 403 and no user
// record may be created.
func TestAdminCreateUser_AdminCannotMintSuperadmin(t *testing.T) {
	h := newRoleGrantTestHandler(t)
	body := `{"email":"new@example.com","password":"Test123456","firstName":"New","lastName":"User","roleName":"superadmin"}`
	c, w := roleGrantTestCtx(http.MethodPost, "u-admin", body)
	h.AdminCreateUser(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("admin minting superadmin: status=%d, want 403, body=%s", w.Code, w.Body.String())
	}
	exists, err := h.services.User.EmailExists(context.Background(), "new@example.com")
	if err != nil {
		t.Fatalf("email exists check: %v", err)
	}
	if exists {
		t.Fatal("admin minted a superadmin account; user must not be created")
	}
}

// Admin attempting to promote a user to admin via AdminUpdateUserRole must be
// rejected with 403 and the target role left unchanged.
func TestAdminUpdateUserRole_AdminCannotPromoteToAdmin(t *testing.T) {
	h := newRoleGrantTestHandler(t)
	body := `{"roleName":"admin"}`
	c, w := roleGrantTestCtx(http.MethodPut, "u-admin", body)
	c.Params = gin.Params{{Key: "id", Value: "u-cust"}}
	h.AdminUpdateUserRole(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("admin promoting to admin: status=%d, want 403, body=%s", w.Code, w.Body.String())
	}
	got, err := h.services.User.GetByID(context.Background(), "u-cust")
	if err != nil {
		t.Fatalf("get target user: %v", err)
	}
	if got.RoleID != "r-customer" {
		t.Fatalf("target roleId=%q, want unchanged r-customer", got.RoleID)
	}
}

// Admin attempting to promote a user to superadmin via AdminUpdateUser (the
// /admin/users/{id} PUT profile path) must be rejected with 403.
func TestAdminUpdateUser_AdminCannotPromoteToSuperadmin(t *testing.T) {
	h := newRoleGrantTestHandler(t)
	body := `{"roleId":"r-super"}`
	c, w := roleGrantTestCtx(http.MethodPut, "u-admin", body)
	c.Params = gin.Params{{Key: "id", Value: "u-cust"}}
	h.AdminUpdateUser(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("admin promoting to superadmin via profile update: status=%d, want 403, body=%s", w.Code, w.Body.String())
	}
	got, err := h.services.User.GetByID(context.Background(), "u-cust")
	if err != nil {
		t.Fatalf("get target user: %v", err)
	}
	if got.RoleID != "r-customer" {
		t.Fatalf("target roleId=%q, want unchanged r-customer", got.RoleID)
	}
}

// A superadmin must still be able to mint a superadmin account.
func TestAdminCreateUser_SuperadminCanMintSuperadmin(t *testing.T) {
	h := newRoleGrantTestHandler(t)
	body := `{"email":"newsup@example.com","password":"Test123456","firstName":"New","lastName":"Super","roleName":"superadmin"}`
	c, w := roleGrantTestCtx(http.MethodPost, "u-super", body)
	h.AdminCreateUser(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("superadmin minting superadmin: status=%d, want 201, body=%s", w.Code, w.Body.String())
	}
	got, err := h.services.User.GetByEmail(context.Background(), "newsup@example.com")
	if err != nil {
		t.Fatalf("get created user: %v", err)
	}
	if got.RoleID != "r-super" {
		t.Fatalf("roleId=%q, want r-super", got.RoleID)
	}
}

// A superadmin must still be able to promote a user to admin.
func TestAdminUpdateUserRole_SuperadminCanPromoteToAdmin(t *testing.T) {
	h := newRoleGrantTestHandler(t)
	body := `{"roleName":"admin"}`
	c, w := roleGrantTestCtx(http.MethodPut, "u-super", body)
	c.Params = gin.Params{{Key: "id", Value: "u-cust"}}
	h.AdminUpdateUserRole(c)

	if w.Code != http.StatusOK {
		t.Fatalf("superadmin promoting to admin: status=%d, want 200, body=%s", w.Code, w.Body.String())
	}
	got, err := h.services.User.GetByID(context.Background(), "u-cust")
	if err != nil {
		t.Fatalf("get target user: %v", err)
	}
	if got.RoleID != "r-admin" {
		t.Fatalf("target roleId=%q, want r-admin", got.RoleID)
	}
}

// The guard must not break non-privileged grants: an admin may still create a
// customer account.
func TestAdminCreateUser_AdminCanCreateCustomer(t *testing.T) {
	h := newRoleGrantTestHandler(t)
	body := `{"email":"newcust@example.com","password":"Test123456","firstName":"New","lastName":"Cust","roleName":"customer"}`
	c, w := roleGrantTestCtx(http.MethodPost, "u-admin", body)
	h.AdminCreateUser(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("admin creating customer: status=%d, want 201, body=%s", w.Code, w.Body.String())
	}
	got, err := h.services.User.GetByEmail(context.Background(), "newcust@example.com")
	if err != nil {
		t.Fatalf("get created user: %v", err)
	}
	if got.RoleID != "r-customer" {
		t.Fatalf("roleId=%q, want r-customer", got.RoleID)
	}
}

// A request that carries no caller identity must be denied even if the target
// role is privileged (defense in depth when the auth middleware fails to set
// userID).
func TestAdminCreateUser_MissingCallerDenied(t *testing.T) {
	h := newRoleGrantTestHandler(t)
	body := `{"email":"noid@example.com","password":"Test123456","firstName":"N","lastName":"O","roleName":"admin"}`
	c, w := roleGrantTestCtx(http.MethodPost, "", body)
	h.AdminCreateUser(c)

	if w.Code != http.StatusForbidden {
		t.Fatalf("missing caller minting admin: status=%d, want 403, body=%s", w.Code, w.Body.String())
	}
}
