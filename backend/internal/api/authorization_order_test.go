package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	adminRoutes "candypro/api/internal/api/routes/adminportal"
	userRoutes "candypro/api/internal/api/routes/userportal"
	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/handlers/admin"
	"candypro/api/internal/handlers/auth"
	"candypro/api/internal/handlers/customer"
	"candypro/api/internal/handlers/system"
	modelsAuth "candypro/api/internal/models/auth"
	"candypro/api/internal/pkg/jwtutil"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

func TestPortalAuthorizationPrecedesIdempotency(t *testing.T) {
	const signingSecret = "isolated-route-order-test-secret-at-least-32-characters"
	configuration := &config.Config{JWT: config.JWTConfig{Secret: signingSecret}}
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Exec(`CREATE TABLE users (id TEXT PRIMARY KEY, status TEXT, deleted_at DATETIME)`).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Exec(`INSERT INTO users (id, status) VALUES (?, ?)`, "route-owner", "pending").Error; err != nil {
		t.Fatal(err)
	}
	routeHandlers := &handlers.Handlers{AuthScope: &auth.Handler{}, AdminPortal: &admin.Handler{}, UserPortal: &customer.Handler{}, System: &system.Handler{}}
	router := gin.New()
	postAuthorization := func(requestContext *gin.Context) {
		if requestContext.GetString("userID") != "route-owner" {
			t.Error("idempotency reached without authenticated owner")
		}
		if requestContext.GetString("userRole") == modelsAuth.Admin {
			if _, found := requestContext.Get("userPermissions"); !found {
				t.Error("idempotency reached before permissions attachment")
			}
		}
		requestContext.AbortWithStatus(http.StatusAccepted)
	}
	adminRoutes.Register(router.Group("/admin"), routeHandlers, configuration, postAuthorization)
	userRoutes.Register(router.Group("/user"), routeHandlers, configuration, database, postAuthorization)
	for _, testCase := range []struct {
		path, role string
		wantStatus int
	}{
		{"/admin/orders", "", http.StatusUnauthorized},
		{"/admin/orders", modelsAuth.User, http.StatusForbidden},
		{"/admin/orders", modelsAuth.Admin, http.StatusAccepted},
		{"/user/orders", modelsAuth.Admin, http.StatusForbidden},
		{"/user/orders", modelsAuth.User, http.StatusAccepted},
		{"/user/inquiries", modelsAuth.User, http.StatusForbidden},
	} {
		t.Run(testCase.path+testCase.role, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, testCase.path, nil)
			if testCase.role != "" {
				token, err := jwtutil.GenerateJWT(jwt.MapClaims{"sub": "route-owner", "role": testCase.role, "exp": time.Now().Add(time.Hour).Unix()}, signingSecret)
				if err != nil {
					t.Fatal(err)
				}
				request.Header.Set("Authorization", "Bearer "+token)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != testCase.wantStatus {
				t.Fatalf("status=%d want=%d body=%s", response.Code, testCase.wantStatus, response.Body.String())
			}
		})
	}
}
