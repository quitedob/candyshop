package retry

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

// ChatModel wraps a model.ToolCallingChatModel with automatic retry on failures.
// Default: 10 attempts, 3s interval. Configure via AI_RETRY_MAX_ATTEMPTS, AI_RETRY_INTERVAL_SECONDS env vars.
type ChatModel struct {
	inner     model.ToolCallingChatModel
	maxRetry  int
	interval  time.Duration
}

// New wraps an existing ToolCallingChatModel with retry logic.
// Reads AI_RETRY_MAX_ATTEMPTS (default 10) and AI_RETRY_INTERVAL_SECONDS (default 3) from env.
func New(inner model.ToolCallingChatModel, maxRetry int, intervalSec int) *ChatModel {
	if maxRetry <= 0 {
		maxRetry = 10
	}
	if intervalSec <= 0 {
		intervalSec = 3
	}
	return &ChatModel{
		inner:    inner,
		maxRetry: maxRetry,
		interval:  time.Duration(intervalSec) * time.Second,
	}
}

func (r *ChatModel) WithTools(tools []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	newInner, err := r.inner.WithTools(tools)
	if err != nil {
		return nil, err
	}
	return &ChatModel{inner: newInner, maxRetry: r.maxRetry, interval: r.interval}, nil
}

func (r *ChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (output *schema.Message, err error) {
	for attempt := 1; attempt <= r.maxRetry; attempt++ {
		output, err = r.inner.Generate(ctx, input, opts...)
		if err == nil {
			return output, nil
		}
		log.Printf("[retry] Generate attempt %d/%d failed: %v", attempt, r.maxRetry, err)
		if attempt < r.maxRetry {
			time.Sleep(r.interval)
		}
	}
	return nil, err
}

func (r *ChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (output *schema.StreamReader[*schema.Message], err error) {
	for attempt := 1; attempt <= r.maxRetry; attempt++ {
		output, err = r.inner.Stream(ctx, input, opts...)
		if err == nil {
			return output, nil
		}
		log.Printf("[retry] Stream attempt %d/%d failed: %v", attempt, r.maxRetry, err)
		if attempt < r.maxRetry {
			time.Sleep(r.interval)
		}
	}
	return nil, err
}
