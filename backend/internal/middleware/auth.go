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
// When Redis session store is enabled, validates access jti in Redis when available;
// falls back to JWT claims on Redis outage or missing session (eviction / write miss).
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
				case errors.Is(serr, authsession.ErrSessionNotFound),
					errors.Is(serr, authsession.ErrStoreUnavailable):
					// JWT 签名与 exp 已校验，Redis 会话缺失或不可达时降级为纯 JWT
					log.Printf("auth: Redis session unavailable (%v), using JWT claims for jti=%s", serr, jti)
				case serr != nil:
					log.Printf("auth: Redis session check error (%v), using JWT claims for jti=%s", serr, jti)
				}
			}
		}

		c.Set("userID", sub)
		c.Set("userEmail", email)
		c.Set("userRole", role)

		c.Next()
	}
}
