package inquiry

import (
	"candypro/api/internal/config"
	modelsProduct "candypro/api/internal/models/product"
	content "candypro/api/internal/services/content"
	"candypro/api/internal/pkg/i18n"
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
	CountByStatusGrouped(ctx context.Context) (map[string]int64, error)
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
func (s *InquiryService) SubmitInquiry(ctx context.Context, inquiry *modelsProduct.Inquiry, locale string) error {
	inquiry.Status = "pending"
	inquiry.CreatedAt = time.Now()
	inquiry.UpdatedAt = time.Now()

	if err := s.repo.Create(ctx, inquiry); err != nil {
		return err
	}

	if locale == "" {
		locale = i18n.DefaultLocale()
	}
	go s.sendNotificationEmail(inquiry, locale)

	log.Printf("Inquiry submitted: %s from %s", inquiry.ID, inquiry.CompanyName)
	return nil
}

// sendNotificationEmail sends notification email about new inquiry.
func (s *InquiryService) sendNotificationEmail(inquiry *modelsProduct.Inquiry, locale string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	emailService := content.NewEmailService(s.cfg.Email)

	subject := i18n.TranslateWithVars(locale, "emails.inquiry_admin_subject", map[string]string{
		"company": sanitizeEmailField(inquiry.CompanyName),
	})
	body := s.buildNotificationBody(inquiry, locale)
	if err := emailService.SendEmail(ctx, s.cfg.Email.FromEmail, subject, body); err != nil {
		log.Printf("Failed to send notification email: %v", err)
	}

	customerSubject := i18n.Translate(locale, "emails.inquiry_confirmation_subject")
	customerBody := s.buildConfirmationBody(inquiry, locale)
	if err := emailService.SendEmail(ctx, inquiry.Email, customerSubject, customerBody); err != nil {
		log.Printf("Failed to send confirmation email: %v", err)
	}
}

func (s *InquiryService) buildNotificationBody(inquiry *modelsProduct.Inquiry, locale string) string {
	return fmt.Sprintf("%s\n\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n%s: %s\n\n%s:\n%s\n\n---\n%s: %s\n%s: %s",
		i18n.Translate(locale, "emails.inquiry_admin_intro"),
		i18n.Translate(locale, "emails.inquiry_field_company"),
		sanitizeEmailField(inquiry.CompanyName),
		i18n.Translate(locale, "emails.inquiry_field_contact"),
		sanitizeEmailField(inquiry.ContactPerson),
		i18n.Translate(locale, "emails.inquiry_field_email"),
		sanitizeEmailField(inquiry.Email),
		i18n.Translate(locale, "emails.inquiry_field_whatsapp"),
		sanitizeEmailField(inquiry.WhatsApp),
		i18n.Translate(locale, "emails.inquiry_field_country"),
		sanitizeEmailField(inquiry.TargetCountry),
		i18n.Translate(locale, "emails.inquiry_field_quantity"),
		sanitizeEmailField(inquiry.EstimatedQuantity),
		i18n.Translate(locale, "emails.inquiry_field_products"),
		sanitizeEmailField(strings.Join(inquiry.InterestedProducts, ", ")),
		i18n.Translate(locale, "emails.inquiry_field_oem"),
		boolToLocalized(inquiry.OEMNeeded, locale),
		i18n.Translate(locale, "emails.inquiry_field_delivery"),
		sanitizeEmailField(inquiry.ExpectedDelivery),
		i18n.Translate(locale, "emails.inquiry_field_message"),
		sanitizeEmailField(inquiry.Message),
		i18n.Translate(locale, "emails.inquiry_submitted_at"),
		inquiry.CreatedAt.Format("2006-01-02 15:04:05"),
		i18n.Translate(locale, "emails.inquiry_field_id"),
		inquiry.ID,
	)
}

func (s *InquiryService) buildConfirmationBody(inquiry *modelsProduct.Inquiry, locale string) string {
	frontendURL := "http://localhost:3000"
	if s.cfg != nil && s.cfg.Security.FrontendURL != "" {
		frontendURL = strings.TrimRight(s.cfg.Security.FrontendURL, "/")
	}
	fromEmail := ""
	if s.cfg != nil {
		fromEmail = s.cfg.Email.FromEmail
	}
	return fmt.Sprintf("%s\n\n%s\n\n%s\n- %s: %s\n- %s: %s\n\n%s\n- %s: %s\n- %s\n\n%s\n%s",
		i18n.TranslateWithVars(locale, "emails.inquiry_confirmation_greeting", map[string]string{
			"name": sanitizeEmailField(inquiry.ContactPerson),
		}),
		i18n.Translate(locale, "emails.inquiry_confirmation_body"),
		i18n.Translate(locale, "emails.inquiry_confirmation_details"),
		i18n.Translate(locale, "emails.inquiry_field_company"),
		sanitizeEmailField(inquiry.CompanyName),
		i18n.Translate(locale, "emails.inquiry_field_id"),
		inquiry.ID,
		i18n.Translate(locale, "emails.inquiry_confirmation_contact"),
		i18n.Translate(locale, "emails.inquiry_field_email"),
		fromEmail,
		i18n.Translate(locale, "emails.inquiry_confirmation_whatsapp"),
		i18n.Translate(locale, "emails.inquiry_confirmation_signoff"),
		frontendURL,
	)
}

func boolToLocalized(b bool, locale string) string {
	if b {
		return i18n.Translate(locale, "common.yes")
	}
	return i18n.Translate(locale, "common.no")
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
	"pending":              true,
	"contacted":            true,
	"quoted":               true,
	"negotiating":          true,
	"won":                  true,
	"lost":                 true,
	"closed":               true,
	"confirmed":            true,
	"pending_confirmation": true,
}

// inquiryStatusTransitions 询价状态合法流转矩阵
var inquiryStatusTransitions = map[string]map[string]bool{
	"pending":              {"contacted": true, "lost": true, "closed": true},
	"contacted":            {"quoted": true, "negotiating": true, "lost": true, "closed": true},
	"quoted":               {"negotiating": true, "won": true, "lost": true, "closed": true},
	"negotiating":          {"quoted": true, "won": true, "lost": true, "closed": true},
	"won":                  {"confirmed": true, "closed": true},
	"confirmed":            {"closed": true},
	"pending_confirmation": {"confirmed": true, "negotiating": true, "lost": true, "closed": true},
	"lost":                 {"closed": true},
	"closed":               {},
}

// ValidateInquiryStatus 校验询价状态值是否合法
func (s *InquiryService) ValidateInquiryStatus(status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}
	return nil
}

// ValidateInquiryStatusTransition 校验询价状态流转是否允许
func (s *InquiryService) ValidateInquiryStatusTransition(from, to string) error {
	if from == to {
		return nil
	}
	if err := s.ValidateInquiryStatus(to); err != nil {
		return err
	}
	allowed, ok := inquiryStatusTransitions[from]
	if !ok || !allowed[to] {
		return fmt.Errorf("invalid status transition: %s -> %s", from, to)
	}
	return nil
}

// UpdateInquiryStatus updates inquiry status.
func (s *InquiryService) UpdateInquiryStatus(ctx context.Context, id string, status string) error {
	if err := s.ValidateInquiryStatus(status); err != nil {
		return err
	}
	inquiry, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.ValidateInquiryStatusTransition(inquiry.Status, status); err != nil {
		return err
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

// CountInquiriesByStatusGrouped returns counts for all statuses in one query (M-1).
func (s *InquiryService) CountInquiriesByStatusGrouped(ctx context.Context) (map[string]int64, error) {
	return s.repo.CountByStatusGrouped(ctx)
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
