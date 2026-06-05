package api

import (
	"crypto/subtle"
	"net/http"

	"candypro/api/internal/config"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/storage"

	"github.com/gin-gonic/gin"
)

func registerUploadScanHook(group *gin.RouterGroup, st *storage.Manager, cfg *config.Config) {
	group.POST("/uploads/:fileId/scan-result", func(c *gin.Context) {
		if st == nil || cfg.Upload.VirusScanWebhookSecret == "" {
			response.ErrorResp(c, http.StatusNotFound, "not_found")
			return
		}
		actual := []byte(c.GetHeader("X-Virus-Scan-Secret"))
		expected := []byte(cfg.Upload.VirusScanWebhookSecret)
		if subtle.ConstantTimeCompare(actual, expected) != 1 {
			response.ErrorResp(c, http.StatusUnauthorized, "unauthorized")
			return
		}
		var req struct {
			Status  string `json:"status" binding:"required"`
			Message string `json:"message"`
		}
		if !response.BindJSONOrInvalid(c, &req) {
			return
		}
		file, err := st.RecordScanResult(c.Request.Context(), c.Param("fileId"), req.Status, req.Message)
		if err != nil {
			response.InvalidResp(c, "invalid_request")
			return
		}
		c.JSON(http.StatusOK, gin.H{"file": file})
	})
}
