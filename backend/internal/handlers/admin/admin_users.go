package admin

import (
	modelsAuth "candypro/api/internal/models/auth"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminGetUsers returns all users.
// @Summary Admin get users
// @Tags admin-users
// @Produce json
// @Router /admin/users [get]
func (h *Handler) AdminGetUsers(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	users, total, err := h.services.User.GetUsers(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       users,
		"pagination": pagination.BuildPagination(total, page, limit),
	})
}

// AdminGetUser returns a user.
// @Summary Admin get user
// @Tags admin-users
// @Produce json
// @Param id path string true "User ID"
// @Router /admin/users/{id} [get]
func (h *Handler) AdminGetUser(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}

	c.JSON(http.StatusOK, user)
}

// AdminUpdateUserStatus updates a user status.
// @Summary Admin update user status
// @Tags admin-users
// @Produce json
// @Param id path string true "User ID"
// @Param request body map[string]interface{} true "User status"
// @Router /admin/users/{id}/status [put]
func (h *Handler) AdminUpdateUserStatus(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}
	oldStatus := user.Status

	user.Status = req.Status
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_status_update_failed")
		return
	}

	h.logActivityAudit(c, "status_change", "user", id, oldStatus, req.Status)
	c.JSON(http.StatusOK, user)
}

// authorizeRoleGrant enforces the invariant that only a superadmin may assign
// the admin or superadmin roles. It writes a 403 and returns false when the
// caller is not a superadmin and the target role is privileged; otherwise it
// returns true and leaves the caller's own role untouched.
func (h *Handler) authorizeRoleGrant(c *gin.Context, target *modelsAuth.Role) bool {
	if h.services == nil || h.services.Auth == nil {
		response.ErrorResp(c, http.StatusForbidden, "forbidden_insufficient_role")
		return false
	}
	callerID, exists := c.Get("userID")
	uid, _ := callerID.(string)
	if !exists || strings.TrimSpace(uid) == "" {
		response.ErrorResp(c, http.StatusForbidden, "forbidden_insufficient_role")
		return false
	}
	if err := h.services.Auth.CanGrantRole(c.Request.Context(), uid, target); err != nil {
		response.ErrorResp(c, http.StatusForbidden, "forbidden_insufficient_role")
		return false
	}
	return true
}

// AdminUpdateUser updates a user profile.
// @Summary Admin update user
// @Tags admin-users
// @Produce json
// @Param id path string true "User ID"
// @Param request body map[string]interface{} true "User data"
// @Router /admin/users/{id} [put]
func (h *Handler) AdminUpdateUser(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	var req struct {
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Company   string `json:"company"`
		Phone     string `json:"phone"`
		RoleID    string `json:"roleId"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "user_not_found")
		return
	}
	oldRoleID := user.RoleID

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Company != "" {
		user.Company = req.Company
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.RoleID != "" {
		resolvedRole, roleErr := h.services.Auth.ResolveRole(c.Request.Context(), req.RoleID, "")
		if roleErr != nil {
			response.ErrorResp(c, http.StatusBadRequest, "invalid_role")
			return
		}
		// Only a superadmin may promote a user to admin/superadmin.
		if !h.authorizeRoleGrant(c, resolvedRole) {
			return
		}
		user.RoleID = resolvedRole.ID
		// Drop the stale preloaded Role snapshot so GORM Save cannot revert the
		// foreign key back to the previous role.
		user.Role = nil
	}

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "user_update_failed")
		return
	}

	if req.RoleID != "" && req.RoleID != oldRoleID {
		h.logActivityAudit(c, "update", "user", id, oldRoleID, req.RoleID)
	}

	c.JSON(http.StatusOK, user)
}
