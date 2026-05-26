package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// csrfSkipPrefixes 外部回调等无需 Origin 校验的路径前缀。
var csrfSkipPrefixes = []string{
	"/api/v1/system/stripe-webhook",
	"/api/v1/system/paypal-webhook",
}

// CookieCSRFGuard 对 Cookie 认证的变更请求校验 Origin/Referer，降低 CSRF 风险。
// 带 Authorization: Bearer 的请求视为 API 客户端，跳过校验。
func CookieCSRFGuard(cfg *config.Config) gin.HandlerFunc {
	allowed := buildOriginSet(cfg.Security.CORSAllowedOrigins)

	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		path := c.Request.URL.Path
		for _, prefix := range csrfSkipPrefixes {
			if strings.HasPrefix(path, prefix) {
				c.Next()
				return
			}
		}

		auth := c.GetHeader("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			c.Next()
			return
		}

		origin := strings.TrimSpace(c.GetHeader("Origin"))
		if origin == "" {
			origin = originFromReferer(c.GetHeader("Referer"))
		}
		if origin == "" {
			// 非浏览器客户端（如 curl）无 Origin；生产环境拒绝 Cookie 变更请求。
			if cfg.Server.Environment == "production" {
				response.ErrorResp(c, http.StatusForbidden, "csrf_origin_required")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		if !allowed[origin] {
			response.ErrorResp(c, http.StatusForbidden, "csrf_origin_invalid")
			c.Abort()
			return
		}

		c.Next()
	}
}

func buildOriginSet(origins []string) map[string]bool {
	set := make(map[string]bool, len(origins))
	for _, o := range origins {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			set[trimmed] = true
		}
	}
	return set
}

func originFromReferer(referer string) string {
	referer = strings.TrimSpace(referer)
	if referer == "" {
		return ""
	}
	u, err := url.Parse(referer)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
