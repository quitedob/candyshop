package admin

import (
	"candypro/api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminGetUsers returns all users.
// @Summary Admin get users
// @Tags admin-users
// @Produce json
// @Router /admin/users [get]
func (h *Handler) AdminGetUsers(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	page, limit := utils.ParsePagination(c, 20, 100)
	users, total, err := h.services.User.GetUsers(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch users")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       users,
		"pagination": utils.BuildPagination(total, page, limit),
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
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
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
		return
	}

	user.Status = req.Status
	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update user status")
		return
	}

	c.JSON(http.StatusOK, user)
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
		utils.ServiceUnavailableResponse(c)
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
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	user, err := h.services.User.GetByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "User not found")
		return
	}

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
		user.RoleID = req.RoleID
	}

	if err := h.services.User.UpdateUser(c.Request.Context(), user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update user")
		return
	}

	c.JSON(http.StatusOK, user)
}
