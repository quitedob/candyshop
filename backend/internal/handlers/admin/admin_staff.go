package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetStaffList returns paginated admin/superadmin users.
func (h *Handler) GetStaffList(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	page, limit := utils.ParsePagination(c, 20, 100)

	users, total, err := h.services.User.GetUsers(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "staff_fetch_failed")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: users,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// GetStaffActivity returns activity logs for a specific user.
func (h *Handler) GetStaffActivity(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	userID := c.Param("id")
	if strings.TrimSpace(userID) == "" {
		utils.InvalidResp(c, "invalid_request")
		return
	}
	page, limit := utils.ParsePagination(c, 20, 100)

	logs, total, err := h.services.ActivityLog.FindByUser(c.Request.Context(), userID, page, limit)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "staff_fetch_failed")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: logs,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}

// GetAuditLog returns paginated activity log entries with optional filters.
func (h *Handler) GetAuditLog(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResp(c)
		return
	}
	page, limit := utils.ParsePagination(c, 20, 100)

	logs, total, err := h.services.ActivityLog.FindAll(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResp(c, http.StatusInternalServerError, "audit_log_fetch_failed")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: logs,
		Pagination: modelsProduct.Pagination{
			Total:      int(total),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	})
}
