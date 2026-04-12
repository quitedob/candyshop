package admin

import (
	"candypro/api/internal/utils"
	"time"

	modelsProduct "candypro/api/internal/models/product"

	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type adminCertificationRequest struct {
	Name           *string `json:"name"`
	Abbreviation   *string `json:"abbreviation"`
	Description    *string `json:"description"`
	Issuer         *string `json:"issuer"`
	ValidUntil     *string `json:"validUntil"`
	CertificateURL *string `json:"certificateUrl"`
	BadgeURL       *string `json:"badgeUrl"`
}

// AdminGetCertifications returns all certifications
// @Summary Admin get certifications
// @Tags admin-certifications
// @Produce json
// @Router /admin/certifications [get]
func (h *Handler) AdminGetCertifications(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	certifications, err := h.services.Factory.GetCertifications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to fetch certifications",
		})
		return
	}

	c.JSON(http.StatusOK, certifications)
}

// AdminGetCertification returns a single certification
// @Summary Admin get certification by ID
// @Tags admin-certifications
// @Produce json
// @Param id path string true "Certification ID"
// @Router /admin/certifications/{id} [get]
func (h *Handler) AdminGetCertification(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	certification, err := h.services.Factory.GetCertificationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Certification not found",
		})
		return
	}

	c.JSON(http.StatusOK, certification)
}

// AdminCreateCertification creates a new certification
// @Summary Admin create certification
// @Tags admin-certifications
// @Accept json
// @Produce json
// @Router /admin/certifications [post]
func (h *Handler) AdminCreateCertification(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	var req modelsProduct.Certification
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: "name is required",
		})
		return
	}

	if strings.TrimSpace(req.ID) == "" {
		req.ID = utils.GenerateID()
	}

	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now

	if err := h.services.Factory.CreateCertification(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to create certification",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":       "Certification created successfully",
		"certification": req,
	})
}

// AdminUpdateCertification updates a certification
// @Summary Admin update certification
// @Tags admin-certifications
// @Accept json
// @Produce json
// @Param id path string true "Certification ID"
// @Router /admin/certifications/{id} [put]
func (h *Handler) AdminUpdateCertification(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	certification, err := h.services.Factory.GetCertificationByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Certification not found",
		})
		return
	}

	var req adminCertificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, modelsProduct.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	if req.Name != nil {
		certification.Name = strings.TrimSpace(*req.Name)
	}
	if req.Abbreviation != nil {
		certification.Abbreviation = strings.TrimSpace(*req.Abbreviation)
	}
	if req.Description != nil {
		certification.Description = strings.TrimSpace(*req.Description)
	}
	if req.Issuer != nil {
		certification.Issuer = strings.TrimSpace(*req.Issuer)
	}
	if req.ValidUntil != nil {
		certification.ValidUntil = strings.TrimSpace(*req.ValidUntil)
	}
	if req.CertificateURL != nil {
		certification.CertificateURL = strings.TrimSpace(*req.CertificateURL)
	}
	if req.BadgeURL != nil {
		certification.BadgeURL = strings.TrimSpace(*req.BadgeURL)
	}

	certification.UpdatedAt = time.Now()

	if err := h.services.Factory.UpdateCertification(c.Request.Context(), certification); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to update certification",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Certification updated successfully",
		"certification": certification,
	})
}

// AdminDeleteCertification deletes a certification
// @Summary Admin delete certification
// @Tags admin-certifications
// @Produce json
// @Param id path string true "Certification ID"
// @Router /admin/certifications/{id} [delete]
func (h *Handler) AdminDeleteCertification(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")

	if _, err := h.services.Factory.GetCertificationByID(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, modelsProduct.ErrorResponse{
			Error:   "not_found",
			Message: "Certification not found",
		})
		return
	}

	if err := h.services.Factory.DeleteCertification(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, modelsProduct.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to delete certification",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certification deleted successfully",
		"id":      id,
	})
}
