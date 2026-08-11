package eino

import (
	"context"
	"testing"

	"candypro/api/internal/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
)

func TestResolveDeepSeekModel(t *testing.T) {
	if got := resolveDeepSeekModel(config.AIConfig{}); got != DeepSeekV4Flash {
		t.Fatalf("empty model: expected %q, got %q", DeepSeekV4Flash, got)
	}
	if got := resolveDeepSeekModel(config.AIConfig{OpenAIModel: DeepSeekV4Pro}); got != DeepSeekV4Pro {
		t.Fatalf("explicit model: expected %q, got %q", DeepSeekV4Pro, got)
	}
}

func TestResolveDeepSeekBaseURL(t *testing.T) {
	if got := resolveDeepSeekBaseURL(config.AIConfig{}); got != DeepSeekBaseURL {
		t.Fatalf("empty base url: expected %q, got %q", DeepSeekBaseURL, got)
	}
	if got := resolveDeepSeekBaseURL(config.AIConfig{OpenAIBaseURL: "http://example.com/v1"}); got != "http://example.com/v1" {
		t.Fatalf("explicit base url: expected http://example.com/v1, got %q", got)
	}
}

func TestNewDeepSeekChatModel(t *testing.T) {
	cfg := config.AIConfig{OpenAIAPIKey: "sk-test", OpenAIModel: DeepSeekV4Flash}
	m, err := NewDeepSeekChatModel(context.Background(), cfg, nil)
	if err != nil {
		t.Fatalf("NewDeepSeekChatModel: %v", err)
	}
	if m.Model != DeepSeekV4Flash {
		t.Fatalf("expected model %q, got %q", DeepSeekV4Flash, m.Model)
	}
	if m.BaseURL != DeepSeekBaseURL {
		t.Fatalf("expected base url %q, got %q", DeepSeekBaseURL, m.BaseURL)
	}
	if m.ToolCallingChatModel == nil {
		t.Fatal("expected underlying chat model to be set")
	}
}

func TestNewDeepSeekChatModelDefaultsWithJSONOverride(t *testing.T) {
	// No model/base url in cfg: the class resolves DeepSeek defaults, and the
	// caller-provided ResponseFormat (JSON mode) is preserved.
	cfg := config.AIConfig{OpenAIAPIKey: "sk-test"}
	jsonFmt := &openai.ChatCompletionResponseFormat{Type: openai.ChatCompletionResponseFormatTypeJSONObject}
	m, err := NewDeepSeekChatModel(context.Background(), cfg, &openai.ChatModelConfig{ResponseFormat: jsonFmt})
	if err != nil {
		t.Fatalf("NewDeepSeekChatModel(json): %v", err)
	}
	if m.Model != DeepSeekV4Flash {
		t.Fatalf("expected default model %q, got %q", DeepSeekV4Flash, m.Model)
	}
	if m.BaseURL != DeepSeekBaseURL {
		t.Fatalf("expected default base url %q, got %q", DeepSeekBaseURL, m.BaseURL)
	}
}
