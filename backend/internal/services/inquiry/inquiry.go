package inquiry

import (
	"candypro/api/internal/config"
	modelsProduct "candypro/api/internal/models/product"
	content "candypro/api/internal/services/content"
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

type inquiryRepository interface {
	Create(ctx context.Context, inquiry *modelsProduct.Inquiry) error
	FindByID(ctx context.Context, id string) (*modelsProduct.Inquiry, error)
	FindAll(ctx context.Context, page, limit int) ([]modelsProduct.Inquiry, int64, error)
	FindByUserID(ctx context.Context, userID string, page, limit int) ([]modelsProduct.Inquiry, int64, error)
	UpdateStatus(ctx context.Context, id string, status string) error
	Update(ctx context.Context, inquiry *modelsProduct.Inquiry) error
	Delete(ctx context.Context, id string) error
	CountAll(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status string) (int64, error)
	FindRecent(ctx context.Context, limit int) ([]modelsProduct.Inquiry, error)
	ConversionByMonth(ctx context.Context, months int) ([]map[string]interface{}, error)
}

// sanitizeEmailField sanitizes input for email body to prevent injection and abuse.
func sanitizeEmailField(input string) string {
	input = strings.ReplaceAll(input, "\r", "")
	if len(input) > 10000 {
		input = input[:10000]
	}
	return input
}

// InquiryService handles inquiry business logic.
type InquiryService struct {
	repo inquiryRepository
	cfg  *config.Config
}

// NewInquiryService creates a new InquiryService.
func NewInquiryService(repo inquiryRepository, cfg *config.Config) *InquiryService {
	return &InquiryService{repo: repo, cfg: cfg}
}

// SubmitInquiry creates and processes a new inquiry.
func (s *InquiryService) SubmitInquiry(ctx context.Context, inquiry *modelsProduct.Inquiry) error {
	inquiry.Status = "pending"
	inquiry.CreatedAt = time.Now()
	inquiry.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, inquiry); err != nil {
		return err
	}

	go s.sendNotificationEmail(inquiry)

	log.Printf("Inquiry submitted: %s from %s", inquiry.ID, inquiry.CompanyName)
	return nil
}

// sendNotificationEmail sends notification email about new inquiry.
func (s *InquiryService) sendNotificationEmail(inquiry *modelsProduct.Inquiry) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	emailService := content.NewEmailService(s.cfg.Email)

	subject := "New Inquiry from " + inquiry.CompanyName
	body := s.buildNotificationBody(inquiry)
	if err := emailService.SendEmail(ctx, s.cfg.Email.FromEmail, subject, body); err != nil {
		log.Printf("Failed to send notification email: %v", err)
	}

	customerSubject := "Thank you for your inquiry - CandyPro OEM"
	customerBody := s.buildConfirmationBody(inquiry)
	if err := emailService.SendEmail(ctx, inquiry.Email, customerSubject, customerBody); err != nil {
		log.Printf("Failed to send confirmation email: %v", err)
	}
}

func (s *InquiryService) buildNotificationBody(inquiry *modelsProduct.Inquiry) string {
	return fmt.Sprintf(`
New inquiry received:

Company: %s
Contact Person: %s
Email: %s
WhatsApp: %s
Target Country: %s
Estimated Quantity: %s
Interested Products: %s
OEM Needed: %s
Expected Delivery: %s

Message:
%s

---
Submitted at: %s
Inquiry ID: %s
`,
		sanitizeEmailField(inquiry.CompanyName),
		sanitizeEmailField(inquiry.ContactPerson),
		sanitizeEmailField(inquiry.Email),
		sanitizeEmailField(inquiry.WhatsApp),
		sanitizeEmailField(inquiry.TargetCountry),
		sanitizeEmailField(inquiry.EstimatedQuantity),
		sanitizeEmailField(strings.Join(inquiry.InterestedProducts, ", ")),
		boolToString(inquiry.OEMNeeded),
		sanitizeEmailField(inquiry.ExpectedDelivery),
		sanitizeEmailField(inquiry.Message),
		inquiry.CreatedAt.Format("2006-01-02 15:04:05"),
		inquiry.ID,
	)
}

