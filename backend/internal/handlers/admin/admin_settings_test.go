package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"candypro/api/internal/config"
	"candypro/api/internal/database"
	repositoryCommon "candypro/api/internal/repository/common"
	adminportalscope "candypro/api/internal/repository/scopes/adminportalscope"
	servicesCommon "candypro/api/internal/services/common"
	adminportalSvc "candypro/api/internal/services/scopes/adminportalscope"

	"github.com/gin-gonic/gin"
)

func TestGetSettingsGeneral(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("JWT_SECRET", "b32558c730eabe690c0dff656c3d89efe4c31387ee700320eefb5def7e7c1718")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_USER", "postgres")
	t.Setenv("DB_PASSWORD", "1234")
	t.Setenv("DB_NAME", "candypro")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("ENVIRONMENT", "development")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}
	db, err := database.Connect(&cfg.Database)
	if err != nil {
		t.Skip("database unavailable:", err)
	}

	repos := repositoryCommon.NewRepositories(db)
	svcs := servicesCommon.NewServices(repos, cfg, db)
	h := NewHandler(cfg, svcs.AdminPortal, svcs.CountryPaymentPolicy, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/admin/settings?category=general", nil)

	h.GetSettings(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var body struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("json: %v body=%s", err, w.Body.String())
	}
	t.Logf("settings count: %d", len(body.Data))
	_ = adminportalscope.Repositories{}
	_ = adminportalSvc.Services{}
}
