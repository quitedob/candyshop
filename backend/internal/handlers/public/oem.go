package public

import (
	"candypro/api/internal/utils"

	modelsProduct "candypro/api/internal/models/product"
)

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ===== OEM =====

// GetOEMFlows returns all OEM flows
// @Summary Get OEM flows
// @Tags oem
// @Produce json
// @Success 200 {array} modelsProduct.OEMFlow
// @Router /oem/flows [get]
func (h *Handler) GetOEMFlows(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	flows, err := h.services.OEM.GetFlows(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch OEM flows",
		})
		return
	}

	c.JSON(http.StatusOK, flows)
}

// GetOEMFlow returns a single OEM flow by ID
// @Summary Get OEM flow by ID
// @Tags oem
// @Produce json
// @Param id path string true "OEM Flow ID"
// @Success 200 {object} modelsProduct.OEMFlow
// @Failure 404 {object} modelsProduct.ErrorResponse
// @Router /oem/flows/{id} [get]
func (h *Handler) GetOEMFlow(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")

	flow, err := h.services.OEM.GetFlowByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "OEM flow not found",
		})
		return
	}

	c.JSON(http.StatusOK, flow)
}

// GetOEMSolutions returns all OEM solutions
// @Summary Get OEM solutions
// @Tags oem
// @Produce json
// @Success 200 {array} modelsProduct.OEMSolution
// @Router /oem/solutions [get]
func (h *Handler) GetOEMSolutions(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	solutions, err := h.services.OEM.GetSolutions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch OEM solutions",
		})
		return
	}

	c.JSON(http.StatusOK, solutions)
}

// GetOEMSolution returns a single OEM solution by slug
// @Summary Get OEM solution by slug
// @Tags oem
// @Produce json
// @Param slug path string true "OEM Solution slug"
// @Success 200 {object} modelsProduct.OEMSolution
// @Failure 404 {object} modelsProduct.ErrorResponse
// @Router /oem/solutions/{slug} [get]
func (h *Handler) GetOEMSolution(c *gin.Context) {
	if !(h.services != nil) {
		utils.ServiceUnavailableResponse(c)
		return
	}

	slug := c.Param("slug")

	solution, err := h.services.OEM.GetSolutionBySlug(c.Request.Context(), slug)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "OEM solution not found",
		})
		return
	}

	c.JSON(http.StatusOK, solution)
}
