package admin

import (
	modelsCommon "candypro/api/internal/models/common"
	modelsProduct "candypro/api/internal/models/product"
	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/response"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type adminCreateInquiryRequest struct {
	UserID                *string  `json:"userId"`
	CompanyName           string   `json:"companyName" binding:"required"`
	ContactPerson         string   `json:"contactPerson" binding:"required"`
	Email                 string   `json:"email" binding:"required,email"`
	WhatsApp              string   `json:"whatsapp"`
	TargetCountry         string   `json:"targetCountry"`
	EstimatedQuantity     string   `json:"estimatedQuantity"`
	InterestedProducts    []string `json:"interestedProducts"`
	ProductIDs            []string `json:"productIds"`
	PackagingRequirements string   `json:"packagingRequirements"`
	FlavorRequirements    string   `json:"flavorRequirements"`
	OEMNeeded             bool     `json:"oemNeeded"`
	ExpectedDelivery      string   `json:"expectedDelivery"`
	Message               string   `json:"message"`
	Status                string   `json:"status"`
	Priority              string   `json:"priority"`
	AssignedTo            *string  `json:"assignedTo"`
}

type adminUpdateInquiryRequest struct {
	UserID                *string   `json:"userId"`
	CompanyName           *string   `json:"companyName"`
	ContactPerson         *string   `json:"contactPerson"`
	Email                 *string   `json:"email"`
	WhatsApp              *string   `json:"whatsapp"`
	TargetCountry         *string   `json:"targetCountry"`
	EstimatedQuantity     *string   `json:"estimatedQuantity"`
	InterestedProducts    *[]string `json:"interestedProducts"`
	ProductIDs            *[]string `json:"productIds"`
	PackagingRequirements *string   `json:"packagingRequirements"`
	FlavorRequirements    *string   `json:"flavorRequirements"`
	OEMNeeded             *bool     `json:"oemNeeded"`
	ExpectedDelivery      *string   `json:"expectedDelivery"`
	Message               *string   `json:"message"`
	Status                *string   `json:"status"`
	Priority              *string   `json:"priority"`
	AssignedTo            *string   `json:"assignedTo"`
	CustomerNotes         *string   `json:"customerNotes"`
}

// AdminCreateInquiry creates a new inquiry.
func (h *Handler) AdminCreateInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	var req adminCreateInquiryRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	var userID *string
	if req.UserID != nil && strings.TrimSpace(*req.UserID) != "" {
		trimmed := strings.TrimSpace(*req.UserID)
		if _, err := h.services.User.GetByID(c.Request.Context(), trimmed); err != nil {
			response.ErrorResp(c, http.StatusNotFound, "user_not_found")
			return
		}
		userID = &trimmed
	}

	inquiry := &modelsProduct.Inquiry{
		ID:                    crypto.GenerateID(),
		UserID:                userID,
		CompanyName:           strings.TrimSpace(req.CompanyName),
		ContactPerson:         strings.TrimSpace(req.ContactPerson),
		Email:                 strings.TrimSpace(req.Email),
		WhatsApp:              strings.TrimSpace(req.WhatsApp),
		TargetCountry:         strings.TrimSpace(req.TargetCountry),
		EstimatedQuantity:     strings.TrimSpace(req.EstimatedQuantity),
		InterestedProducts:    modelsCommon.StringArray(req.InterestedProducts),
		ProductIDs:            modelsCommon.StringArray(req.ProductIDs),
		PackagingRequirements: strings.TrimSpace(req.PackagingRequirements),
		FlavorRequirements:    strings.TrimSpace(req.FlavorRequirements),
		OEMNeeded:             req.OEMNeeded,
		ExpectedDelivery:      strings.TrimSpace(req.ExpectedDelivery),
		Message:               strings.TrimSpace(req.Message),
		Status:                strings.TrimSpace(req.Status),
		Priority:              strings.TrimSpace(req.Priority),
	}

	if inquiry.Status == "" {
		inquiry.Status = "pending"
	}
	if inquiry.Priority == "" {
		inquiry.Priority = "normal"
	}
	if req.AssignedTo != nil && strings.TrimSpace(*req.AssignedTo) != "" {
		assigned := strings.TrimSpace(*req.AssignedTo)
		inquiry.AssignedTo = &assigned
	}

	if err := h.services.Inquiry.CreateInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_create_failed")
		return
	}

	c.JSON(http.StatusCreated, inquiry)
}

