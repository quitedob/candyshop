package middleware

import (
	modelsAuth "candypro/api/internal/models/auth"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AttachAdminPermissions 将角色权限写入 context，供 RequirePermission 使用。
func AttachAdminPermissions() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")
		roleStr, _ := role.(string)
		var perms []string
		switch roleStr {
		case modelsAuth.SuperAdmin:
			perms = []string{"*"}
		case modelsAuth.Admin:
			perms = []string{"admin:*"}
		default:
			perms = []string{}
		}
		c.Set("userPermissions", perms)
		c.Next()
	}
}

// permissionMatches 判断用户权限是否满足所需权限（支持 * 与 admin:* 前缀）。
func permissionMatches(userPerm, required string) bool {
	if userPerm == "*" || required == "*" {
		return true
	}
	if userPerm == required {
		return true
	}
	if strings.HasSuffix(userPerm, ":*") {
		prefix := strings.TrimSuffix(userPerm, ":*")
		return required == prefix || strings.HasPrefix(required, prefix+":")
	}
	return false
}

// RequireAdminWritePermission 对非 GET/HEAD/OPTIONS 请求校验 admin 写权限。
func RequireAdminWritePermission() gin.HandlerFunc {
	writeCheck := RequirePermission("admin:write", "admin:*", "*")
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		writeCheck(c)
	}
}

// RequirePermission restricts access to users with specific permissions
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userPerms := permissionsFromContext(c)
		if len(userPerms) == 0 {
			response.ErrorResp(c, http.StatusForbidden, "forbidden_no_permissions")
			c.Abort()
			return
		}

		for _, reqPerm := range permissions {
			for _, userPerm := range userPerms {
				if permissionMatches(userPerm, reqPerm) {
					c.Next()
					return
				}
			}
		}

		response.ErrorResp(c, http.StatusForbidden, "forbidden_insufficient_permissions")
		c.Abort()
	}
}

func permissionsFromContext(c *gin.Context) []string {
	raw, exists := c.Get("userPermissions")
	if !exists {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []interface{}:
		out := make([]string, 0, len(v))
		for _, p := range v {
			if s, ok := p.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
