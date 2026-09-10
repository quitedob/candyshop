package systemroutes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"candypro/api/internal/config"
	"candypro/api/internal/handlers"
	"candypro/api/internal/handlers/auth"
	"candypro/api/internal/handlers/system"
	"candypro/api/internal/pkg/jwtutil"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// TestGenerateQuotationRoleGuard 回归验证 H5：/system/generate-quotation 会向模型与
// 响应注入内部成本栈（物流/关税/标签摊销/目标毛利），属于内部定价数据，因此该路由
// 必须仅允许 admin/superadmin 访问，任何已登录客户（customer/supplier）都不得调用。
func TestGenerateQuotationRoleGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const secret = "unit-test-secret-key-at-least-32-chars-long"
	cfg := &config.Config{JWT: config.JWTConfig{Secret: secret}}

	// 路由注册只要求 AuthScope.SessionStore()（零值 handler 返回 nil）与 System handler
	// 方法绑定，无需完整装配依赖；角色守卫在 handler 执行之前完成判定。
	h := &handlers.Handlers{
		AuthScope: &auth.Handler{},
		System:    &system.Handler{},
	}

	r := gin.New()
	Register(r.Group("/api/v1/system"), h, cfg)

	mint := func(role string) string {
		tok, err := jwtutil.GenerateJWT(jwt.MapClaims{
			"sub":  "user-1",
			"role": role,
			"exp":  time.Now().Add(time.Hour).Unix(),
		}, secret)
		if err != nil {
			t.Fatalf("mint token: %v", err)
		}
		return tok
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "no token rejected", wantStatus: http.StatusUnauthorized},
		{name: "customer forbidden", authHeader: "Bearer " + mint("customer"), wantStatus: http.StatusForbidden},
		{name: "supplier forbidden", authHeader: "Bearer " + mint("supplier"), wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/generate-quotation", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}

	// admin / superadmin 必须通过角色守卫（不应得到 401/403；零值 handler 因 AI 未启用
	// 返回 503，仅用于确认守卫放行而非路由缺失）。
	for _, role := range []string{"admin", "superadmin"} {
		t.Run(role+" passes guard", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/generate-quotation", nil)
			req.Header.Set("Authorization", "Bearer "+mint(role))
			r.ServeHTTP(w, req)
			if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
				t.Fatalf("expected %s to pass the role guard, got %d", role, w.Code)
			}
		})
	}
}

func TestResumeAgentRoleGuard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const secret = "unit-test-secret-key-at-least-32-chars-long"
	cfg := &config.Config{JWT: config.JWTConfig{Secret: secret}}

	// 路由注册只要求 AuthScope.SessionStore()（零值 handler 返回 nil）与 System handler
	// 方法绑定，无需完整装配依赖；角色守卫在 handler 执行之前完成判定。
	h := &handlers.Handlers{
		AuthScope: &auth.Handler{},
		System:    &system.Handler{},
	}

	r := gin.New()
	Register(r.Group("/api/v1/system"), h, cfg)

	mint := func(role string) string {
		tok, err := jwtutil.GenerateJWT(jwt.MapClaims{
			"sub":  "user-1",
			"role": role,
			"exp":  time.Now().Add(time.Hour).Unix(),
		}, secret)
		if err != nil {
			t.Fatalf("mint token: %v", err)
		}
		return tok
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{name: "no token rejected", wantStatus: http.StatusUnauthorized},
		{name: "customer forbidden", authHeader: "Bearer " + mint("customer"), wantStatus: http.StatusForbidden},
		{name: "supplier forbidden", authHeader: "Bearer " + mint("supplier"), wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/ai/resume", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}

	// admin / superadmin 必须通过角色守卫（不应得到 401/403；零值 handler 因 AI 未启用
	// 返回 503，仅用于确认守卫放行而非路由缺失）。
	for _, role := range []string{"admin", "superadmin"} {
		t.Run(role+" passes guard", func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/v1/system/ai/resume", nil)
			req.Header.Set("Authorization", "Bearer "+mint(role))
			r.ServeHTTP(w, req)
			if w.Code == http.StatusUnauthorized || w.Code == http.StatusForbidden {
				t.Fatalf("expected %s to pass the role guard, got %d", role, w.Code)
			}
		})
	}
}
