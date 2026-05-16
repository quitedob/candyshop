package order

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	modelsOrder "candypro/api/internal/models/order"
)

type eventRepo interface {
	Create(ctx context.Context, event *modelsOrder.Event) error
	List(ctx context.Context, limit, offset int) ([]modelsOrder.Event, int64, error)
	Get(ctx context.Context, id uint) (*modelsOrder.Event, error)
}

type hookFinder interface {
	FindByEvent(ctx context.Context, eventName string) ([]modelsOrder.HookConfig, error)
}

type executionRepo interface {
	Create(ctx context.Context, exec *modelsOrder.HookExecution) error
	Update(ctx context.Context, exec *modelsOrder.HookExecution) error
}

// InternalHookFunc is called for hooks of type "internal".
type InternalHookFunc func(ctx context.Context, eventName string, payload json.RawMessage) error

// EventBus emits lifecycle events and dispatches to registered hooks.
type EventBus struct {
	eventRepo    eventRepo
	hookRepo     hookFinder
	execRepo     executionRepo
	webhookSvc   *WebhookService
	internalFunc InternalHookFunc
}

// NewEventBus creates a new EventBus.
func NewEventBus(eventRepo eventRepo, hookRepo hookFinder, execRepo executionRepo, webhookSvc *WebhookService) *EventBus {
	return &EventBus{
		eventRepo:  eventRepo,
		hookRepo:   hookRepo,
		execRepo:   execRepo,
		webhookSvc: webhookSvc,
	}
}

// RegisterInternalHandler sets the handler for "internal" type hooks.
func (eb *EventBus) RegisterInternalHandler(fn InternalHookFunc) {
	eb.internalFunc = fn
}

// Emit creates an event and dispatches to all matching active hooks.
func (eb *EventBus) Emit(ctx context.Context, eventName string, payload any) (*modelsOrder.Event, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("event emit marshal: %w", err)
	}

	event := &modelsOrder.Event{
		Name:      eventName,
		Payload:   payloadBytes,
		CreatedAt: time.Now(),
	}
	if err := eb.eventRepo.Create(ctx, event); err != nil {
		return nil, fmt.Errorf("event emit create: %w", err)
	}

	hooks, err := eb.hookRepo.FindByEvent(ctx, eventName)
	if err != nil {
		log.Printf("eventbus: FindByEvent(%s) error: %v", eventName, err)
		return event, nil
	}

	for _, hook := range hooks {
		exec := &modelsOrder.HookExecution{
			HookID:    hook.ID,
			EventID:   event.ID,
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		if err := eb.execRepo.Create(ctx, exec); err != nil {
			log.Printf("eventbus: create execution error: %v", err)
			continue
		}

		switch hook.Type {
		case "webhook":
			eb.webhookSvc.Dispatch(ctx, eventName, fmt.Sprintf("event_%d", event.ID), payload)
		case "internal":
			if eb.internalFunc != nil {
				go eb.executeInternal(ctx, exec, eventName, payloadBytes)
			}
		default:
			log.Printf("eventbus: unknown hook type %q for hook %d", hook.Type, hook.ID)
		}
	}

	return event, nil
}

// ListEvents returns paginated events.
func (eb *EventBus) ListEvents(ctx context.Context, limit, offset int) ([]modelsOrder.Event, int64, error) {
	return eb.eventRepo.List(ctx, limit, offset)
}

// GetEvent returns a single event by id.
func (eb *EventBus) GetEvent(ctx context.Context, id uint) (*modelsOrder.Event, error) {
	return eb.eventRepo.Get(ctx, id)
}

func (eb *EventBus) executeInternal(ctx context.Context, exec *modelsOrder.HookExecution, eventName string, payload json.RawMessage) {
	start := time.Now()
	err := eb.internalFunc(ctx, eventName, payload)
	durationMs := int(time.Since(start).Milliseconds())

	exec.Status = "success"
	if err != nil {
		exec.Status = "failed"
		exec.Error = err.Error()
	}
	exec.DurationMs = durationMs
	if ue := eb.execRepo.Update(context.Background(), exec); ue != nil {
		log.Printf("event_bus: Update execution failed for exec %d: %v", exec.ID, ue)
	}
}
