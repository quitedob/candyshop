package config

import (
	"fmt"
	"time"
)

// AIConfig holds AI-related configuration
type AIConfig struct {
	OpenAIAPIKey         string
	OpenAIBaseURL        string
	OpenAIModel          string
	OpenAIEmbeddingModel string
	DefaultLanguage      string
	SupportedLanguages   []string
	SemanticSearchMode   string
	RetryMaxAttempts     int
	RetryIntervalSec     int
	HTTPTimeout          time.Duration
	TranslateRetryMax    int
	// PublicRoutesDisabled 为 true 时关闭 /system 下无需登录的 AI（chatbot、推荐、语义搜索），防滥用与控成本
	PublicRoutesDisabled bool
	// RAGComplianceEnabled 为 true 时启用合规语料 RAG（compliance_lookup 工具），默认关闭
	RAGComplianceEnabled bool
}

// LoadAIConfig creates AIConfig from environment variables
func LoadAIConfig() AIConfig {
	return AIConfig{
		OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
			OpenAIBaseURL:        getEnv("OPENAI_BASE_URL", ""),
		OpenAIModel:          getEnv("OPENAI_MODEL", "deepseek-v4-flash"),
		OpenAIEmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		DefaultLanguage:      getEnv("AI_DEFAULT_LANGUAGE", "en"),
		SupportedLanguages:   []string{"en", "zh", "ar", "es", "fr", "de", "ja"},
		SemanticSearchMode:   getSemanticSearchMode(),
		RetryMaxAttempts:     getEnvInt("AI_RETRY_MAX_ATTEMPTS", 10),
		RetryIntervalSec:     getEnvInt("AI_RETRY_INTERVAL_SEC", 3),
		HTTPTimeout:          time.Duration(getEnvInt("AI_HTTP_TIMEOUT_SEC", 90)) * time.Second,
		TranslateRetryMax:    getEnvInt("AI_TRANSLATE_RETRY_MAX", 2),
		PublicRoutesDisabled: getEnv("DISABLE_PUBLIC_AI_ROUTES", "") == "true",
		RAGComplianceEnabled: getEnv("AI_RAG_COMPLIANCE_ENABLED", "") == "true",
	}
}

// IsEnabled returns true if AI is properly configured
func (c AIConfig) IsEnabled() bool {
	return c.OpenAIAPIKey != ""
}

// String returns a sanitized representation of AIConfig (hides sensitive fields).
func (c AIConfig) String() string {
	if c.OpenAIAPIKey == "" {
		return "AIConfig{APIKey: <empty>}"
	}
	return fmt.Sprintf("AIConfig{APIKey: %s***}", c.OpenAIAPIKey[:4])
}

// IsSemanticSearchRequired returns true if semantic search must be available.
func (c AIConfig) IsSemanticSearchRequired() bool {
	return c.SemanticSearchMode == "required"
}

// IsSemanticSearchDisabled returns true if semantic search is explicitly disabled.
func (c AIConfig) IsSemanticSearchDisabled() bool {
	return c.SemanticSearchMode == "disabled"
}

func getSemanticSearchMode() string {
	switch mode := getEnv("SEMANTIC_SEARCH_MODE", "auto"); mode {
	case "required", "disabled", "auto":
		return mode
	default:
		return "auto"
	}
}
