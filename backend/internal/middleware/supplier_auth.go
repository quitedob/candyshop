package middleware

import (
	productRepo "candypro/api/internal/repository/product"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// SupplierAuthMiddleware authenticates suppliers via X-Supplier-Key header.
// It sets "supplierID" in the gin context on success.
func SupplierAuthMiddleware(supplierRepo *productRepo.SupplierRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(c.GetHeader("X-Supplier-Key"))
		if key == "" {
			// Also check Bearer token
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				key = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if key == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "supplier_key_required", "message": "X-Supplier-Key header is required"})
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
