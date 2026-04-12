package customer

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) CustomerGetOEMProjects(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, _ := c.Get("userID")
	userIDStr := ""
	if id, ok := userID.(string); ok {
		userIDStr = id
	}
	page, limit := utils.ParsePagination(c, 20, 100)
	projects, total, err := h.services.OEM.GetUserProjects(c.Request.Context(), userIDStr, page, limit)
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

type customerCreateOEMProjectRequest struct {
	ProductName  string                       `json:"productName" binding:"required"`
	InquiryID    *string                      `json:"inquiryId"`
	Requirements modelsProduct.OEMRequirements `json:"requirements"`
	Notes        string                       `json:"notes"`
}

func (h *Handler) CustomerCreateOEMProject(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	userID, _ := c.Get("userID")
	userIDStr := ""
	if id, ok := userID.(string); ok {
		userIDStr = id
	}
	var req customerCreateOEMProjectRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}
	project := &modelsProduct.OEMProject{
		ID:           utils.GenerateID(),
		UserID:       userIDStr,
		InquiryID:    req.InquiryID,
		ProductName:  strings.TrimSpace(req.ProductName),
		Status:       "inquiry",
		CurrentStep:  0,
		Requirements: req.Requirements,
		Notes:        strings.TrimSpace(req.Notes),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.services.OEM.CreateProject(c.Request.Context(), project); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to create OEM project")
		return
	}
	c.JSON(http.StatusCreated, project)
}

func (h *Handler) CustomerGetOEMProject(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}
	id := c.Param("id")
	project, err := h.services.OEM.GetProject(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "OEM project not found")
		return
	}
	userID, _ := c.Get("userID")
	userIDStr := ""
	if uid, ok := userID.(string); ok {
		userIDStr = uid
	}
	if project.UserID != userIDStr {
		utils.ErrorResponse(c, http.StatusForbidden, "forbidden", "Access denied")
		return
	}
	c.JSON(http.StatusOK, project)
}
