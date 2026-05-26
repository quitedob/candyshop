package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	"candypro/api/internal/pkg/response"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var systemSettingMeta = map[string]struct {
	Label       string
	Description string
}{
	"site_name":             {Label: "Site name", Description: "Public site / brand name"},
	"default_currency":      {Label: "Default currency", Description: "Default ISO currency code for quotes and orders"},
	"support_email":         {Label: "Support email", Description: "Customer support contact email"},
	"smtp_host":             {Label: "SMTP host", Description: "Outbound mail server hostname"},
	"smtp_port":             {Label: "SMTP port", Description: "Outbound mail server port"},
	"smtp_from":             {Label: "From address", Description: "Default sender email address"},
	"default_incoterms":     {Label: "Default incoterms", Description: "Default trade terms for new transactions"},
	"default_payment_terms": {Label: "Default payment terms", Description: "Default payment terms text for trade documents"},
	"notify_new_inquiry":    {Label: "Notify on new inquiry", Description: "Send admin notification when a new inquiry arrives"},
	"notify_new_order":      {Label: "Notify on new order", Description: "Send admin notification when a new order is placed"},
}

type systemSettingView struct {
	Key         string  `json:"key"`
	Value       string  `json:"value"`
	Category    string  `json:"category"`
	Label       string  `json:"label,omitempty"`
	Description string  `json:"description,omitempty"`
	UpdatedBy   *string `json:"updatedBy,omitempty"`
	UpdatedAt   string  `json:"updatedAt,omitempty"`
}

func toSystemSettingViews(settings []modelsCommon.SystemSetting) []systemSettingView {
	out := make([]systemSettingView, 0, len(settings))
	for _, setting := range settings {
		view := systemSettingView{
			Key:       setting.Key,
			Value:     setting.Value,
			Category:  setting.Category,
			UpdatedBy: setting.UpdatedBy,
		}
		if !setting.UpdatedAt.IsZero() {
			view.UpdatedAt = setting.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00")
		}
		if meta, ok := systemSettingMeta[setting.Key]; ok {
			view.Label = meta.Label
			view.Description = meta.Description
		}
		out = append(out, view)
	}
	return out
}

// GetSettings returns system settings, optionally filtered by category.
func (h *Handler) GetSettings(c *gin.Context) {
	if h.services == nil || h.services.SystemSetting == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	category := strings.TrimSpace(c.Query("category"))

	var settings []modelsCommon.SystemSetting
	var err error

	if category != "" {
		settings, err = h.services.SystemSetting.FindByCategory(c.Request.Context(), category)
	} else {
		settings, err = h.services.SystemSetting.FindAll(c.Request.Context())
	}

	if err != nil {
		log.Printf("admin GetSettings category=%q: %v", category, err)
		response.ErrorResp(c, http.StatusInternalServerError, "settings_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": toSystemSettingViews(settings)})
}

// UpdateSetting upserts a single system setting.
func (h *Handler) UpdateSetting(c *gin.Context) {
	if h.services == nil || h.services.SystemSetting == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	key := strings.TrimSpace(c.Param("key"))
	if key == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}

	var req struct {
		Key      string `json:"key"`
		Value    string `json:"value" binding:"required"`
		Category string `json:"category"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Key != "" && req.Key != key {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if err := h.services.SystemSetting.Upsert(c.Request.Context(), key, req.Value, req.Category); err != nil {
		log.Printf("admin UpdateSetting key=%q: %v", key, err)
		response.ErrorResp(c, http.StatusInternalServerError, "settings_update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Setting updated", "key": key})
}
