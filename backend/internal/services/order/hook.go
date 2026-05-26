package order

import (
	"context"
	"time"

	modelsOrder "candypro/api/internal/models/order"
)

type hookConfigRepo interface {
	List(ctx context.Context) ([]modelsOrder.HookConfig, error)
	Get(ctx context.Context, id uint) (*modelsOrder.HookConfig, error)
	Create(ctx context.Context, hook *modelsOrder.HookConfig) error
	Update(ctx context.Context, hook *modelsOrder.HookConfig) error
	Delete(ctx context.Context, id uint) error
}

type hookExecutionRepo interface {
	ListByHook(ctx context.Context, hookID uint, limit, offset int) ([]modelsOrder.HookExecution, int64, error)
}

// HookConfigService manages hook configurations.
type HookConfigService struct {
	hookRepo      hookConfigRepo
	executionRepo hookExecutionRepo
}

// NewHookConfigService creates a new HookConfigService.
func NewHookConfigService(hookRepo hookConfigRepo, execRepo hookExecutionRepo) *HookConfigService {
	return &HookConfigService{
		hookRepo:      hookRepo,
		executionRepo: execRepo,
	}
}

func (s *HookConfigService) List(ctx context.Context) ([]modelsOrder.HookConfig, error) {
	return s.hookRepo.List(ctx)
}

func (s *HookConfigService) Get(ctx context.Context, id uint) (*modelsOrder.HookConfig, error) {
	return s.hookRepo.Get(ctx, id)
}

func (s *HookConfigService) Create(ctx context.Context, hook *modelsOrder.HookConfig) error {
	hook.CreatedAt = time.Now()
	hook.UpdatedAt = time.Now()
	return s.hookRepo.Create(ctx, hook)
}

func (s *HookConfigService) Update(ctx context.Context, hook *modelsOrder.HookConfig) error {
	return s.hookRepo.Update(ctx, hook)
}

func (s *HookConfigService) Delete(ctx context.Context, id uint) error {
	return s.hookRepo.Delete(ctx, id)
}

func (s *HookConfigService) ListExecutionsByHook(ctx context.Context, hookID uint, limit, offset int) ([]modelsOrder.HookExecution, int64, error) {
	return s.executionRepo.ListByHook(ctx, hookID, limit, offset)
}

// FindActiveByEvent 返回指定事件的启用 Hook 配置
func (s *HookConfigService) FindActiveByEvent(ctx context.Context, eventName string) ([]modelsOrder.HookConfig, error) {
	all, err := s.hookRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]modelsOrder.HookConfig, 0)
	for _, h := range all {
		if h.Status == "active" && h.EventName == eventName {
			out = append(out, h)
		}
	}
	return out, nil
}
