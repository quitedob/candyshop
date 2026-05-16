package order

import (
	"context"
	"slices"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// WebhookRepository manages webhook configuration and delivery logs.
type WebhookRepository struct {
	db *gorm.DB
}

func NewWebhookRepository(db *gorm.DB) *WebhookRepository {
	return &WebhookRepository{db: db}
}

// ── Config CRUD ──

func (r *WebhookRepository) List(ctx context.Context) ([]modelsOrder.WebhookConfig, error) {
	var configs []modelsOrder.WebhookConfig
	err := r.db.WithContext(ctx).Order("id ASC").Find(&configs).Error
	return configs, err
}

func (r *WebhookRepository) Get(ctx context.Context, id uint) (*modelsOrder.WebhookConfig, error) {
	var cfg modelsOrder.WebhookConfig
	err := r.db.WithContext(ctx).First(&cfg, id).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *WebhookRepository) Create(ctx context.Context, cfg *modelsOrder.WebhookConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *WebhookRepository) Update(ctx context.Context, cfg *modelsOrder.WebhookConfig) error {
	cfg.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Model(cfg).Select("*").Updates(cfg).Error
}

func (r *WebhookRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsOrder.WebhookConfig{}, id).Error
}

// FindByEvent returns all active webhooks subscribed to the given event type.
func (r *WebhookRepository) FindByEvent(ctx context.Context, eventType string) ([]modelsOrder.WebhookConfig, error) {
	var configs []modelsOrder.WebhookConfig
	err := r.db.WithContext(ctx).
		Where("status = ?", "active").
		Order("id ASC").
		Find(&configs).Error
	if err != nil {
		return nil, err
	}

	// Filter in-memory for event subscription match.
	var matched []modelsOrder.WebhookConfig
	for _, c := range configs {
		if slices.Contains(c.Events, eventType) {
			matched = append(matched, c)
		}
	}
	return matched, nil
}

// ── Delivery log ──

func (r *WebhookRepository) CreateDelivery(ctx context.Context, d *modelsOrder.WebhookDelivery) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *WebhookRepository) UpdateDelivery(ctx context.Context, d *modelsOrder.WebhookDelivery) error {
	return r.db.WithContext(ctx).Model(d).Select("*").Updates(d).Error
}

// ListDeliveries returns recent delivery logs with pagination.
func (r *WebhookRepository) ListDeliveries(ctx context.Context, webhookID uint, limit, offset int) ([]modelsOrder.WebhookDelivery, int64, error) {
	var deliveries []modelsOrder.WebhookDelivery
	var total int64

	q := r.db.WithContext(ctx).Model(&modelsOrder.WebhookDelivery{})
	if webhookID > 0 {
		q = q.Where("webhook_id = ?", webhookID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&deliveries).Error
	return deliveries, total, err
}

// PendingDeliveries returns deliveries that are pending and due for retry (locked for update).
func (r *WebhookRepository) PendingDeliveries(ctx context.Context, limit int) ([]modelsOrder.WebhookDelivery, error) {
	var deliveries []modelsOrder.WebhookDelivery
	err := r.db.WithContext(ctx).
		Where("status = ?", modelsOrder.WebhookDeliveryPending).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Order("created_at ASC").
		Limit(limit).
		Find(&deliveries).Error
	return deliveries, err
}
