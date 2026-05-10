package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"
	"time"

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

// AdminGetCertification returns a single certification
// @Summary Admin get certification by ID
// @Tags admin-certifications
// @Produce json
// @Param id path string true "Certification ID"
// @Router /admin/certifications/{id} [get]
func (h *Handler) AdminGetCertification(c *gin.Context) {
	if h.services == nil {
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

// AdminCreateCertification creates a new certification
// @Summary Admin create certification
// @Tags admin-certifications
// @Accept json
// @Produce json
// @Router /admin/certifications [post]
func (h *Handler) AdminCreateCertification(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req modelsProduct.Certification
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		response.InvalidResp(c, "certification_name_required")
		return
	}

	if strings.TrimSpace(req.ID) == "" {
		req.ID = crypto.GenerateID()
	}

	now := time.Now()
	req.CreatedAt = now
	req.UpdatedAt = now

	if err := h.services.Factory.CreateCertification(c.Request.Context(), &req); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "certification_create_failed")
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
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")
	certification, err := h.services.Factory.GetCertificationByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "certification_not_found")
		return
	}

	var req adminCertificationRequest
	if !response.BindJSONOrInvalid(c, &req) {
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
		response.ErrorResp(c, http.StatusInternalServerError, "certification_update_failed")
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
		response.ServiceUnavailableResp(c)
		return
	}

	id := c.Param("id")

	if _, err := h.services.Factory.GetCertificationByID(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "certification_not_found")
		return
	}

	if err := h.services.Factory.DeleteCertification(c.Request.Context(), id); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "certification_delete_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certification deleted successfully",
		"id":      id,
	})
}