func (s *InquiryService) buildConfirmationBody(inquiry *modelsProduct.Inquiry) string {
	return fmt.Sprintf(`
Dear %s,

Thank you for your inquiry to CandyPro OEM. We have received your request and our team will review it shortly.

Our business development team will contact you within 24 hours to discuss your requirements in detail.

Inquiry Details:
- Company: %s
- Inquiry ID: %s

If you have any urgent questions, please feel free to contact us at:
- Email: %s
- WhatsApp: Available on our website

Best regards,
CandyPro OEM Team
https://candypro-oem.com
`,
		sanitizeEmailField(inquiry.ContactPerson),
		sanitizeEmailField(inquiry.CompanyName),
		inquiry.ID,
		s.cfg.Email.FromEmail,
	)
}

func boolToString(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// GetInquiry returns an inquiry by ID.
func (s *InquiryService) GetInquiry(ctx context.Context, id string) (*modelsProduct.Inquiry, error) {
	return s.repo.FindByID(ctx, id)
}

// GetInquiries returns all inquiries with pagination.
func (s *InquiryService) GetInquiries(ctx context.Context, page, limit int) ([]modelsProduct.Inquiry, int64, error) {
	return s.repo.FindAll(ctx, page, limit)
}

// GetUserInquiries returns inquiries for a specific user with pagination.
func (s *InquiryService) GetUserInquiries(ctx context.Context, userID string, page, limit int) ([]modelsProduct.Inquiry, int64, error) {
	return s.repo.FindByUserID(ctx, userID, page, limit)
}

// validStatuses contains the allowed inquiry status values.
var validStatuses = map[string]bool{
	"pending":     true,
	"contacted":   true,
	"quoted":      true,
	"negotiating": true,
	"won":         true,
	"lost":        true,
	"closed":      true,
}

// UpdateInquiryStatus updates inquiry status.
func (s *InquiryService) UpdateInquiryStatus(ctx context.Context, id string, status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	return s.repo.UpdateStatus(ctx, id, status)
}

// UpdateInquiry updates an inquiry.
func (s *InquiryService) UpdateInquiry(ctx context.Context, inquiry *modelsProduct.Inquiry) error {
	inquiry.UpdatedAt = time.Now()
	return s.repo.Update(ctx, inquiry)
}

// CreateInquiry creates inquiry without email side effects.
func (s *InquiryService) CreateInquiry(ctx context.Context, inquiry *modelsProduct.Inquiry) error {
	if inquiry.Status == "" {
		inquiry.Status = "pending"
	}
	if inquiry.Priority == "" {
		inquiry.Priority = "normal"
	}
	inquiry.CreatedAt = time.Now()
	inquiry.UpdatedAt = time.Now()
	return s.repo.Create(ctx, inquiry)
}

// DeleteInquiry deletes an inquiry by id.
func (s *InquiryService) DeleteInquiry(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// CountInquiries returns total inquiry count.
func (s *InquiryService) CountInquiries(ctx context.Context) (int64, error) {
	return s.repo.CountAll(ctx)
}

// CountInquiriesByStatus returns inquiry count for status.
func (s *InquiryService) CountInquiriesByStatus(ctx context.Context, status string) (int64, error) {
	return s.repo.CountByStatus(ctx, status)
}

// GetRecentInquiries returns latest inquiries.
func (s *InquiryService) GetRecentInquiries(ctx context.Context, limit int) ([]modelsProduct.Inquiry, error) {
	return s.repo.FindRecent(ctx, limit)
}

// GetConversionByMonth returns monthly inquiry conversion data.
func (s *InquiryService) GetConversionByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return s.repo.ConversionByMonth(ctx, months)
}

// ConversionByMonth returns monthly inquiry conversion data (alias for handler compatibility).
func (s *InquiryService) ConversionByMonth(ctx context.Context, months int) ([]map[string]interface{}, error) {
	return s.repo.ConversionByMonth(ctx, months)
}
