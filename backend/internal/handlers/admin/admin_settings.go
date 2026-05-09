package admin

import (
	"candypro/api/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetSettings returns system settings, optionally filtered by category.
func (h *Handler) GetSettings(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	category := strings.TrimSpace(c.Query("category"))

	var result interface{}
	var err error

	if category != "" {
		result, err = h.services.SystemSetting.FindByCategory(c.Request.Context(), category)
	} else {
		result, err = h.services.SystemSetting.FindAll(c.Request.Context())
	}

	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "settings_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// UpdateSetting upserts a single system setting.
func (h *Handler) UpdateSetting(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Key      string `json:"key" binding:"required"`
		Value    string `json:"value" binding:"required"`
		Category string `json:"category"`
	}
	if !utils.BindJSONOrInvalid(c, &req) {
		return
	}

	setting := &settingPayload{
		Key:      req.Key,
		Value:    req.Value,
		Category: req.Category,
	}

	if err := h.services.SystemSetting.Upsert(c.Request.Context(), setting.Key, setting.Value, setting.Category); err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "settings_update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Setting updated", "key": setting.Key})
}

// settingPayload is a local DTO for system setting data.
type settingPayload struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Category string `json:"category"`
}
