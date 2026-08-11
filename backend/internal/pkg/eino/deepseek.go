package eino

import (
	"context"
	"fmt"
	"time"

	"candypro/api/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// DeepSeek model constants. DeepSeek exposes an OpenAI-compatible HTTP API, so the
// eino-ext openai component is reused under the hood; this file is the DeepSeek
// "class" that owns the DeepSeek endpoint and model naming so callers never
// hardcode either.
const (
	// DeepSeekBaseURL is DeepSeek's OpenAI-compatible endpoint.
	DeepSeekBaseURL = "https://api.deepseek.com"
	// DeepSeekV4Flash is the default chat model (fast, low-cost).
	DeepSeekV4Flash = "deepseek-v4-flash"
	// DeepSeekV4Pro is the higher-capability chat model.
	DeepSeekV4Pro = "deepseek-v4-pro"
)

// DeepSeekChatModel is the DeepSeek chat-model class: a typed wrapper over the
// OpenAI-compatible client pre-configured for DeepSeek's API. It carries the
// resolved Model and BaseURL so callers can inspect the effective configuration,
// and it satisfies model.ToolCallingChatModel via the embedded value.
type DeepSeekChatModel struct {
	model.ToolCallingChatModel
	// Model is the resolved DeepSeek model name (e.g. "deepseek-v4-flash").
	Model string
	// BaseURL is the resolved API endpoint.
	BaseURL string
}

// NewDeepSeekChatModel builds a DeepSeek chat model from AI config. base is
// optional; callers may pre-fill fields such as ResponseFormat (JSON mode) or
// Temperature, while the DeepSeek-specific values (Model, APIKey, BaseURL,
// Timeout) are always resolved from cfg with safe defaults.
func NewDeepSeekChatModel(ctx context.Context, cfg config.AIConfig, base *openai.ChatModelConfig) (*DeepSeekChatModel, error) {
	chatCfg := openai.ChatModelConfig{}
	if base != nil {
		chatCfg = *base
	}
	chatCfg.Model = resolveDeepSeekModel(cfg)
	chatCfg.APIKey = cfg.OpenAIAPIKey
	chatCfg.BaseURL = resolveDeepSeekBaseURL(cfg)
	if chatCfg.Timeout <= 0 {
		chatCfg.Timeout = cfg.HTTPTimeout
		if chatCfg.Timeout <= 0 {
			chatCfg.Timeout = 90 * time.Second
		}
	}

	raw, err := openai.NewChatModel(ctx, &chatCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize deepseek chat model: %w", err)
	}
	return &DeepSeekChatModel{
		ToolCallingChatModel: raw,
		Model:                chatCfg.Model,
		BaseURL:              chatCfg.BaseURL,
	}, nil
}

// resolveDeepSeekModel returns the effective model, defaulting to DeepSeekV4Flash
// when OPENAI_MODEL is unset.
func resolveDeepSeekModel(cfg config.AIConfig) string {
	if cfg.OpenAIModel == "" {
		return DeepSeekV4Flash
	}
	return cfg.OpenAIModel
}

// resolveDeepSeekBaseURL returns the effective base URL, defaulting to
// DeepSeekBaseURL when OPENAI_BASE_URL is unset.
func resolveDeepSeekBaseURL(cfg config.AIConfig) string {
	if cfg.OpenAIBaseURL == "" {
		return DeepSeekBaseURL
	}
	return cfg.OpenAIBaseURL
}
