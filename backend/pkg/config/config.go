// Package config provides centralized configuration structures and loading logic
// for the Prim backend service, integrating YAML file definitions and environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// DatabaseConfig holds connection settings for PostgreSQL database instances.
type DatabaseConfig struct {
	// URL specifies the PostgreSQL connection URL (DATABASE_URL or DATABASE).
	URL string `yaml:"url" env:"DATABASE_URL" env-default:""`

	// Legacy discrete fields for test fixtures (no defaults)
	DBHost     string `yaml:"host" env:"DB_HOST" env-default:""`
	DBPort     string `yaml:"port" env:"DB_PORT" env-default:""`
	DBUser     string `yaml:"user" env:"DB_USER" env-default:""`
	DBPassword string `yaml:"password" env:"DB_PASSWORD" env-default:""`
	DBName     string `yaml:"name" env:"DB_NAME" env-default:""`
	SSLMode    string `yaml:"ssl_mode" env:"DB_SSL_MODE" env-default:""`

	// MaxConns is the maximum number of connections in the PostgreSQL pool.
	MaxConns int `yaml:"max_conns" env:"DB_MAX_CONNS" env-default:"25"`
	// MinConns is the minimum number of idle connections kept in the pool.
	MinConns int `yaml:"min_conns" env:"DB_MIN_CONNS" env-default:"5"`
	// MaxConnLifetime is the maximum time a pooled connection may be reused.
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime" env:"DB_MAX_CONN_LIFETIME" env-default:"1h"`
	// MaxConnIdleTime is the maximum time an idle connection may be kept in the pool.
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env:"DB_MAX_CONN_IDLE_TIME" env-default:"30m"`
}

// RedisConfig holds connection and database index configuration for Redis cache/rate-limiter.
type RedisConfig struct {
	// URL specifies the Redis connection URL (REDIS_URL or REDIS).
	URL string `yaml:"url" env:"REDIS_URL" env-default:""`

	// Legacy discrete fields for test fixtures (no defaults)
	Host     string `yaml:"host" env:"REDIS_HOST" env-default:""`
	Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"0"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
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

// SMSConfig configures outbound SMS delivery via Twilio or logging driver.
type SMSConfig struct {
	// Provider selects the SMS delivery provider ("twilio" or "log"). Defaults to "log".
	Provider string `yaml:"provider" env:"SMS_PROVIDER" env-default:"log"`

	// TwilioAccountSID is the Twilio Account SID.
	TwilioAccountSID string `yaml:"twilio_account_sid" env:"TWILIO_ACCOUNT_SID" env-default:""`

	// TwilioAuthToken is the Twilio Auth Token.
	TwilioAuthToken string `yaml:"twilio_auth_token" env:"TWILIO_AUTH_TOKEN" env-default:""`

	// TwilioFromPhone is the Twilio sender phone number (E.164 format, e.g. +1234567890).
	TwilioFromPhone string `yaml:"twilio_from_phone" env:"TWILIO_FROM_PHONE" env-default:""`
}

// PaymentConfig configures online payment gateway providers (Stripe).
type PaymentConfig struct {
	// Provider is the primary payment provider name (e.g. "stripe").
	Provider string `yaml:"provider" env:"PAYMENT_PROVIDER" env-default:"stripe"`

	// StripeSecretKey is the secret API key for Stripe operations (starts with sk_).
	StripeSecretKey string `yaml:"stripe_secret_key" env:"STRIPE_SECRET_KEY" env-default:""`

	// StripePublishableKey is the browser-facing publishable key (starts with pk_).
	StripePublishableKey string `yaml:"stripe_publishable_key" env:"STRIPE_PUBLISHABLE_KEY" env-default:""`

	// StripeWebhookSecret is the endpoint signing secret for webhook verification (starts with whsec_).
	StripeWebhookSecret string `yaml:"stripe_webhook_secret" env:"STRIPE_WEBHOOK_SECRET" env-default:""`
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
	// Endpoint is the internal network endpoint for MinIO/S3 operations.
	Endpoint string `yaml:"endpoint" env:"MINIO_ENDPOINT" env-default:""`

	// AccessKey is the S3 access key ID.
	AccessKey string `yaml:"access_key" env:"MINIO_ACCESS_KEY" env-default:""`

	// SecretKey is the S3 secret access key.
	SecretKey string `yaml:"secret_key" env:"MINIO_SECRET_KEY" env-default:""`

	// PublicURL is the publicly routable asset download URL prefix.
	PublicURL string `yaml:"public_url" env:"MINIO_PUBLIC_URL" env-default:"http://localhost:9000"`

	// Bucket is the default S3 bucket name.
	Bucket string `yaml:"bucket" env:"MINIO_BUCKET" env-default:""`

	// Region specifies the target S3 bucket region (defaults to us-east-1).
	Region string `yaml:"region" env:"MINIO_REGION" env-default:"us-east-1"`

	// UseSSL explicitly toggles HTTPS for S3 operations. If unset (""), auto-detected from endpoint.
	UseSSL string `yaml:"use_ssl" env:"MINIO_USE_SSL" env-default:""`

	// S3_* variants (AWS S3 / LocalStack). Take precedence over MINIO_* when set.
	S3Endpoint  string `yaml:"s3_endpoint" env:"S3_ENDPOINT" env-default:""`
	S3PublicURL string `yaml:"s3_public_url" env:"S3_PUBLIC_URL" env-default:""`
	S3AccessKey string `yaml:"s3_access_key" env:"S3_ACCESS_KEY" env-default:""`
	S3SecretKey string `yaml:"s3_secret_key" env:"S3_SECRET_KEY" env-default:""`
	S3Bucket    string `yaml:"s3_bucket" env:"S3_BUCKET" env-default:""`
	S3Region    string `yaml:"s3_region" env:"S3_REGION" env-default:""`
	S3UseSSL    string `yaml:"s3_use_ssl" env:"S3_USE_SSL" env-default:""`
}

// AuthConfig defines token lifespans and session retention durations.
type AuthConfig struct {
	// ChallengeTTL defines auth challenge expiration as a Go duration (e.g. 5m). Must be > 0.
	ChallengeTTL time.Duration `yaml:"challenge_ttl" env:"AUTH_CHALLENGE_TTL" env-default:"5m"`

	// SessionTTL defines user refresh session lifetime as a Go duration (e.g. 720h for 30 days). Must be > 0.
	SessionTTL time.Duration `yaml:"session_ttl" env:"AUTH_SESSION_TTL" env-default:"720h"`
}

// RateLimitConfig configures request quotas and sliding window durations for sensitive endpoints.
// Quotas are request counts; windows are Go durations with an explicit unit (e.g. 1m).
type RateLimitConfig struct {
	// AuthStartRequests is maximum allowed auth-start requests within AuthStartWindow.
	AuthStartRequests int64 `yaml:"auth_start_requests" env:"RL_AUTH_START_REQUESTS" env-default:"5"`
	// AuthStartWindow is the sliding window duration (e.g. 1m). Must be > 0.
	AuthStartWindow time.Duration `yaml:"auth_start_window" env:"RL_AUTH_START_WINDOW" env-default:"1m"`

	// AuthResendRequests is maximum allowed auth code resend requests within AuthResendWindow.
	AuthResendRequests int64 `yaml:"auth_resend_requests" env:"RL_AUTH_RESEND_REQUESTS" env-default:"3"`
	// AuthResendWindow is the sliding window duration (e.g. 1m). Must be > 0.
	AuthResendWindow time.Duration `yaml:"auth_resend_window" env:"RL_AUTH_RESEND_WINDOW" env-default:"1m"`

	// AuthVerifyRequests is maximum allowed verification requests within AuthVerifyWindow.
	AuthVerifyRequests int64 `yaml:"auth_verify_requests" env:"RL_AUTH_VERIFY_REQUESTS" env-default:"10"`
	// AuthVerifyWindow is the sliding window duration (e.g. 1m). Must be > 0.
	AuthVerifyWindow time.Duration `yaml:"auth_verify_window" env:"RL_AUTH_VERIFY_WINDOW" env-default:"1m"`

	// HistoryWriteRequests is maximum allowed history record/sync requests within HistoryWriteWindow.
	HistoryWriteRequests int64 `yaml:"history_write_requests" env:"RL_HISTORY_WRITE_REQUESTS" env-default:"120"`
	// HistoryWriteWindow is the sliding window duration (e.g. 1m). Must be > 0.
	HistoryWriteWindow time.Duration `yaml:"history_write_window" env:"RL_HISTORY_WRITE_WINDOW" env-default:"1m"`
}

// ServerConfig configures HTTP listener and environment options.
type ServerConfig struct {
	// Port is the HTTP listener port (standard: PORT, legacy fallback: HTTP_PORT).
	Port string `yaml:"port" env:"PORT" env-default:"8080"`

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

	// SMSCfg holds outbound SMS transport configuration.
	SMSCfg SMSConfig `yaml:"sms"`

	// PaymentCfg holds payment provider and webhook credentials.
	PaymentCfg PaymentConfig `yaml:"payment"`

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

// Load reads configuration from environment variables. Explicit YAML paths
// (an argument or CONFIG_PATH) are still honored for tooling/tests, but there
// is no implicit config.yaml discovery: services run env-only.
func Load(configPath ...string) (Config, error) {
	var cfg Config

	targetPath := ""
	for _, p := range configPath {
		if p != "" && fileExists(p) {
			targetPath = p
			break
		}
	}
	if targetPath == "" {
		if envPath := os.Getenv("CONFIG_PATH"); envPath != "" && fileExists(envPath) {
			targetPath = envPath
		}
	}

	if targetPath != "" {
		if err := cleanenv.ReadConfig(targetPath, &cfg); err != nil {
			return cfg, fmt.Errorf("read config file (%s): %w", targetPath, err)
		}
	} else {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return cfg, fmt.Errorf("read config environment: %w", err)
		}
	}

	// Fallback to DATABASE or REDIS environment aliases if URL is unset
	if cfg.DBCfg.URL == "" {
		if u := os.Getenv("DATABASE"); u != "" {
			cfg.DBCfg.URL = u
		}
	}
	if cfg.RedisCfg.URL == "" {
		if u := os.Getenv("REDIS"); u != "" {
			cfg.RedisCfg.URL = u
		}
	}

	// Fallback to legacy HTTP_PORT if PORT is unset
	if os.Getenv("PORT") == "" {
		if hp := os.Getenv("HTTP_PORT"); hp != "" {
			cfg.ServerCfg.Port = strings.TrimSpace(hp)
		}
	}

	// S3_* config falls back to MINIO_* when unset (backward compatibility).
	if cfg.MinioCfg.S3Endpoint == "" {
		cfg.MinioCfg.S3Endpoint = cfg.MinioCfg.Endpoint
	}
	if cfg.MinioCfg.S3PublicURL == "" {
		cfg.MinioCfg.S3PublicURL = cfg.MinioCfg.PublicURL
	}
	if cfg.MinioCfg.S3AccessKey == "" {
		cfg.MinioCfg.S3AccessKey = cfg.MinioCfg.AccessKey
	}
	if cfg.MinioCfg.S3SecretKey == "" {
		cfg.MinioCfg.S3SecretKey = cfg.MinioCfg.SecretKey
	}
	if cfg.MinioCfg.S3Bucket == "" {
		cfg.MinioCfg.S3Bucket = cfg.MinioCfg.Bucket
	}
	if cfg.MinioCfg.S3Region == "" {
		cfg.MinioCfg.S3Region = cfg.MinioCfg.Region
	}
	if cfg.MinioCfg.S3UseSSL == "" {
		cfg.MinioCfg.S3UseSSL = cfg.MinioCfg.UseSSL
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

	// Validate AuthConfig durations
	if cfg.AuthCfg.ChallengeTTL <= 0 {
		return cfg, fmt.Errorf("auth challenge_ttl must be greater than 0")
	}
	if cfg.AuthCfg.SessionTTL <= 0 {
		return cfg, fmt.Errorf("auth session_ttl must be greater than 0")
	}

	// Validate RateLimitConfig windows
	if cfg.RateLimitCfg.AuthStartWindow <= 0 {
		return cfg, fmt.Errorf("rate_limit auth_start_window must be greater than 0")
	}
	if cfg.RateLimitCfg.AuthResendWindow <= 0 {
		return cfg, fmt.Errorf("rate_limit auth_resend_window must be greater than 0")
	}
	if cfg.RateLimitCfg.AuthVerifyWindow <= 0 {
		return cfg, fmt.Errorf("rate_limit auth_verify_window must be greater than 0")
	}
	if cfg.RateLimitCfg.HistoryWriteWindow <= 0 {
		return cfg, fmt.Errorf("rate_limit history_write_window must be greater than 0")
	}

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

// fileExists checks whether the specified path exists and is a regular file (not a directory).
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
