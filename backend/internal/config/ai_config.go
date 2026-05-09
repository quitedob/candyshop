package config

// AIConfig holds AI-related configuration
type AIConfig struct {
	OpenAIAPIKey         string
	OpenAIModel          string
	OpenAIEmbeddingModel string
	DefaultLanguage      string
	SupportedLanguages   []string
	SemanticSearchMode   string
	// PublicRoutesDisabled 为 true 时关闭 /system 下无需登录的 AI（chatbot、推荐、语义搜索），防滥用与控成本
	PublicRoutesDisabled bool
}

// LoadAIConfig creates AIConfig from environment variables
func LoadAIConfig() AIConfig {
	return AIConfig{
		OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:          getEnv("OPENAI_MODEL", "gpt-4o"),
		OpenAIEmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		DefaultLanguage:      getEnv("AI_DEFAULT_LANGUAGE", "en"),
		SupportedLanguages:   []string{"en", "zh", "ar", "es", "fr", "de", "ja"},
		SemanticSearchMode:   getSemanticSearchMode(),
		PublicRoutesDisabled: getEnv("DISABLE_PUBLIC_AI_ROUTES", "") == "true",
	}
}

// IsEnabled returns true if AI is properly configured
func (c AIConfig) IsEnabled() bool {
	return c.OpenAIAPIKey != ""
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
