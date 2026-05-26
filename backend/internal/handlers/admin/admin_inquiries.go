package admin

import (
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
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
		response.ServiceUnavailableResp(c)
		return
	}

	page, limit := pagination.ParsePagination(c, 20, 100)
	inquiries, total, err := h.services.Inquiry.GetInquiries(c.Request.Context(), page, limit)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       inquiries,
		"pagination": pagination.BuildPagination(total, page, limit),
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

	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}
	previousStatus := inquiry.Status

	if err := h.services.Inquiry.UpdateInquiryStatus(c.Request.Context(), id, req.Status); err != nil {
		response.InvalidResp(c, "invalid_request")
		return
	}

	if previousStatus != req.Status {
		h.logInquiryAudit(c, id, c.GetString("userID"), previousStatus, req.Status)
	}
	inquiry.Status = req.Status
	c.JSON(http.StatusOK, inquiry)
}
