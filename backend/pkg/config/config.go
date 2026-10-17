// Package config provides centralized configuration structures and loading logic
// for the Prim backend service, integrating YAML file definitions and environment variables.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// DatabaseConfig holds connection settings for PostgreSQL database instances.
type DatabaseConfig struct {
	// DBHost is the network hostname or IP address of the PostgreSQL server.
	DBHost string `yaml:"host" env:"DB_HOST" env-default:"localhost"`

	// DBPort is the network port for PostgreSQL connections (standard default: 5432).
	DBPort string `yaml:"port" env:"DB_PORT" env-default:"5432"`

	// DBUser is the authentication username for the database connection.
	DBUser string `yaml:"user" env:"DB_USER" env-default:"prim"`

	// DBPassword is the authentication password for the database connection.
	DBPassword string `yaml:"password" env:"DB_PASSWORD" env-default:"prim"`

	// DBName specifies the target PostgreSQL database name.
	DBName string `yaml:"name" env:"DB_NAME" env-default:"prim"`
	// SSLMode defines the SSL connection mode ("disable", "require", "verify-ca", "verify-full").
	SSLMode string `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:"disable"`
}

// RedisConfig holds connection and database index configuration for Redis cache/rate-limiter.
type RedisConfig struct {
	// Host is the Redis server hostname or IP address.
	Host string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`

	// Port is the Redis server network port (standard default: 6379).
	Port int `yaml:"port" env:"REDIS_PORT" env-default:"6379"`

	// Password is the optional authentication password for Redis AUTH.
	Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`

	// DB specifies the logical Redis database index to select (default: 0).
	DB int `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

// KafkaConfig holds bootstrap broker addresses and consumer group configuration.
type KafkaConfig struct {
	// Brokers is a comma-separated or list of seed Kafka broker host:port addresses.
	Brokers []string `yaml:"brokers" env:"KAFKA_BROKERS" env-separator:"," env-default:"localhost:9092"`

	// GroupID defines the default consumer group identification across workers.
	GroupID string `yaml:"group_id" env:"KAFKA_GROUP_ID" env-default:"prim-notification-workers"`
}

// SMTPConfig encapsulates outbound email server credentials and settings.
type SMTPConfig struct {
	// Host is the SMTP server hostname (e.g., smtp.mailgun.org, smtp.sendgrid.net).
	Host string `yaml:"host" env:"SMTP_HOST" env-default:""`

	// Port is the SMTP server port (e.g., 587 for STARTTLS, 465 for SSL, 25 for plain).
	Port int `yaml:"port" env:"SMTP_PORT" env-default:"0"`

	// Username is the SMTP authentication username.
	Username string `yaml:"username" env:"SMTP_USERNAME" env-default:""`

	// Password is the SMTP authentication password or API key.
	Password string `yaml:"password" env:"SMTP_PASSWORD" env-default:""`

	// FromEmail is the envelope sender address.
	FromEmail string `yaml:"from_email" env:"SMTP_FROM_EMAIL" env-default:""`

	// FromName is the human-readable display name of the email sender.
	FromName string `yaml:"from_name" env:"SMTP_FROM_NAME" env-default:"Prim"`
}

// Secrets contains sensitive cryptographic keys used by the authentication flow.
// Default fallback values are strictly prohibited in production environments.
type Secrets struct {
	// OTPHashPepper is the server-side pepper keying HMAC-SHA-256 for OTP challenge hashes.
	// It must be kept secret; it prevents precomputation attacks over the small OTP space
	// if Redis contents are exposed. Never use the development default in production.
	OTPHashPepper string `yaml:"otp_hash_pepper" env:"OTP_HASH_PEPPER" env-default:"dev-otp-pepper-change-me-in-production"`
}

// ClientConfig holds frontend application URLs for CORS, redirects, and transactional email links.
type ClientConfig struct {
	// BaseURL is the origin URL of the consumer-facing frontend app.
	BaseURL string `yaml:"base_url" env:"CLIENT_BASE_URL" env-default:"http://localhost:3000"`
}

// MinioConfig holds object storage connection credentials for S3/MinIO compatible stores.
type MinioConfig struct {
	// Endpoint is the internal network endpoint for MinIO operations.
	Endpoint string `yaml:"endpoint" env:"MINIO_ENDPOINT" env-default:"minio:9000"`

	// AccessKey is the S3 access key ID.
	AccessKey string `yaml:"access_key" env:"MINIO_ACCESS_KEY" env-default:"admin"`

	// SecretKey is the S3 secret access key.
	SecretKey string `yaml:"secret_key" env:"MINIO_SECRET_KEY" env-default:"supersecret"`

	// PublicURL is the publicly routable asset download URL prefix.
	PublicURL string `yaml:"public_url" env:"MINIO_PUBLIC_URL" env-default:"http://localhost:9000"`
}

// AuthConfig defines token lifespans and session retention durations.
type AuthConfig struct {
	// ChallengeTTL is the derived time.Duration computed from ChallengeTTLMinutes.
	ChallengeTTL time.Duration

	// SessionTTL is the derived time.Duration computed from SessionTTLDays.
	SessionTTL time.Duration

	// ChallengeTTLMinutes defines auth challenge expiration in minutes. Must be > 0.
	ChallengeTTLMinutes int `yaml:"challenge_ttl_minutes" env:"AUTH_CHALLENGE_TTL_MINUTES" env-default:"5"`

	// SessionTTLDays defines user refresh session lifetime in days. Must be > 0.
	SessionTTLDays int `yaml:"session_ttl_days" env:"AUTH_SESSION_TTL_DAYS" env-default:"30"`
}

// RateLimitConfig configures request quotas and sliding window durations for sensitive endpoints.
type RateLimitConfig struct {
	// AuthStartRequests is maximum allowed auth-start attempts within AuthStartWindow.
	AuthStartRequests int64 `yaml:"auth_start_requests" env:"RL_AUTH_START_REQ" env-default:"5"`
	// AuthStartWindow is the derived window duration.
	AuthStartWindow time.Duration
	// AuthStartWindowMin defines duration in minutes (must be > 0).
	AuthStartWindowMin int `yaml:"auth_start_window_minutes" env:"RL_AUTH_START_MIN" env-default:"1"`

	// AuthResendRequests is maximum allowed auth code resend attempts within AuthResendWindow.
	AuthResendRequests int64 `yaml:"auth_resend_requests" env:"RL_AUTH_RESEND_REQ" env-default:"3"`
	// AuthResendWindow is the derived window duration.
	AuthResendWindow time.Duration
	// AuthResendWindowMin defines duration in minutes (must be > 0).
	AuthResendWindowMin int `yaml:"auth_resend_window_minutes" env:"RL_AUTH_RESEND_MIN" env-default:"1"`

	// AuthVerifyRequests is maximum allowed verification attempts within AuthVerifyWindow.
	AuthVerifyRequests int64 `yaml:"auth_verify_requests" env:"RL_AUTH_VERIFY_REQ" env-default:"10"`
	// AuthVerifyWindow is the derived window duration.
	AuthVerifyWindow time.Duration
	// AuthVerifyWindowMin defines duration in minutes (must be > 0).
	AuthVerifyWindowMin int `yaml:"auth_verify_window_minutes" env:"RL_AUTH_VERIFY_MIN" env-default:"1"`
}

// ServerConfig configures HTTP listener and environment options.
type ServerConfig struct {
	// Port is the HTTP listener port.
	Port string `yaml:"port" env:"HTTP_PORT" env-default:"8080"`

	// Env is the runtime environment ("development", "staging", "production").
	Env string `yaml:"env" env:"APP_ENV" env-default:"development"`

	// AllowedOrigins is the slice of allowed CORS origin strings.
	AllowedOrigins []string `yaml:"allowed_origins" env:"CORS_ALLOWED_ORIGINS" env-separator:","`
}

// Config aggregates all application configuration sub-trees into a single object.
type Config struct {
	// ServerCfg holds nested server configuration.
	ServerCfg ServerConfig `yaml:"server"`

	// ClientCfg holds client app URLs.
	ClientCfg ClientConfig `yaml:"client"`

	// DBCfg holds PostgreSQL connection parameters.
	DBCfg DatabaseConfig `yaml:"database"`

	// RedisCfg holds Redis caching and rate-limiting connection parameters.
	RedisCfg RedisConfig `yaml:"redis"`

	// KafkaCfg holds Kafka broker topology parameters.
	KafkaCfg KafkaConfig `yaml:"kafka"`

	// SMTPCfg holds outbound mail transport configuration.
	SMTPCfg SMTPConfig `yaml:"smtp"`

	// KeysCfg holds cryptographic keys used by the authentication flow.
	KeysCfg Secrets `yaml:"secrets"`

	// SvPort is maintained for backward compatibility. Canonical source is ServerCfg.Port.
	SvPort string `yaml:"-" json:"-"`

	// MinioCfg holds S3/MinIO bucket and endpoint configurations.
	MinioCfg MinioConfig `yaml:"minio"`

	// AuthCfg holds token and session lifecycle parameters.
	AuthCfg AuthConfig `yaml:"auth"`

	// RateLimitCfg holds throttling quotas and windows.
	RateLimitCfg RateLimitConfig `yaml:"rate_limit"`

	// AllowedOrigins is maintained for backward compatibility. Canonical source is ServerCfg.AllowedOrigins.
	AllowedOrigins []string `yaml:"-" json:"-"`

	// Env is maintained for backward compatibility. Canonical source is ServerCfg.Env.
	Env string `yaml:"-" json:"-"`

	// IsProduction is a computed flag indicating if Env equals "production".
	IsProduction bool
}

// Load reads configuration from a YAML file (if provided or found at default paths)
// and overrides values from environment variables.
func Load(configPath ...string) (Config, error) {
	var cfg Config

	targetPath := findConfigPath(configPath...)
	if targetPath != "" {
		if err := cleanenv.ReadConfig(targetPath, &cfg); err != nil {
			return cfg, fmt.Errorf("read config file (%s): %w", targetPath, err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return cfg, fmt.Errorf("read config environment: %w", err)
		}
	}

	// Synchronize backward-compatible top-level fields from canonical ServerCfg
	cfg.SvPort = cfg.ServerCfg.Port
	cfg.Env = cfg.ServerCfg.Env
	cfg.AllowedOrigins = cfg.ServerCfg.AllowedOrigins

	// Fallback CORS allowed origins if empty
	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"http://localhost:3000"}
		cfg.ServerCfg.AllowedOrigins = cfg.AllowedOrigins
	}

	// Post-processing durations & environment flags
	cfg.IsProduction = strings.EqualFold(cfg.Env, "production")

	// Validate production secrets
	if cfg.IsProduction {
		if cfg.KeysCfg.OTPHashPepper == "" || cfg.KeysCfg.OTPHashPepper == "dev-otp-pepper-change-me-in-production" {
			return cfg, fmt.Errorf("insecure default OTP hash pepper must not be used in production")
		}
	}

	// Validate AuthConfig durations
	if cfg.AuthCfg.ChallengeTTLMinutes <= 0 {
		return cfg, fmt.Errorf("auth challenge_ttl_minutes must be greater than 0")
	}
	if cfg.AuthCfg.SessionTTLDays <= 0 {
		return cfg, fmt.Errorf("auth session_ttl_days must be greater than 0")
	}

	// Validate RateLimitConfig windows
	if cfg.RateLimitCfg.AuthStartWindowMin <= 0 {
		return cfg, fmt.Errorf("rate_limit auth_start_window_minutes must be greater than 0")
	}
	if cfg.RateLimitCfg.AuthResendWindowMin <= 0 {
		return cfg, fmt.Errorf("rate_limit auth_resend_window_minutes must be greater than 0")
	}
	if cfg.RateLimitCfg.AuthVerifyWindowMin <= 0 {
		return cfg, fmt.Errorf("rate_limit auth_verify_window_minutes must be greater than 0")
	}

	// Derive time.Duration from minutes/days integer fields
	cfg.AuthCfg.ChallengeTTL = time.Duration(cfg.AuthCfg.ChallengeTTLMinutes) * time.Minute
	cfg.AuthCfg.SessionTTL = time.Duration(cfg.AuthCfg.SessionTTLDays) * 24 * time.Hour

	cfg.RateLimitCfg.AuthStartWindow = time.Duration(cfg.RateLimitCfg.AuthStartWindowMin) * time.Minute
	cfg.RateLimitCfg.AuthResendWindow = time.Duration(cfg.RateLimitCfg.AuthResendWindowMin) * time.Minute
	cfg.RateLimitCfg.AuthVerifyWindow = time.Duration(cfg.RateLimitCfg.AuthVerifyWindowMin) * time.Minute

	return cfg, nil
}

// MustLoad calls Load and panics if loading or validating configuration fails.
func MustLoad(configPath ...string) Config {
	cfg, err := Load(configPath...)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return cfg
}

// findConfigPath resolves the configuration file path from explicit parameters,
// the CONFIG_PATH environment variable, or fallback relative directory paths.
func findConfigPath(explicitPaths ...string) string {
	// 1. Check explicitly passed file paths first
	for _, p := range explicitPaths {
		if p != "" && fileExists(p) {
			return p
		}
	}

	// 2. Check CONFIG_PATH environment variable override
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" && fileExists(envPath) {
		return envPath
	}

	// 3. Search relative to current working directory and walk up parents
	var searchDirs []string
	if wd, err := os.Getwd(); err == nil {
		searchDirs = append(searchDirs, wd)
		dir := wd
		for i := 0; i < 4; i++ {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			searchDirs = append(searchDirs, parent)
			dir = parent
		}
	}

	if exePath, err := os.Executable(); err == nil {
		searchDirs = append(searchDirs, filepath.Dir(exePath))
		searchDirs = append(searchDirs, filepath.Dir(filepath.Dir(exePath)))
	}

	candidates := []string{
		"configs/config.yaml",
		"configs/config.yml",
		"config.yaml",
		"config.yml",
	}

	for _, dir := range searchDirs {
		for _, c := range candidates {
			p := filepath.Join(dir, c)
			if fileExists(p) {
				return p
			}
		}
	}

	// 4. Fallback search paths relative to execution directory
	defaults := []string{
		"configs/config.yaml",
		"configs/config.yml",
		"config.yaml",
		"config.yml",
		"../configs/config.yaml",
		"../../configs/config.yaml",
	}

	for _, p := range defaults {
		if fileExists(p) {
			return p
		}
	}

	return ""
}

// fileExists checks whether the specified path exists and is a regular file (not a directory).
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
