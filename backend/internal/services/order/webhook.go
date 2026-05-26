package order

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	modelsOrder "candypro/api/internal/models/order"
	"candypro/api/internal/pkg/safego"
)

type webhookRepository interface {
	List(ctx context.Context) ([]modelsOrder.WebhookConfig, error)
	Get(ctx context.Context, id uint) (*modelsOrder.WebhookConfig, error)
	Create(ctx context.Context, cfg *modelsOrder.WebhookConfig) error
	Update(ctx context.Context, cfg *modelsOrder.WebhookConfig) error
	Delete(ctx context.Context, id uint) error
	FindByEvent(ctx context.Context, eventType string) ([]modelsOrder.WebhookConfig, error)
	CreateDelivery(ctx context.Context, d *modelsOrder.WebhookDelivery) error
	UpdateDelivery(ctx context.Context, d *modelsOrder.WebhookDelivery) error
	ListDeliveries(ctx context.Context, webhookID uint, limit, offset int) ([]modelsOrder.WebhookDelivery, int64, error)
}

// WebhookService manages outgoing webhook delivery.
type WebhookService struct {
	repo       webhookRepository
	httpClient *http.Client
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(repo webhookRepository) *WebhookService {
	return &WebhookService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Dispatch fires webhooks for the given event asynchronously.
func (s *WebhookService) Dispatch(ctx context.Context, eventType, aggregateKey string, payload any) {
	configs, err := s.repo.FindByEvent(ctx, eventType)
	if err != nil {
		log.Printf("webhook: FindByEvent(%s) error: %v", eventType, err)
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("webhook: marshal payload error: %v", err)
		return
	}

	for _, cfg := range configs {
		delivery := &modelsOrder.WebhookDelivery{
			WebhookID:    cfg.ID,
			EventType:    eventType,
			AggregateKey: aggregateKey,
			Payload:      string(payloadBytes),
			Status:       modelsOrder.WebhookDeliveryPending,
			Attempts:     1,
			CreatedAt:    time.Now(),
		}
		if err := s.repo.CreateDelivery(ctx, delivery); err != nil {
			log.Printf("webhook: create delivery error: %v", err)
			continue
		}
		cfgCopy := cfg
		deliveryRef := delivery
		safego.Go("webhook.deliver", func() {
			s.deliver(cfgCopy, deliveryRef, payloadBytes)
		})
	}
}

// deliver sends a single webhook with signature and retry.
func (s *WebhookService) deliver(cfg modelsOrder.WebhookConfig, delivery *modelsOrder.WebhookDelivery, payload []byte) {
	maxRetries := 5
	backoff := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute, 1 * time.Hour, 6 * time.Hour}

	for attempt := 1; attempt <= maxRetries; attempt++ {
		statusCode, respBody, err := s.send(cfg, payload)
		now := time.Now()

		if err == nil && statusCode >= 200 && statusCode < 300 {
			delivery.Status = modelsOrder.WebhookDeliverySuccess
			delivery.ResponseCode = &statusCode
			delivery.ResponseBody = respBody
			delivery.DeliveredAt = &now
			delivery.Attempts = attempt
			if ue := s.repo.UpdateDelivery(context.Background(), delivery); ue != nil {
				log.Printf("webhook: UpdateDelivery failed for delivery %d: %v", delivery.ID, ue)
			}
			return
		}

		if err != nil {
			delivery.LastError = fmt.Sprintf("attempt %d: %v", attempt, err)
		} else {
			delivery.LastError = fmt.Sprintf("attempt %d: HTTP %d", attempt, statusCode)
			delivery.ResponseCode = &statusCode
			delivery.ResponseBody = respBody
		}
		delivery.Attempts = attempt

		if attempt < maxRetries {
			if ue := s.repo.UpdateDelivery(context.Background(), delivery); ue != nil {
				log.Printf("webhook: UpdateDelivery failed for delivery %d: %v", delivery.ID, ue)
			}
			time.Sleep(backoff[attempt-1])
		}
	}

	delivery.Status = modelsOrder.WebhookDeliveryFailed
	_ = s.repo.UpdateDelivery(context.Background(), delivery)
}

func (s *WebhookService) send(cfg modelsOrder.WebhookConfig, payload []byte) (int, string, error) {
	req, err := http.NewRequest(http.MethodPost, cfg.URL, bytes.NewReader(payload))
	if err != nil {
		return 0, "", err
	}

	sig := signPayload(cfg.Secret, payload)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CandyPro-Signature", sig)
	req.Header.Set("X-CandyPro-Webhook-ID", fmt.Sprintf("%d", cfg.ID))
	req.Header.Set("User-Agent", "CandyPro-Webhook/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	return resp.StatusCode, buf.String(), nil
}

func signPayload(secret string, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// ── Admin CRUD ──

func (s *WebhookService) List(ctx context.Context) ([]modelsOrder.WebhookConfig, error) {
	return s.repo.List(ctx)
}

func (s *WebhookService) Get(ctx context.Context, id uint) (*modelsOrder.WebhookConfig, error) {
	return s.repo.Get(ctx, id)
}

func (s *WebhookService) Create(ctx context.Context, cfg *modelsOrder.WebhookConfig) error {
	return s.repo.Create(ctx, cfg)
}

func (s *WebhookService) Update(ctx context.Context, cfg *modelsOrder.WebhookConfig) error {
	return s.repo.Update(ctx, cfg)
}

func (s *WebhookService) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *WebhookService) ListDeliveries(ctx context.Context, webhookID uint, limit, offset int) ([]modelsOrder.WebhookDelivery, int64, error) {
	return s.repo.ListDeliveries(ctx, webhookID, limit, offset)
}
