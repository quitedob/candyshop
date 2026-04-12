package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminGetOEMProjects(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	page, limit := utils.ParsePagination(c, 20, 100)
	projects, total, err := h.services.Project.GetProjects(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch OEM projects")
		return
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	c.JSON(http.StatusOK, modelsProduct.PaginatedResponse{
		Data: projects,
		Pagination: modelsProduct.Pagination{
			Total: int(total), Page: page, Limit: limit, TotalPages: totalPages,
		},
	})
}

func (h *Handler) AdminGetOEMProject(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id := c.Param("id")
	project, err := h.services.Project.GetProject(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "OEM project not found")
		return
	}
	c.JSON(http.StatusOK, project)
}

type adminUpdateOEMProjectRequest struct {
	ProductName  *string                        `json:"productName"`
	Status       *string                        `json:"status"`
	CurrentStep  *int                           `json:"currentStep"`
	Requirements *modelsProduct.OEMRequirements `json:"requirements"`
	AssignedTo   *string                        `json:"assignedTo"`
	Notes        *string                        `json:"notes"`
}

func (h *Handler) AdminUpdateOEMProject(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id := c.Param("id")
	project, err := h.services.Project.GetProject(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "OEM project not found")
		return
	}
	var req adminUpdateOEMProjectRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	if req.ProductName != nil {
		project.ProductName = strings.TrimSpace(*req.ProductName)
	}
	if req.Status != nil {
		project.Status = strings.TrimSpace(*req.Status)
	}
	if req.CurrentStep != nil {
		project.CurrentStep = *req.CurrentStep
	}
	if req.Requirements != nil {
		project.Requirements = *req.Requirements
	}
	if req.AssignedTo != nil {
		assigned := strings.TrimSpace(*req.AssignedTo)
		if assigned == "" {
			project.AssignedTo = nil
		} else {
			project.AssignedTo = &assigned
		}
	}
	if req.Notes != nil {
		project.Notes = strings.TrimSpace(*req.Notes)
	}
	project.UpdatedAt = time.Now()
	if err := h.services.Project.UpdateProject(c.Request.Context(), project); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update OEM project")
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *Handler) AdminUpdateOEMStatus(c *gin.Context) {
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
	if err := h.services.Project.UpdateProjectStatus(c.Request.Context(), id, strings.TrimSpace(req.Status)); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update OEM status")
		return
	}
	c.JSON(http.StatusOK, modelsProduct.ErrorResponse{Error: "success", Message: "OEM project status updated"})
}
