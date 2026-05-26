package common

import (
	"context"
	"fmt"
	"log"
	"time"

	"candypro/api/internal/config"
	modelsCommon "candypro/api/internal/models/common"
	commonrepo "candypro/api/internal/repository/common"
	emailsvc "candypro/api/internal/services/content"
)

// NotificationRelay drains the notification_outbox table and dispatches via
// per-channel adapters. Implements C-8: external sends are decoupled from the
// originating DB transaction so a failed send never leaves the DB ahead of the
// real-world side effect.
type NotificationRelay struct {
	repo  *commonrepo.NotificationOutboxRepository
	email *emailsvc.EmailService

	maxAttempts int
	baseBackoff time.Duration
}

// NewNotificationRelay creates a relay bound to the supplied outbox repo and
// email service.
func NewNotificationRelay(repo *commonrepo.NotificationOutboxRepository, cfg *config.Config) *NotificationRelay {
	r := &NotificationRelay{
		repo:        repo,
		maxAttempts: 5,
		baseBackoff: 30 * time.Second,
	}
	if cfg != nil {
		r.email = emailsvc.NewEmailService(cfg.Email)
	}
	return r
}

// ProcessBatch claims up to `limit` pending rows and tries to dispatch each.
// Returns the number of successfully processed rows.
func (r *NotificationRelay) ProcessBatch(ctx context.Context, limit int) (int, error) {
	if r == nil || r.repo == nil {
		return 0, nil
	}
	rows, err := r.repo.ClaimPending(ctx, limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, row := range rows {
		if err := r.dispatchOne(ctx, &row); err != nil {
			backoff := r.baseBackoff * time.Duration(1<<minInt(row.Attempts, 6))
			if mfErr := r.repo.MarkFailed(ctx, row.ID, err.Error(), row.Attempts, r.maxAttempts, backoff); mfErr != nil {
				log.Printf("notification relay: MarkFailed for %d: %v", row.ID, mfErr)
			}
			continue
		}
		if mErr := r.repo.MarkProcessed(ctx, row.ID); mErr != nil {
			log.Printf("notification relay: MarkProcessed for %d: %v", row.ID, mErr)
			continue
		}
		processed++
	}
	return processed, nil
}

// dispatchOne dispatches a single row according to its channel.
func (r *NotificationRelay) dispatchOne(ctx context.Context, row *modelsCommon.NotificationOutbox) error {
	switch row.Channel {
	case modelsCommon.NotificationChannelEmail:
		if r.email == nil {
			return fmt.Errorf("email service not configured")
		}
		// Bound the dispatch to a sane timeout so a hung SMTP server can't
		// hold the relay loop forever.
		sendCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		return r.email.SendEmail(sendCtx, row.Recipient, row.Subject, row.Body)
	default:
		return fmt.Errorf("unsupported channel %q", row.Channel)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
