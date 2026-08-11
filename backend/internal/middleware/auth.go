package middleware

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/authcookie"
	"candypro/api/internal/pkg/authsession"
	"candypro/api/internal/pkg/jwtutil"
	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware verifies JWT from Authorization header or HttpOnly cookie.
// When Redis session store is enabled, it REQUIRES the access jti session to exist:
// a revoked/missing session is denied with 401 (logout, admin revoke, password
// rotation must take effect), and a Redis outage fails closed with 503 rather than
// trusting an unverifiable JWT. When Redis is disabled the JWT claims are used
// as-is (the Postgres fallback does not track access sessions).
func AuthMiddleware(cfg *config.Config, sessions authsession.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := authcookie.TokenFromRequest(c)
		if tokenString == "" {
			response.ErrorResp(c, http.StatusUnauthorized, "auth_header_required")
			c.Abort()
			return
		}

		claims, err := jwtutil.ValidateJWT(tokenString, cfg.JWT.Secret)
		if err != nil {
			response.ErrorResp(c, http.StatusUnauthorized, "token_invalid")
			c.Abort()
			return
		}

		if purpose, hasPurpose := claims["purpose"]; hasPurpose && purpose != "" {
			response.ErrorResp(c, http.StatusUnauthorized, "token_type_mismatch")
			c.Abort()
			return
		}

		sub, ok := claims["sub"].(string)
		if !ok || strings.TrimSpace(sub) == "" {
			response.ErrorResp(c, http.StatusUnauthorized, "token_subject_invalid")
			c.Abort()
			return
		}

		email, _ := claims["email"].(string)
		role, _ := claims["role"].(string)

		if sessions != nil && sessions.UsesRedis() {
			jti, _ := claims["jti"].(string)
			if jti != "" {
				sess, serr := sessions.ValidateAccessSession(c.Request.Context(), jti)
				switch {
				case serr == nil && sess != nil:
					if sess.UserID != sub {
						response.ErrorResp(c, http.StatusUnauthorized, "token_invalid")
						c.Abort()
						return
					}
					if email == "" {
						email = sess.Email
					}
					if role == "" {
						role = sess.Role
					}
				case serr == nil && sess == nil:
					// Store returned no session and no error — revocation state cannot
					// be confirmed, so the token must not be trusted.
					log.Printf("auth: access session check returned empty for jti=%s, denying", jti)
					response.ErrorResp(c, http.StatusUnauthorized, "token_invalid")
					c.Abort()
					return
				case errors.Is(serr, authsession.ErrSessionNotFound):
					// Session was revoked (logout / admin revoke / password rotation) or
					// already expired — the token is no longer valid. This is NOT a Redis
					// outage; a missing session must deny the request.
					log.Printf("auth: access session for jti=%s not found/revoked, denying request", jti)
					response.ErrorResp(c, http.StatusUnauthorized, "token_invalid")
					c.Abort()
					return
				default:
					// ErrStoreUnavailable or any unexpected error: the revocation state
					// cannot be verified, so fail closed instead of trusting an
					// unverifiable JWT.
					log.Printf("auth: access session check failed for jti=%s: %v; failing closed", jti, serr)
					response.ErrorResp(c, http.StatusServiceUnavailable, "service_unavailable")
					c.Abort()
					return
				}
			}
		}

		c.Set("userID", sub)
		c.Set("userEmail", email)
		c.Set("userRole", role)

		c.Next()
	}
}
