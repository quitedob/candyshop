package admin

import (
	modelsProduct "candypro/api/internal/models/product"
	modelsUser "candypro/api/internal/models/user"
	"candypro/api/internal/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminGetCompanies returns all companies with pagination.
func (h *Handler) AdminGetCompanies(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	page, limit := utils.ParsePagination(c, 20, 100)
	result, err := h.services.Company.GetCompanies(c.Request.Context(), page, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to fetch companies")
		return
	}

	c.JSON(http.StatusOK, result)
}

// AdminGetCompany returns a single company by ID.
func (h *Handler) AdminGetCompany(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	company, err := h.services.Company.GetCompany(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Company not found")
		return
	}

	c.JSON(http.StatusOK, company)
}

type adminUpdateCompanyRequest struct {
	Name            *string                `json:"name"`
	TaxID           *string                `json:"taxId"`
	RegistrationNo  *string                `json:"registrationNo"`
	BusinessLicense *string                `json:"businessLicense"`
	Address         *modelsUser.CompanyAddress `json:"address"`
	Phone           *string                `json:"phone"`
	Website         *string                `json:"website"`
	PriceListID     *string                `json:"priceListId"`
	CreditLimit     *float64               `json:"creditLimit"`
	PaymentTerms    *string                `json:"paymentTerms"`
}

// AdminUpdateCompany updates a company.
func (h *Handler) AdminUpdateCompany(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")
	company, err := h.services.Company.GetCompany(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Company not found")
		return
	}

	var req adminUpdateCompanyRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	if req.Name != nil {
		company.Name = strings.TrimSpace(*req.Name)
	}
	if req.TaxID != nil {
		company.TaxID = strings.TrimSpace(*req.TaxID)
	}
	if req.RegistrationNo != nil {
		company.RegistrationNo = strings.TrimSpace(*req.RegistrationNo)
	}
	if req.BusinessLicense != nil {
		company.BusinessLicense = strings.TrimSpace(*req.BusinessLicense)
	}
	if req.Address != nil {
		company.Address = *req.Address
	}
	if req.Phone != nil {
		company.Phone = strings.TrimSpace(*req.Phone)
	}
	if req.Website != nil {
		company.Website = strings.TrimSpace(*req.Website)
	}
	if req.PriceListID != nil {
		trimmed := strings.TrimSpace(*req.PriceListID)
		if trimmed == "" {
			company.PriceListID = nil
		} else {
			company.PriceListID = &trimmed
		}
	}
	if req.CreditLimit != nil {
		company.CreditLimit = *req.CreditLimit
	}
	if req.PaymentTerms != nil {
		company.PaymentTerms = strings.TrimSpace(*req.PaymentTerms)
	}

	company.UpdatedAt = time.Now()
	if err := h.services.Company.UpdateCompany(c.Request.Context(), company); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to update company")
		return
	}

	c.JSON(http.StatusOK, company)
}

type adminVerifyCompanyRequest struct {
	Status string `json:"status" binding:"required"` // verified or rejected
}

// AdminVerifyCompany verifies or rejects a company.
func (h *Handler) AdminVerifyCompany(c *gin.Context) {
	if h.services == nil {
		utils.ServiceUnavailableResponse(c)
		return
	}

	id := c.Param("id")

	var req adminVerifyCompanyRequest
	if !utils.BindJSONOrInvalidRequest(c, &req) {
		return
	}

	status := strings.TrimSpace(req.Status)
	if status != "verified" && status != "rejected" {
		utils.InvalidRequestResponse(c, "status must be 'verified' or 'rejected'")
		return
	}

	// Verify company exists
	if _, err := h.services.Company.GetCompany(c.Request.Context(), id); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "not_found", "Company not found")
		return
	}

	if err := h.services.Company.VerifyCompany(c.Request.Context(), id, status); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "internal_error", "Failed to verify company")
		return
	}

	c.JSON(http.StatusOK, modelsProduct.ErrorResponse{
		Error:   "success",
		Message: "Company status updated to " + status,
	})
}
