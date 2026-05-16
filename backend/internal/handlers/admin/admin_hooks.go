package admin

import (
	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ── Hook Config CRUD ──

// AdminListHooks returns all hook configurations.
func (h *Handler) AdminListHooks(c *gin.Context) {
	if h.services == nil || h.services.HookConfig == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	hooks, err := h.services.HookConfig.List(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "hook_list_failed")
		return
	}
	if hooks == nil {
		hooks = []modelsOrder.HookConfig{}
	}
	c.JSON(http.StatusOK, hooks)
}

// AdminGetHook returns a single hook configuration.
func (h *Handler) AdminGetHook(c *gin.Context) {
	if h.services == nil || h.services.HookConfig == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_hook_id")
		return
	}
	hook, err := h.services.HookConfig.Get(c.Request.Context(), uint(id))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "hook_not_found")
		return
	}
	c.JSON(http.StatusOK, hook)
}

// AdminCreateHook creates a new hook configuration.
func (h *Handler) AdminCreateHook(c *gin.Context) {
	if h.services == nil || h.services.HookConfig == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var req struct {
		Name      string            `json:"name" binding:"required"`
		EventName string            `json:"eventName" binding:"required"`
		Type      string            `json:"type" binding:"required"`
		Config    modelsOrder.JSONMap `json:"config"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	hookType := strings.TrimSpace(req.Type)
	if hookType != "webhook" && hookType != "plugin" && hookType != "internal" {
		response.InvalidResp(c, "invalid_hook_type")
		return
	}
	hook := &modelsOrder.HookConfig{
		Name:      strings.TrimSpace(req.Name),
		EventName: strings.TrimSpace(req.EventName),
		Type:      hookType,
		Config:    req.Config,
		Status:    "active",
	}
	if err := h.services.HookConfig.Create(c.Request.Context(), hook); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "hook_create_failed")
		return
	}
	c.JSON(http.StatusCreated, hook)
}

// AdminUpdateHook updates a hook configuration.
func (h *Handler) AdminUpdateHook(c *gin.Context) {
	if h.services == nil || h.services.HookConfig == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_hook_id")
		return
	}
	hook, err := h.services.HookConfig.Get(c.Request.Context(), uint(id))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "hook_not_found")
		return
	}
	var req struct {
		Name      *string             `json:"name"`
		EventName *string             `json:"eventName"`
		Type      *string             `json:"type"`
		Config    *modelsOrder.JSONMap `json:"config"`
		Status    *string             `json:"status"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Name != nil {
		hook.Name = strings.TrimSpace(*req.Name)
	}
	if req.EventName != nil {
		hook.EventName = strings.TrimSpace(*req.EventName)
	}
	if req.Type != nil {
		t := strings.TrimSpace(*req.Type)
		if t != "webhook" && t != "plugin" && t != "internal" {
			response.InvalidResp(c, "invalid_hook_type")
			return
		}
		hook.Type = t
	}
	if req.Config != nil {
		hook.Config = *req.Config
	}
	if req.Status != nil {
		s := strings.TrimSpace(*req.Status)
		if s != "active" && s != "inactive" {
			response.InvalidResp(c, "invalid_hook_status")
			return
		}
		hook.Status = s
	}
	if err := h.services.HookConfig.Update(c.Request.Context(), hook); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "hook_update_failed")
		return
	}
	c.JSON(http.StatusOK, hook)
}

// AdminDeleteHook deletes a hook configuration.
func (h *Handler) AdminDeleteHook(c *gin.Context) {
	if h.services == nil || h.services.HookConfig == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_hook_id")
		return
	}
	if err := h.services.HookConfig.Delete(c.Request.Context(), uint(id)); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "hook_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

// ── Events ──

// AdminListEvents returns paginated events.
func (h *Handler) AdminListEvents(c *gin.Context) {
	if h.services == nil || h.services.EventBus == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	offset := (page - 1) * limit
	events, total, err := h.services.EventBus.ListEvents(c.Request.Context(), limit, offset)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "event_list_failed")
		return
	}
	if events == nil {
		events = []modelsOrder.Event{}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       events,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}

// ── Hook Executions ──

// AdminListHookExecutions returns executions for a hook.
func (h *Handler) AdminListHookExecutions(c *gin.Context) {
	if h.services == nil || h.services.HookConfig == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	hookID, err := strconv.ParseUint(c.Query("hookId"), 10, 64)
	if err != nil {
		response.InvalidResp(c, "invalid_hook_id")
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	offset := (page - 1) * limit
	executions, total, err := h.services.HookConfig.ListExecutionsByHook(c.Request.Context(), uint(hookID), limit, offset)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "hook_executions_failed")
		return
	}
	if executions == nil {
		executions = []modelsOrder.HookExecution{}
	}
	c.JSON(http.StatusOK, gin.H{
		"data":       executions,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": (total + int64(limit) - 1) / int64(limit),
	})
}
