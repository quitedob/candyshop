package config

// AIConfig holds AI-related configuration
type AIConfig struct {
	OpenAIAPIKey         string
	OpenAIModel          string
	OpenAIEmbeddingModel string
	DefaultLanguage      string
	SupportedLanguages   []string
}

// LoadAIConfig creates AIConfig from environment variables
func LoadAIConfig() AIConfig {
	return AIConfig{
		OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:          getEnv("OPENAI_MODEL", "gpt-4o"),
		OpenAIEmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		DefaultLanguage:      getEnv("AI_DEFAULT_LANGUAGE", "en"),
		SupportedLanguages:   []string{"en", "zh", "ar", "es", "fr", "de", "ja"},
	}
}

// IsEnabled returns true if AI is properly configured
func (c AIConfig) IsEnabled() bool {
	return c.OpenAIAPIKey != ""
}
