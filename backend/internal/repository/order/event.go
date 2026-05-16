package order

import (
	"context"
	"time"

	modelsOrder "candypro/api/internal/models/order"

	"gorm.io/gorm"
)

// ── Event Repository ──

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Create(ctx context.Context, event *modelsOrder.Event) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *EventRepository) List(ctx context.Context, limit, offset int) ([]modelsOrder.Event, int64, error) {
	var events []modelsOrder.Event
	var total int64
	q := r.db.WithContext(ctx).Model(&modelsOrder.Event{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&events).Error
	return events, total, err
}

func (r *EventRepository) Get(ctx context.Context, id uint) (*modelsOrder.Event, error) {
	var event modelsOrder.Event
	err := r.db.WithContext(ctx).First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// ── Hook Repository ──

type HookRepository struct {
	db *gorm.DB
}

func NewHookRepository(db *gorm.DB) *HookRepository {
	return &HookRepository{db: db}
}

func (r *HookRepository) List(ctx context.Context) ([]modelsOrder.HookConfig, error) {
	var hooks []modelsOrder.HookConfig
	err := r.db.WithContext(ctx).Order("id ASC").Find(&hooks).Error
	return hooks, err
}

func (r *HookRepository) Get(ctx context.Context, id uint) (*modelsOrder.HookConfig, error) {
	var hook modelsOrder.HookConfig
	err := r.db.WithContext(ctx).First(&hook, id).Error
	if err != nil {
		return nil, err
	}
	return &hook, nil
}

func (r *HookRepository) Create(ctx context.Context, hook *modelsOrder.HookConfig) error {
	return r.db.WithContext(ctx).Create(hook).Error
}

func (r *HookRepository) Update(ctx context.Context, hook *modelsOrder.HookConfig) error {
	hook.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Model(hook).Select("*").Updates(hook).Error
}

func (r *HookRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&modelsOrder.HookConfig{}, id).Error
}

func (r *HookRepository) FindByEvent(ctx context.Context, eventName string) ([]modelsOrder.HookConfig, error) {
	var hooks []modelsOrder.HookConfig
	err := r.db.WithContext(ctx).
		Where("event_name = ? AND status = ?", eventName, "active").
		Order("id ASC").
		Find(&hooks).Error
	return hooks, err
}

// ── Hook Execution Repository ──

type HookExecutionRepository struct {
	db *gorm.DB
}

func NewHookExecutionRepository(db *gorm.DB) *HookExecutionRepository {
	return &HookExecutionRepository{db: db}
}

func (r *HookExecutionRepository) Create(ctx context.Context, exec *modelsOrder.HookExecution) error {
	return r.db.WithContext(ctx).Create(exec).Error
}

func (r *HookExecutionRepository) Update(ctx context.Context, exec *modelsOrder.HookExecution) error {
	return r.db.WithContext(ctx).Model(exec).Select("*").Updates(exec).Error
}

func (r *HookExecutionRepository) ListByHook(ctx context.Context, hookID uint, limit, offset int) ([]modelsOrder.HookExecution, int64, error) {
	var executions []modelsOrder.HookExecution
	var total int64
	q := r.db.WithContext(ctx).Model(&modelsOrder.HookExecution{}).Where("hook_id = ?", hookID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&executions).Error
	return executions, total, err
}

func (r *HookExecutionRepository) ListByEvent(ctx context.Context, eventID uint, limit, offset int) ([]modelsOrder.HookExecution, int64, error) {
	var executions []modelsOrder.HookExecution
	var total int64
	q := r.db.WithContext(ctx).Model(&modelsOrder.HookExecution{}).Where("event_id = ?", eventID)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&executions).Error
	return executions, total, err
}
