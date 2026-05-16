package system

import (
	"net/http"

	"candypro/api/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler holds dependencies for health checks.
type HealthHandler struct {
	db *gorm.DB
}

// NewHealthHandler creates a health check handler.
func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health responds with basic service status.
func (hh *HealthHandler) Health(c *gin.Context) {
	sqlDB, err := hh.db.DB()
	if err != nil {
		response.ErrorResp(c, http.StatusServiceUnavailable, "db_unreachable")
		return
	}
	if err := sqlDB.Ping(); err != nil {
		response.ErrorResp(c, http.StatusServiceUnavailable, "db_unreachable")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Ready responds when all dependencies are available.
func (hh *HealthHandler) Ready(c *gin.Context) {
	sqlDB, err := hh.db.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "db error"})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready", "reason": "db unreachable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
