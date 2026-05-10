package admin

import (
	"net/http"
	"strconv"

	modelsCommon "candypro/api/internal/models/common"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	translationSvc "candypro/api/internal/services/translation"

	"github.com/gin-gonic/gin"
)

type TranslationHandler struct {
	svc *translationSvc.TranslationService
}

func NewTranslationHandler(svc *translationSvc.TranslationService) *TranslationHandler {
	return &TranslationHandler{svc: svc}
}

// ListTranslations returns translations with optional filters.
func (h *TranslationHandler) ListTranslations(c *gin.Context) {
	group := c.Query("group")
	locale := c.Query("locale")
	search := c.Query("search")
	page, limit := pagination.ParsePagination(c, 20, 100)

	translations, total, err := h.svc.List(group, locale, search, page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       translations,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}

// GetTranslation returns a single translation by ID.
func (h *TranslationHandler) GetTranslation(c *gin.Context) {
	id, err := parseTranslationID(c)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	t, err := h.svc.GetByID(id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	c.JSON(http.StatusOK, t)
}

// CreateTranslation creates a new translation override.
func (h *TranslationHandler) CreateTranslation(c *gin.Context) {
	var req struct {
		Key    string `json:"key" binding:"required"`
		Locale string `json:"locale" binding:"required"`
		Value  string `json:"value" binding:"required"`
		Group  string `json:"group"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	userID := c.GetString("userID")
	t := &modelsCommon.Translation{
		Key:       req.Key,
		Locale:    req.Locale,
		Value:     req.Value,
		Group:     req.Group,
		IsActive:  true,
		UpdatedBy: &userID,
	}

	if err := h.svc.Create(t); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusCreated, t)
}

// UpdateTranslation updates an existing translation.
func (h *TranslationHandler) UpdateTranslation(c *gin.Context) {
	id, err := parseTranslationID(c)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	var req struct {
		Value    string `json:"value"`
		Group    string `json:"group"`
		IsActive *bool  `json:"isActive"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	existing, err := h.svc.GetByID(id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	userID := c.GetString("userID")
	existing.Value = req.Value
	if req.Group != "" {
		existing.Group = req.Group
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}
	existing.UpdatedBy = &userID

	if err := h.svc.Update(existing); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteTranslation removes a translation override.
func (h *TranslationHandler) DeleteTranslation(c *gin.Context) {
	id, err := parseTranslationID(c)
	if err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_request")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "not_found")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// ListGroups returns translation groups with counts.
func (h *TranslationHandler) ListGroups(c *gin.Context) {
	groups, err := h.svc.Groups()
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, groups)
}

// ImportTranslations bulk imports translations from JSON.
func (h *TranslationHandler) ImportTranslations(c *gin.Context) {
	var req struct {
		Translations []struct {
			Key    string `json:"key"`
			Locale string `json:"locale"`
			Value  string `json:"value"`
			Group  string `json:"group"`
		} `json:"translations" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	userID := c.GetString("userID")
	translations := make([]modelsCommon.Translation, len(req.Translations))
	for i, t := range req.Translations {
		translations[i] = modelsCommon.Translation{
			Key:       t.Key,
			Locale:    t.Locale,
			Value:     t.Value,
			Group:     t.Group,
			IsActive:  true,
			UpdatedBy: &userID,
		}
	}

	if err := h.svc.Import(translations); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "imported",
		"count":   len(translations),
	})
}

// ExportTranslations exports all translations as JSON.
func (h *TranslationHandler) ExportTranslations(c *gin.Context) {
	translations, err := h.svc.Export()
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "internal_error")
		return
	}

	c.JSON(http.StatusOK, translations)
}

func parseTranslationID(c *gin.Context) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
