package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/dberror"
	"candypro/api/internal/pkg/pagination"
	"candypro/api/internal/pkg/response"
	oemSvc "candypro/api/internal/services/oem"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminGetOEMProjects(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	page, limit := pagination.ParsePagination(c, 20, 100)
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	projects, total, err := h.services.Project.GetProjects(c.Request.Context(), page, limit, search, status)
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_project_fetch_failed")
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
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	project, err := h.services.Project.GetProject(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_project_not_found")
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
	AdminNotes   *string                        `json:"adminNotes"`
}

func (h *Handler) AdminUpdateOEMProject(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	project, err := h.services.Project.GetProject(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_project_not_found")
		return
	}
	var req adminUpdateOEMProjectRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	// R2 A-4 / G-OEM-4: route status changes through the guarded transition so
	// this PUT path can't bypass the optimistic-concurrency check that
	// UpdateProjectStatus enforces. A blind Save here would let two concurrent
	// admins silently overwrite each other's status change.
	if req.Status != nil {
		newStatus := strings.TrimSpace(*req.Status)
		if newStatus != "" && newStatus != project.Status {
			if serr := h.services.Project.UpdateProjectStatus(c.Request.Context(), id, newStatus); serr != nil {
				if errors.Is(serr, oemSvc.ErrOEMStatusConflict) {
					response.ErrorResp(c, http.StatusConflict, "oem_status_conflict")
					return
				}
				response.InvalidResp(c, "oem_status_transition_invalid")
				return
			}
			project, err = h.services.Project.GetProject(c.Request.Context(), id)
			if err != nil {
				response.ErrorResp(c, http.StatusNotFound, "oem_project_not_found")
				return
			}
		}
	}
	if req.ProductName != nil {
		project.ProductName = strings.TrimSpace(*req.ProductName)
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
	if req.AdminNotes != nil {
		project.AdminNotes = strings.TrimSpace(*req.AdminNotes)
	}
	project.UpdatedAt = time.Now()
	if err := h.services.Project.UpdateProject(c.Request.Context(), project); err != nil {
		if errors.Is(err, oemSvc.ErrOEMProjectConflict) {
			response.ErrorResp(c, http.StatusConflict, "oem_project_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "oem_project_update_failed")
		return
	}
	c.JSON(http.StatusOK, project)
}

func (h *Handler) AdminUpdateOEMStatus(c *gin.Context) {
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
	if err := h.services.Project.UpdateProjectStatus(c.Request.Context(), id, strings.TrimSpace(req.Status)); err != nil {
		if errors.Is(err, oemSvc.ErrOEMStatusConflict) {
			// R2 A-4: surface the lost-race as 409 so the admin UI can refresh
			// and let the user re-decide instead of silently overwriting.
			response.ErrorResp(c, http.StatusConflict, "oem_status_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "oem_status_update_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"error": "success", "message": "OEM project status updated"})
}

// ===== OEM Flows =====

func (h *Handler) AdminGetOEMFlows(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	flows, err := h.services.OEM.GetFlows(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_flows_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, flows)
}

func (h *Handler) AdminGetOEMFlow(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	flow, err := h.services.OEM.GetFlowByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_flow_not_found")
		return
	}
	c.JSON(http.StatusOK, flow)
}

func (h *Handler) AdminCreateOEMFlow(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var flow modelsProduct.OEMFlow
	if !response.BindJSONOrInvalid(c, &flow) {
		return
	}
	flow.Title = strings.TrimSpace(flow.Title)
	if flow.Title == "" {
		response.InvalidResp(c, "oem_flow_title_required")
		return
	}
	if flow.ID == "" {
		flow.ID = crypto.GenerateID()
	}
	flow.CreatedAt = time.Now()
	flow.UpdatedAt = time.Now()
	if err := h.services.OEM.CreateFlow(c.Request.Context(), &flow); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_flow_create_failed")
		return
	}
	c.JSON(http.StatusCreated, flow)
}

func (h *Handler) AdminUpdateOEMFlow(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	flow, err := h.services.OEM.GetFlowByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_flow_not_found")
		return
	}
	var req struct {
		Title       *string                    `json:"title"`
		Description *string                    `json:"description"`
		Type        *string                    `json:"type"`
		Steps       *modelsCommon.OEMStepArray `json:"steps"`
		Timeline    *string                    `json:"timeline"`
		MOQ         *int                       `json:"moq"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Title != nil {
		flow.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		flow.Description = strings.TrimSpace(*req.Description)
	}
	if req.Type != nil {
		flow.Type = strings.TrimSpace(*req.Type)
	}
	if req.Steps != nil {
		flow.Steps = *req.Steps
	}
	if req.Timeline != nil {
		flow.Timeline = strings.TrimSpace(*req.Timeline)
	}
	if req.MOQ != nil {
		flow.MOQ = *req.MOQ
	}
	flow.UpdatedAt = time.Now()
	if err := h.services.OEM.UpdateFlow(c.Request.Context(), flow); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_flow_update_failed")
		return
	}
	c.JSON(http.StatusOK, flow)
}

func (h *Handler) AdminDeleteOEMFlow(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if err := h.services.OEM.DeleteFlow(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_flow_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OEM flow deleted"})
}

// ===== OEM Solutions =====

func (h *Handler) AdminGetOEMSolutions(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	solutions, err := h.services.OEM.GetSolutions(c.Request.Context())
	if err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_solutions_fetch_failed")
		return
	}
	c.JSON(http.StatusOK, solutions)
}

func (h *Handler) AdminGetOEMSolution(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	solution, err := h.services.OEM.GetSolutionBySlug(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_solution_not_found")
		return
	}
	c.JSON(http.StatusOK, solution)
}

func (h *Handler) AdminCreateOEMSolution(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	var solution modelsProduct.OEMSolution
	if !response.BindJSONOrInvalid(c, &solution) {
		return
	}
	solution.Title = strings.TrimSpace(solution.Title)
	if solution.Title == "" {
		response.InvalidResp(c, "oem_solution_title_required")
		return
	}
	if solution.ID == "" {
		solution.ID = crypto.GenerateID()
	}
	if solution.Slug == "" {
		solution.Slug = buildProductSlug(solution.Title)
	}
	solution.CreatedAt = time.Now()
	solution.UpdatedAt = time.Now()
	if err := h.services.OEM.CreateSolution(c.Request.Context(), &solution); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "oem_solution_slug_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "oem_solution_create_failed")
		return
	}
	c.JSON(http.StatusCreated, solution)
}

func (h *Handler) AdminUpdateOEMSolution(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	solution, err := h.services.OEM.GetSolutionBySlug(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "oem_solution_not_found")
		return
	}
	var req struct {
		Slug         *string   `json:"slug"`
		Title        *string   `json:"title"`
		Description  *string   `json:"description"`
		Thumbnail    *string   `json:"thumbnail"`
		Images       *[]string `json:"images"`
		Category     *string   `json:"category"`
		MOQ          *int      `json:"moq"`
		Applications *[]string `json:"applications"`
	}
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}
	if req.Slug != nil {
		solution.Slug = normalizeSlug(*req.Slug)
	}
	if req.Title != nil {
		solution.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		solution.Description = strings.TrimSpace(*req.Description)
	}
	if req.Thumbnail != nil {
		solution.Thumbnail = strings.TrimSpace(*req.Thumbnail)
	}
	if req.Images != nil {
		solution.Images = modelsCommon.StringArray(*req.Images)
	}
	if req.Category != nil {
		solution.Category = strings.TrimSpace(*req.Category)
	}
	if req.MOQ != nil {
		solution.MOQ = *req.MOQ
	}
	if req.Applications != nil {
		solution.Applications = modelsCommon.StringArray(*req.Applications)
	}
	solution.UpdatedAt = time.Now()
	if err := h.services.OEM.UpdateSolution(c.Request.Context(), solution); err != nil {
		if dberror.IsDuplicateKeyError(err) {
			response.ErrorResp(c, http.StatusConflict, "oem_solution_slug_conflict")
			return
		}
		response.ErrorResp(c, http.StatusInternalServerError, "oem_solution_update_failed")
		return
	}
	c.JSON(http.StatusOK, solution)
}

func (h *Handler) AdminDeleteOEMSolution(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}
	id := c.Param("id")
	if err := h.services.OEM.DeleteSolution(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "oem_solution_delete_failed")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "OEM solution deleted"})
}
