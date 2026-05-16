package public

import (
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	"candypro/api/internal/pkg/crypto"
	"candypro/api/internal/pkg/response"
	"candypro/api/internal/pkg/storage"

	"github.com/gin-gonic/gin"

	modelsProduct "candypro/api/internal/models/product"
)

// ===== Inquiry =====

// SubmitInquiry handles inquiry form submissions
// @Summary Submit inquiry
// @Tags inquiry
// @Accept multipart/form-data
// @Produce json
// @Param companyName formData string true "Company name"
// @Param contactPerson formData string true "Contact person"
// @Param email formData string true "Email address"
// @Param whatsapp formData string false "WhatsApp number"
// @Param targetCountry formData string false "Target country"
// @Param estimatedQuantity formData string false "Estimated quantity"
// @Param interestedProducts formData []string false "Interested products"
// @Param packagingRequirements formData string false "Packaging requirements"
// @Param flavorRequirements formData string false "Flavor requirements"
// @Param oemNeeded formData boolean false "OEM needed"
// @Param expectedDelivery formData string false "Expected delivery"
// @Param message formData string false "Message"
// @Param files formData file false "Attached files"
// @Success 200 {object} map[string]interface{}
// @Router /inquiry [post]
func (h *Handler) SubmitInquiry(c *gin.Context) {
	if !(h.services != nil) {
		response.ServiceUnavailableResp(c)
		return
	}

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		response.ErrorResp(c, http.StatusBadRequest, "form_parse_failed")
		return
	}

	form := c.Request.MultipartForm

	// Extract and validate required fields
	companyName := getFormValue(form, "companyName")
	contactPerson := getFormValue(form, "contactPerson")
	email := getFormValue(form, "email")

	// Validate required fields
	if companyName == "" || contactPerson == "" || email == "" {
		response.ErrorResp(c, http.StatusBadRequest, "inquiry_fields_required")
		return
	}

	// Validate email format
	emailRegex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)
	if !matched {
		response.ErrorResp(c, http.StatusBadRequest, "invalid_email_format")
		return
	}

	// Extract optional fields
	interestedProducts := getFormValues(form, "interestedProducts")
	oemNeeded := getFormValue(form, "oemNeeded") == "true"

	// Handle file uploads
	var files []string
	if fileHeaders, ok := form.File["files"]; ok {
		for _, fileHeader := range fileHeaders {
			if fileHeader.Size > h.cfg.Upload.MaxFileSize {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_size_exceeded", gin.H{
					"filename": fileHeader.Filename,
					"maxBytes": h.cfg.Upload.MaxFileSize,
				})
				return
			}

			contentType := fileHeader.Header.Get("Content-Type")
			if !isAllowedType(contentType, h.cfg.Upload.AllowedTypes) {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_type_not_allowed", gin.H{
					"contentType": contentType,
				})
				return
			}

			f, fErr := fileHeader.Open()
			if fErr != nil {
				response.ErrorRespDetail(c, http.StatusBadRequest, "file_open_failed", gin.H{
					"filename": fileHeader.Filename,
				})
				return
			}

			url, uploadErr := h.storage.Upload(c.Request.Context(), f, storage.UploadOptions{
				FileName: fileHeader.Filename,
			})
			f.Close()
			if uploadErr != nil {
				response.ErrorRespDetail(c, http.StatusInternalServerError, "file_save_failed", gin.H{
					"filename": fileHeader.Filename,
				})
				return
			}

			files = append(files, url)
		}
	}

	// Associate with user if provided (from authenticated contact form)
	var userID *string
	if uid := getFormValue(form, "userId"); uid != "" {
		userID = &uid
	}

	// Create inquiry record
	inquiry := &modelsProduct.Inquiry{
		ID:                    crypto.GenerateID(),
		UserID:                userID,
		CompanyName:           companyName,
		ContactPerson:         contactPerson,
		Email:                 email,
		WhatsApp:              getFormValue(form, "whatsapp"),
		TargetCountry:         getFormValue(form, "targetCountry"),
		EstimatedQuantity:     getFormValue(form, "estimatedQuantity"),
		InterestedProducts:    interestedProducts,
		PackagingRequirements: getFormValue(form, "packagingRequirements"),
		FlavorRequirements:    getFormValue(form, "flavorRequirements"),
		OEMNeeded:             oemNeeded,
		ExpectedDelivery:      getFormValue(form, "expectedDelivery"),
		Message:               getFormValue(form, "message"),
		Files:                 files,
		Status:                "pending",
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	// Submit inquiry (saves to database and sends email)
	if err := h.services.Inquiry.SubmitInquiry(c.Request.Context(), inquiry); err != nil {
		response.ErrorResp(c, http.StatusInternalServerError, "inquiry_create_failed")
		return
	}

	// Early compliance pre-check: warn about potential issues for the target country.
	// This is advisory only — it does not block inquiry submission.
	var complianceWarnings []string
	targetCountry := inquiry.TargetCountry
	if targetCountry != "" && len(interestedProducts) > 0 && h.services.Product != nil {
		var checkProducts []modelsProduct.Product
		for _, pid := range interestedProducts {
			if p, err := h.services.Product.GetProductByID(c.Request.Context(), pid); err == nil {
				checkProducts = append(checkProducts, *p)
			}
		}
		if len(checkProducts) > 0 {
			result := h.services.Product.ValidateComplianceWithMarketProfiles(c.Request.Context(), targetCountry, checkProducts)
			complianceWarnings = append(complianceWarnings, result.Violations...)
			complianceWarnings = append(complianceWarnings, result.Warnings...)
		}
	}

	response := gin.H{
		"success":   true,
		"message":   "Thank you for your inquiry. We will contact you within 24 hours.",
		"inquiryId": inquiry.ID,
	}
	if len(complianceWarnings) > 0 {
		response["complianceWarnings"] = complianceWarnings
	}

	c.JSON(http.StatusOK, response)
}

// ===== Helper Functions =====

// getFormValue gets a single value from multipart form
func getFormValue(form *multipart.Form, key string) string {
	if values, ok := form.Value[key]; ok && len(values) > 0 {
		return strings.TrimSpace(values[0])
	}
	return ""
}

// getFormValues gets multiple values from multipart form
func getFormValues(form *multipart.Form, key string) []string {
	if values, ok := form.Value[key]; ok {
		result := make([]string, 0, len(values))
		for _, item := range values {
			if item = strings.TrimSpace(item); item != "" {
				result = append(result, item)
			}
		}
		return result
	}
	return []string{}
}

// isAllowedType checks if content type is allowed
func isAllowedType(contentType string, allowedTypes []string) bool {
	for _, t := range allowedTypes {
		if t == contentType {
			return true
		}
	}
	return false
}
