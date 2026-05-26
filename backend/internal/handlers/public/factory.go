package public

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	_ modelsProduct.FactoryInfo
	_ modelsProduct.Certification
	_ modelsProduct.ProcessControl
	_ modelsProduct.Milestone
	_ modelsCommon.ErrorResponse
)

// ===== Factory =====

// GetFactoryInfo returns factory information
// @Summary Get factory information
// @Tags factory
// @Produce json
// @Success 200 {object} modelsProduct.FactoryInfo
// @Router /factory [get]
func (h *Handler) GetFactoryInfo(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	info, err := h.services.Factory.GetInfo(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "factory_info_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetCertifications returns all certifications
// @Summary Get certifications
// @Tags certifications
// @Produce json
// @Success 200 {array} modelsProduct.Certification
// @Router /certifications [get]
func (h *Handler) GetCertifications(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	certifications, err := h.services.Factory.GetCertifications(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "certification_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, certifications)
}

// GetCertification returns a single certification by ID
// @Summary Get certification by ID
// @Tags certifications
// @Produce json
// @Param id path string true "Certification ID"
// @Success 200 {object} modelsProduct.Certification
// @Failure 404 {object} modelsCommon.ErrorResponse
// @Router /certifications/{id} [get]
func (h *Handler) GetCertification(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")

	certification, err := h.services.Factory.GetCertificationByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "certification_not_found")
		return
	}

	c.JSON(http.StatusOK, certification)
}

// GetProcessControls returns all process controls
// @Summary Get process controls
// @Tags factory
// @Produce json
// @Success 200 {array} modelsProduct.ProcessControl
// @Router /factory/quality-controls [get]
func (h *Handler) GetProcessControls(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	controls, err := h.services.Factory.GetProcessControls(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "process_controls_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, controls)
}

// GetQualityTimeline returns quality milestones/timeline
// @Summary Get quality timeline
// @Tags factory
// @Produce json
// @Success 200 {array} modelsProduct.Milestone
// @Router /factory/timeline [get]
func (h *Handler) GetQualityTimeline(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	info, err := h.services.Factory.GetInfo(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "timeline_fetch_failed")
		return
	}

	c.JSON(http.StatusOK, info.Milestones)
}
