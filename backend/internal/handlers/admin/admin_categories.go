package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/i18n"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type adminCategoryUpdateRequest struct {
	Name         *string              `json:"name"`
	Alias        *string              `json:"alias"`
	Description  *string              `json:"description"`
	Thumbnail    *string              `json:"thumbnail"`
	Icon         *string              `json:"icon"`
	Translations *modelsCommon.JSONMap `json:"translations"`
}

// AdminGetCategories 返回全部分类（含翻译）。
func (h *Handler) AdminGetCategories(c *gin.Context) {
	if h.services == nil || h.services.Category == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	categories, err := h.services.Category.GetCategories(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "category_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, categories)
}

// AdminGetCategory 按 slug 获取分类详情。
func (h *Handler) AdminGetCategory(c *gin.Context) {
	if h.services == nil || h.services.Category == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	slug := c.Param("slug")
	category, err := h.services.Category.GetCategory(c.Request.Context(), slug)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "category_not_found")
		return
	}
	c.JSON(http.StatusOK, category)
}

// AdminCreateCategory 创建分类。
func (h *Handler) AdminCreateCategory(c *gin.Context) {
	if h.services == nil || h.services.Category == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var category modelsProduct.Category
	if !response.BindJSONOrInvalid(c, &category) {
		return
	}
	category.Slug = normalizeSlug(category.Slug)
	if category.Slug == "" {
		response.InvalidResp(c, "invalid_request")
		return
	}
	syncCategoryScalarsFromLocale(&category, i18n.DefaultLocale())
	if strings.TrimSpace(category.Name) == "" {
		response.InvalidResp(c, "category_name_required")
		return
	}
	now := time.Now()
	category.CreatedAt = now
	category.UpdatedAt = now
	if err := h.services.Category.CreateCategory(c.Request.Context(), &category); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "category_slug_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "category_create_failed")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Category created", "category": category})
}

// AdminUpdateCategory 更新分类。
func (h *Handler) AdminUpdateCategory(c *gin.Context) {
	if h.services == nil || h.services.Category == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	slug := c.Param("slug")
	category, err := h.services.Category.GetCategory(c.Request.Context(), slug)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "category_not_found")
		return
	}
	var req adminCategoryUpdateRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Name != nil {
		category.Name = strings.TrimSpace(*req.Name)
	}
	if req.Alias != nil {
		category.Alias = strings.TrimSpace(*req.Alias)
	}
	if req.Description != nil {
		category.Description = strings.TrimSpace(*req.Description)
	}
	if req.Thumbnail != nil {
		category.Thumbnail = strings.TrimSpace(*req.Thumbnail)
	}
	if req.Icon != nil {
		category.Icon = strings.TrimSpace(*req.Icon)
	}
	if req.Translations != nil {
		category.Translations = *req.Translations
	}
	syncCategoryScalarsFromLocale(category, i18n.DefaultLocale())
	if category.Name == "" {
		response.InvalidResp(c, "category_name_required")
		return
	}
	category.UpdatedAt = time.Now()
	if err := h.services.Category.UpdateCategory(c.Request.Context(), category); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "category_update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category updated", "category": category})
}

// AdminDeleteCategory 删除分类。
func (h *Handler) AdminDeleteCategory(c *gin.Context) {
	if h.services == nil || h.services.Category == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	slug := c.Param("slug")
	if err := h.services.Category.DeleteCategory(c.Request.Context(), slug); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "category_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted", "slug": slug})
}

// syncCategoryScalarsFromLocale 将指定 locale 的翻译同步到分类标量列。
func syncCategoryScalarsFromLocale(category *modelsProduct.Category, locale string) {
	if category.Translations == nil {
		return
	}
	fields, ok := category.Translations[locale]
	if !ok {
		return
	}
	if v := strings.TrimSpace(fields["name"]); v != "" {
		category.Name = v
	}
	if v := strings.TrimSpace(fields["alias"]); v != "" {
		category.Alias = v
	}
	if v := strings.TrimSpace(fields["description"]); v != "" {
		category.Description = v
	}
}
