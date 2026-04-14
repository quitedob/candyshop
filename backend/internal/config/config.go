package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Email    EmailConfig
	Upload   UploadConfig
	Security SecurityConfig
	JWT      JWTConfig
	AI       AIConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         string
	Environment  string
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime int // in minutes
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
}

// UploadConfig holds file upload configuration
type UploadConfig struct {
	MaxFileSize  int64
	AllowedTypes []string
	UploadPath   string
	UploadURL    string
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	CORSAllowedOrigins []string
	RateLimitPerMinute int
	EnableSwagger      bool
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret               string
	AccessTokenDuration  int // in minutes
	RefreshTokenDuration int // in days
}

// Load creates a new Config from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Environment:  getEnv("ENVIRONMENT", "development"),
			ReadTimeout:  getEnvInt("READ_TIMEOUT", 60),
			WriteTimeout: getEnvInt("WRITE_TIMEOUT", 60),
			IdleTimeout:  getEnvInt("IDLE_TIMEOUT", 120),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "candypro"),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", "candypro"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 10),
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 100),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 60),
		},
		Email: EmailConfig{
			SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
			SMTPPort:     getEnvInt("SMTP_PORT", 587),
			SMTPUser:     getEnv("SMTP_USER", ""),
			SMTPPassword: getEnv("SMTP_PASSWORD", ""),
			FromEmail:    getEnv("FROM_EMAIL", "info@candypro.com"),
			FromName:     getEnv("FROM_NAME", "CandyPro OEM"),
		},
		Upload: UploadConfig{
			MaxFileSize:  int64(getEnvInt("MAX_FILE_SIZE", 5*1024*1024)), // 5MB
			AllowedTypes: []string{"image/jpeg", "image/png", "image/webp", "application/pdf"},
			UploadPath:   getEnv("UPLOAD_PATH", "./uploads"),
			UploadURL:    getEnv("UPLOAD_URL", "/uploads"),
		},
		Security: SecurityConfig{
			CORSAllowedOrigins: parseCORSOrigins(),
			RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 60),
			EnableSwagger:      getEnv("ENABLE_SWAGGER", "true") == "true",
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", ""),
			AccessTokenDuration:  getEnvInt("JWT_ACCESS_MINUTES", 15),
			RefreshTokenDuration: getEnvInt("JWT_REFRESH_DAYS", 7),
		},
		AI: AIConfig{
			OpenAIAPIKey:         getEnv("OPENAI_API_KEY", ""),
			OpenAIModel:          getEnv("OPENAI_MODEL", "gpt-4o"),
			OpenAIEmbeddingModel: getEnv("OPENAI_EMBEDDING_MODEL", "text-embedding-3-small"),
		},
	}

	// Validate required fields
	if cfg.Server.Environment == "production" && cfg.Database.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required in production")
	}

	// SEC-3: Fail fast if JWT secret is empty or the old insecure default in ANY environment
	const defaultJWTSecret = "super_secret_candypro_key_change_in_production"
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == defaultJWTSecret {
		return nil, fmt.Errorf("JWT_SECRET environment variable must be set to a strong unique value (min 32 chars)")
	}
	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET is too short (%d chars); must be at least 32 characters", len(cfg.JWT.Secret))
	}

	// Issue 10: Validate CORS — credentials mode is incompatible with wildcard origins
	if cfg.Server.Environment == "production" {
		for _, origin := range cfg.Security.CORSAllowedOrigins {
			if origin == "*" {
				return nil, fmt.Errorf("wildcard CORS origin (*) is not allowed in production (credentials are enabled)")
			}
		}
	}

	return cfg, nil
}

// parseCORSOrigins reads CORS origins from CORS_ORIGINS env var (comma-separated)
// or falls back to FRONTEND_URL + localhost:3001.
func parseCORSOrigins() []string {
	raw := os.Getenv("CORS_ORIGINS")
	if raw != "" {
		parts := strings.Split(raw, ",")
		origins := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
		if len(origins) > 0 {
			return origins
		}
	}
	// Fallback
	return []string{
		getEnv("FRONTEND_URL", "http://localhost:3000"),
		"http://localhost:3001",
		"http://localhost:3002",
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
