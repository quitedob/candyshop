package middleware

import (
	"candypro/api/internal/config"
	modelsAuth "candypro/api/internal/models/auth"
	"candypro/api/internal/pkg/jwtutil"
	productRepo "candypro/api/internal/repository/product"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SupplierAuthMiddleware 支持 X-Supplier-Key、API Key Bearer 或 supplier 角色 JWT
func SupplierAuthMiddleware(cfg *config.Config, supplierRepo *productRepo.SupplierRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("X-Supplier-Key"))
		authHeader := c.GetHeader("Authorization")
		bearer := ""
		if strings.HasPrefix(authHeader, "Bearer ") {
			bearer = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		}

		// JWT：supplier 角色用户按邮箱匹配供应商记录
		if bearer != "" && strings.Count(bearer, ".") == 2 && cfg != nil {
			claims, err := jwtutil.ValidateJWT(bearer, cfg.JWT.Secret)
			if err == nil {
				if purpose, ok := claims["purpose"]; ok && purpose != "" && purpose != nil {
					// 非 access token，跳过 JWT 路径
				} else {
					role, _ := claims["role"].(string)
					email, _ := claims["email"].(string)
					if role == modelsAuth.Supplier && email != "" {
						sup, serr := supplierRepo.FindByEmail(c.Request.Context(), email)
						if serr == nil && sup != nil && sup.IsActive {
							c.Set("supplierID", sup.ID)
							c.Next()
							return
						}
					}
				}
			}
		}

		// API Key 路径
		if key == "" {
			key = bearer
		}
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "supplier_key_required", "message": "Supplier authentication required"})
			return
		}

		sup, err := supplierRepo.FindByAPIKey(c.Request.Context(), key)
		if err != nil || sup == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "supplier_not_found", "message": "Invalid supplier key"})
			return
		}
		if !sup.IsActive {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "supplier_inactive", "message": "Supplier account is inactive"})
			return
		}

		c.Set("supplierID", sup.ID)
		c.Next()
	}
}
