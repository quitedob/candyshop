package content

import (
	"context"
	"fmt"
	"log"
	"time"

	"candypro/api/internal/config"

	"gopkg.in/gomail.v2"
)

// Default timeout for email operations
const emailTimeout = 30 * time.Second

// EmailService handles email operations
type EmailService struct {
	cfg config.EmailConfig
}

// NewEmailService creates a new EmailService
func NewEmailService(cfg config.EmailConfig) *EmailService {
	return &EmailService{cfg: cfg}
}

// SendEmail sends a plain-text email with context timeout to prevent hanging.
func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string) error {
	return s.send(ctx, to, subject, body, "text/plain")
}

// SendHTML sends an HTML email with context timeout to prevent hanging.
func (s *EmailService) SendHTML(ctx context.Context, to, subject, body string) error {
	return s.send(ctx, to, subject, body, "text/html")
}

// send is the shared implementation that honours context cancellation/timeout.
func (s *EmailService) send(ctx context.Context, to, subject, body, contentType string) error {
	// Skip if SMTP is not configured
	if s.cfg.SMTPHost == "" || s.cfg.SMTPUser == "" {
		log.Printf("Email not configured, skipping email to %s", to)
		return nil
	}

	// Ensure we have a deadline so the caller can't hang forever.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, emailTimeout)
		defer cancel()
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.cfg.FromEmail)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody(contentType, body)

	d := gomail.NewDialer(
		s.cfg.SMTPHost,
		s.cfg.SMTPPort,
		s.cfg.SMTPUser,
		s.cfg.SMTPPassword,
	)

	// Use a channel to capture the send result so we can race it against
	// the context deadline. gomail.Dialer.DialAndSend does not accept a
	// context, so we wrap it ourselves.
	type result struct{ err error }
	ch := make(chan result, 1)

	go func() {
		ch <- result{err: d.DialAndSend(m)}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("email send to %s timed out: %w", to, ctx.Err())
	case res := <-ch:
		if res.err != nil {
			log.Printf("Failed to send email: %v", res.err)
			return res.err
		}
	}

	log.Printf("Email sent successfully to %s", to)
	return nil
}