// AdminUpdateInquiry updates inquiry fields.
func (h *Handler) AdminUpdateInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	inquiryID := c.Param("id")
	inquiry, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID)
	if err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}
	previousStatus := inquiry.Status

	var req adminUpdateInquiryRequest
	if !response.BindJSONOrInvalid(c, &req) {
		return
	}

	if req.UserID != nil {
		if strings.TrimSpace(*req.UserID) == "" {
			inquiry.UserID = nil
		} else {
			uid := strings.TrimSpace(*req.UserID)
			if _, userErr := h.services.User.GetByID(c.Request.Context(), uid); userErr != nil {
				response.ErrorResp(c, http.StatusNotFound, "user_not_found")
				return
			}
			inquiry.UserID = &uid
		}
	}

	if req.CompanyName != nil {
		inquiry.CompanyName = strings.TrimSpace(*req.CompanyName)
	}
	if req.ContactPerson != nil {
		inquiry.ContactPerson = strings.TrimSpace(*req.ContactPerson)
	}
	if req.Email != nil {
		inquiry.Email = strings.TrimSpace(*req.Email)
	}
	if req.WhatsApp != nil {
		inquiry.WhatsApp = strings.TrimSpace(*req.WhatsApp)
	}
	if req.TargetCountry != nil {
		inquiry.TargetCountry = strings.TrimSpace(*req.TargetCountry)
	}
	if req.EstimatedQuantity != nil {
		inquiry.EstimatedQuantity = strings.TrimSpace(*req.EstimatedQuantity)
	}
	if req.InterestedProducts != nil {
		inquiry.InterestedProducts = modelsCommon.StringArray(*req.InterestedProducts)
	}
	if req.ProductIDs != nil {
		inquiry.ProductIDs = modelsCommon.StringArray(*req.ProductIDs)
	}
	if req.PackagingRequirements != nil {
		inquiry.PackagingRequirements = strings.TrimSpace(*req.PackagingRequirements)
	}
	if req.FlavorRequirements != nil {
		inquiry.FlavorRequirements = strings.TrimSpace(*req.FlavorRequirements)
	}
	if req.OEMNeeded != nil {
		inquiry.OEMNeeded = *req.OEMNeeded
	}
	if req.ExpectedDelivery != nil {
		inquiry.ExpectedDelivery = strings.TrimSpace(*req.ExpectedDelivery)
	}
	if req.Message != nil {
		inquiry.Message = strings.TrimSpace(*req.Message)
	}
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if err := h.services.Inquiry.ValidateInquiryStatus(st); err != nil {
			response.InvalidResp(c, "invalid_request")
			return
		}
		if err := h.services.Inquiry.ValidateInquiryStatusTransition(inquiry.Status, st); err != nil {
			response.InvalidResp(c, "invalid_request")
			return
		}
		inquiry.Status = st
	}
	if req.Priority != nil {
		inquiry.Priority = strings.TrimSpace(*req.Priority)
	}
	if req.AssignedTo != nil {
		assigned := strings.TrimSpace(*req.AssignedTo)
		if assigned == "" {
			inquiry.AssignedTo = nil
		} else {
			inquiry.AssignedTo = &assigned
		}
	}
	if req.CustomerNotes != nil {
		inquiry.CustomerNotes = strings.TrimSpace(*req.CustomerNotes)
	}

	inquiry.UpdatedAt = time.Now()
	if err := h.services.Inquiry.UpdateInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_update_failed")
		return
	}

	if previousStatus != inquiry.Status {
		h.logInquiryAudit(c, inquiryID, c.GetString("userID"), previousStatus, inquiry.Status)
	}

	c.JSON(http.StatusOK, inquiry)
}

// AdminDeleteInquiry deletes an inquiry.
func (h *Handler) AdminDeleteInquiry(c *gin.Context) {
	if h.services == nil {
		response.ServiceUnavailableResp(c)
		return
	}

	inquiryID := c.Param("id")
	if _, err := h.services.Inquiry.GetInquiry(c.Request.Context(), inquiryID); err != nil {
		response.ErrorResp(c, http.StatusNotFound, "inquiry_not_found")
		return
	}

	if err := h.services.Inquiry.DeleteInquiry(c.Request.Context(), inquiryID); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_delete_failed")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Inquiry deleted successfully",
		"id":      inquiryID,
	})
}
