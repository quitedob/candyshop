package authcookie

import (
	"net/http"
	"strings"
	"time"

	"candypro/api/internal/config"

	"github.com/gin-gonic/gin"
)

const (
	AccessTokenCookie  = "auth_token"
	RefreshTokenCookie = "refresh_token"
)

// SetAuthCookies writes HttpOnly JWT cookies on the response.
//
// Cookie SameSite policy (M-21):
//   - Access token: SameSiteLax — keeps cross-site top-level navigation working
//     (email links, OAuth return URLs to /customer/...).
//   - Refresh token: SameSiteStrict — only ever sent in same-site contexts, since
//     /auth/refresh is invoked from in-app fetch calls. This prevents a refresh
//     token from leaking on third-party top-level GET navigations and tightens
//     CSRF posture on the long-lived credential.
func SetAuthCookies(c *gin.Context, cfg *config.Config, accessToken, refreshToken string) {
	secure := cfg.Server.Environment == "production"
	accessMaxAge := cfg.JWT.AccessTokenDuration * 60
	refreshMaxAge := cfg.JWT.RefreshTokenDuration * 24 * 60 * 60
	setCookie(c, AccessTokenCookie, accessToken, accessMaxAge, secure, http.SameSiteLaxMode)
	setCookie(c, RefreshTokenCookie, refreshToken, refreshMaxAge, secure, http.SameSiteStrictMode)
}

// ClearAuthCookies removes auth cookies.
func ClearAuthCookies(c *gin.Context, cfg *config.Config) {
	secure := cfg.Server.Environment == "production"
	setCookie(c, AccessTokenCookie, "", -1, secure, http.SameSiteLaxMode)
	setCookie(c, RefreshTokenCookie, "", -1, secure, http.SameSiteStrictMode)
}

func setCookie(c *gin.Context, name, value string, maxAge int, secure bool, sameSite http.SameSite) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
	})
}

// TokenFromRequest reads Bearer header or auth_token cookie.
func TokenFromRequest(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" && parts[1] != "" {
			return parts[1]
		}
	}
	if cookie, err := c.Cookie(AccessTokenCookie); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

// RefreshTokenFromRequest reads JSON body refresh_token or cookie.
func RefreshTokenFromRequest(c *gin.Context, bodyToken string) string {
	if strings.TrimSpace(bodyToken) != "" {
		return strings.TrimSpace(bodyToken)
	}
	if cookie, err := c.Cookie(RefreshTokenCookie); err == nil && cookie != "" {
		return cookie
	}
	return ""
}

// CookieLifetime returns access token cookie duration for tests.
func CookieLifetime(cfg *config.Config) time.Duration {
	return time.Duration(cfg.JWT.AccessTokenDuration) * time.Minute
}
