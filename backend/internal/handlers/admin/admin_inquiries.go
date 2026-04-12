package admin

import (
	"candypro/api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AdminGetInquiries returns inquiries for admins.
// @Summary Admin get inquiries
// @Tags admin-inquiries
// @Produce json
// @Router /admin/inquiries [get]
func (h *Handler) AdminGetInquiries(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	page, limit := utils.ParsePagination(c, 20, 100)
	inquiries, total, err := h.services.Inquiry.GetInquiries(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch inquiries")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       inquiries,
		"pagination": utils.BuildPagination(total, page, limit),
	})
}

// AdminUpdateInquiryStatus updates an inquiry status.
// @Summary Admin update inquiry status
// @Tags admin-inquiries
// @Produce json
// @Param id path string true "Inquiry ID"
// @Param request body map[string]interface{} true "Inquiry status"
// @Router /admin/inquiries/{id}/status [put]
func (h *Handler) AdminUpdateInquiryStatus(c *gin.Context) {
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

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Inquiry not found")
		return
	}

	if err := h.services.Inquiry.UpdateInquiryStatus(c.Request.Context(), id, req.Status); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	inquiry.Status = req.Status
	c.JSON(http.StatusOK, inquiry)
}
