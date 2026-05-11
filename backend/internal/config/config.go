package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application
type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Email         EmailConfig
	Upload        UploadConfig
	Security      SecurityConfig
	JWT           JWTConfig
	AI            AIConfig
	KYB           KYBConfig
	ExchangeRates map[string]float64
}

// KYBConfig 客户激活与小额免审策略
type KYBConfig struct {
	// BypassMaxOrderUSD pending 用户允许下单/购物车结算的最大美元金额；0 表示关闭免审
	BypassMaxOrderUSD float64
	// BypassSampleMaxOrderUSD 当订单行全部为样品 SKU 时的更高免审额度；0 表示不启用
	BypassSampleMaxOrderUSD float64
	// SampleProductIDs 样品 SKU（产品 ID），逗号分隔配置于环境变量
	SampleProductIDs []string
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
	Environment     string
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
	// S3 / cloud storage
	StorageDriver string // "local", "s3", or "oss" (Alibaba Cloud OSS, S3-compatible)
	S3Bucket      string
	S3Region      string
	S3AccessKey   string
	S3SecretKey   string
	S3Endpoint    string // optional: MinIO / compatible endpoint
	S3CDNDomain   string // optional: CloudFront CDN domain for public URLs
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	CORSAllowedOrigins []string
	RateLimitPerMinute int
	// PublicAIRateLimitPerMinute 针对未登录 /system AI 路由的独立 IP 限流（更严）
	PublicAIRateLimitPerMinute int
	// PublicInquiryRateLimitPerMinute 针对 POST /public/inquiry 的独立 IP 限流（防刷询盘）
	PublicInquiryRateLimitPerMinute int
	EnableSwagger                   bool
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
			Environment:     getEnv("ENVIRONMENT", "development"),
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
			MaxFileSize:   int64(getEnvInt("MAX_FILE_SIZE", 5*1024*1024)), // 5MB
			AllowedTypes:  []string{"image/jpeg", "image/png", "image/webp", "application/pdf"},
			UploadPath:    getEnv("UPLOAD_PATH", "./uploads"),
			UploadURL:     getEnv("UPLOAD_URL", "/uploads"),
			StorageDriver: getEnv("STORAGE_DRIVER", "local"),
			S3Bucket:      getEnv("S3_BUCKET", ""),
			S3Region:      getEnv("S3_REGION", "us-east-1"),
			S3AccessKey:   getEnv("S3_ACCESS_KEY", ""),
			S3SecretKey:   getEnv("S3_SECRET_KEY", ""),
			S3Endpoint:    getEnv("S3_ENDPOINT", ""),
			S3CDNDomain:   getEnv("S3_CDN_DOMAIN", ""),
		},
		Security: SecurityConfig{
			CORSAllowedOrigins:              parseCORSOrigins(),
			RateLimitPerMinute:              getEnvInt("RATE_LIMIT_PER_MINUTE", 60),
			PublicAIRateLimitPerMinute:      getEnvInt("PUBLIC_AI_RATE_LIMIT_PER_MINUTE", 15),
			PublicInquiryRateLimitPerMinute: getEnvInt("PUBLIC_INQUIRY_RATE_LIMIT_PER_MINUTE", 10),
			EnableSwagger:                   getEnv("ENABLE_SWAGGER", "true") == "true",
		},
		JWT: JWTConfig{
			Secret:               getEnv("JWT_SECRET", ""),
			AccessTokenDuration:  getEnvInt("JWT_ACCESS_MINUTES", 15),
			RefreshTokenDuration: getEnvInt("JWT_REFRESH_DAYS", 7),
		},
		AI:            LoadAIConfig(),
		ExchangeRates: parseExchangeRates(getEnv("EXCHANGE_RATES", "")),
		KYB: KYBConfig{
			BypassMaxOrderUSD:       getEnvFloat("KYB_BYPASS_MAX_ORDER_USD", 0),
			BypassSampleMaxOrderUSD: getEnvFloat("KYB_BYPASS_SAMPLE_MAX_ORDER_USD", 0),
			SampleProductIDs:        parseCommaSeparated(getEnv("KYB_SAMPLE_PRODUCT_IDS", "")),
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
// parseCommaSeparated 解析逗号分隔 ID 列表（去空、去首尾空格）
func parseCommaSeparated(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

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
		log.Printf("Warning: invalid %s=%q, fallback to %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvFloat(key string, defaultValue float64) float64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		log.Printf("Warning: invalid %s=%q, fallback to %f", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

// parseExchangeRates parses EXCHANGE_RATES env var in format "EUR:0.92,CNY:7.24,GBP:0.79"
func parseExchangeRates(raw string) map[string]float64 {
	rates := make(map[string]float64)
	if raw == "" {
		return rates
	}
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		code := strings.ToUpper(strings.TrimSpace(parts[0]))
		rate, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil || rate <= 0 {
			continue
		}
		rates[code] = rate
	}
	return rates
}
