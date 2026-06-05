package api

import (
	"net/http"
	"path/filepath"
	"strings"

	"candypro/api/internal/config"

	"github.com/gin-gonic/gin"
)

// registerUploadRoutes 替代对整棵上传目录的无鉴权 Static：敏感子目录禁止直链，须走 API 下载。
func registerUploadRoutes(router *gin.Engine, cfg *config.Config) {
	router.GET("/uploads/*filepath", serveUploadFile(cfg))
}

func serveUploadFile(cfg *config.Config) gin.HandlerFunc {
	base := filepath.Clean(cfg.Upload.UploadPath)
	return func(c *gin.Context) {
		if cfg.Upload.StorageDriver == "s3" || cfg.Upload.StorageDriver == "oss" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		raw := strings.TrimPrefix(c.Param("filepath"), "/")
		rel := filepath.ToSlash(filepath.Clean(raw))
		if rel == "" || rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if strings.HasPrefix(rel, "payment-proofs/") || rel == "payment-proofs" ||
			strings.HasPrefix(rel, "kyb/") || rel == "kyb" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Direct access to this path is not allowed. Use the authenticated download API.",
			})
			c.Abort()
			return
		}
		full := filepath.Join(base, filepath.FromSlash(rel))
		absBase, err1 := filepath.Abs(base)
		absFull, err2 := filepath.Abs(full)
		if err1 != nil || err2 != nil {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		sep := string(filepath.Separator)
		if absFull != absBase && !strings.HasPrefix(absFull, absBase+sep) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.File(full)
	}
}
