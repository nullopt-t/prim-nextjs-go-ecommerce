package config

import (
	"os"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type DatabaseConfig struct {
	DBHost     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
	DBPort     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
	DBUser     string `yaml:"user" env:"DB_USER" env-default:"prim"`
	DBPassword string `yaml:"password" env:"DB_PASSWORD" env-default:"prim"`
	DBName     string `yaml:"name" env:"DB_NAME" env-default:"prim"`
}

type RedisConfig struct {
	Host     string `yaml:"host" env:"REDIS_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"REDIS_PORT" env-default:"6379"`
	Password string `yaml:"password" env:"REDIS_PASSWORD" env-default:""`
	DB       int    `yaml:"db" env:"REDIS_DB" env-default:"0"`
}

type SMTPConfig struct {
	Host     string `yaml:"host" env:"SMTP_HOST" env-default:""`
	Port     int    `yaml:"port" env:"SMTP_PORT" env-default:"0"`
	Username string `yaml:"username" env:"SMTP_USERNAME" env-default:""`
	Password string `yaml:"password" env:"SMTP_PASSWORD" env-default:""`
}

type Secrets struct {
	JwtAccessTokenSecretKey    string `yaml:"jwt_access_secret" env:"JWT_ACCESS_SECRET" env-default:"jwt-access-secret-key"`
	JwtRefreshTokenSecretKey   string `yaml:"jwt_refresh_secret" env:"JWT_REFRESH_SECRET" env-default:"jwt-refresh-secret-key"`
	JwtResetPassTokenSecretKey string `yaml:"jwt_reset_pass_secret" env:"JWT_RESET_PASS_SECRET" env-default:"jwt-reset-pass-secret-key"`
}

type ClientConfig struct {
	BaseURL string `yaml:"base_url" env:"CLIENT_BASE_URL" env-default:"http://localhost:3000"`
}

type MinioConfig struct {
	Endpoint  string `yaml:"endpoint" env:"MINIO_ENDPOINT" env-default:"minio:9000"`
	AccessKey string `yaml:"access_key" env:"MINIO_ACCESS_KEY" env-default:"admin"`
	SecretKey string `yaml:"secret_key" env:"MINIO_SECRET_KEY" env-default:"supersecret"`
	PublicURL string `yaml:"public_url" env:"MINIO_PUBLIC_URL" env-default:"http://localhost:9000"`
}

type AuthConfig struct {
	ChallengeTTL time.Duration
	SessionTTL   time.Duration

	ChallengeTTLMinutes int `yaml:"challenge_ttl_minutes" env:"AUTH_CHALLENGE_TTL_MINUTES" env-default:"5"`
	SessionTTLDays      int `yaml:"session_ttl_days" env:"AUTH_SESSION_TTL_DAYS" env-default:"30"`
}

type RateLimitConfig struct {
	AuthStartRequests  int64         `yaml:"auth_start_requests" env:"RL_AUTH_START_REQ" env-default:"5"`
	AuthStartWindow    time.Duration
	AuthStartWindowMin int           `yaml:"auth_start_window_minutes" env:"RL_AUTH_START_MIN" env-default:"1"`

	AuthResendRequests int64         `yaml:"auth_resend_requests" env:"RL_AUTH_RESEND_REQ" env-default:"3"`
	AuthResendWindow   time.Duration
	AuthResendWindowMin int          `yaml:"auth_resend_window_minutes" env:"RL_AUTH_RESEND_MIN" env-default:"1"`

	AuthVerifyRequests int64         `yaml:"auth_verify_requests" env:"RL_AUTH_VERIFY_REQ" env-default:"10"`
	AuthVerifyWindow   time.Duration
	AuthVerifyWindowMin int          `yaml:"auth_verify_window_minutes" env:"RL_AUTH_VERIFY_MIN" env-default:"1"`
}

type ServerConfig struct {
	Port           string   `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	Env            string   `yaml:"env" env:"APP_ENV" env-default:"development"`
	AllowedOrigins []string `yaml:"allowed_origins" env:"CORS_ALLOWED_ORIGINS" env-separator:","`
}

type Config struct {
	ServerCfg      ServerConfig    `yaml:"server"`
	ClientCfg      ClientConfig    `yaml:"client"`
	DBCfg          DatabaseConfig  `yaml:"database"`
	RedisCfg       RedisConfig     `yaml:"redis"`
	SMTPCfg        SMTPConfig      `yaml:"smtp"`
	KeysCfg        Secrets         `yaml:"secrets"`
	SvPort         string          `yaml:"port" env:"HTTP_PORT" env-default:"8080"`
	MinioCfg       MinioConfig     `yaml:"minio"`
	AuthCfg        AuthConfig      `yaml:"auth"`
	RateLimitCfg   RateLimitConfig `yaml:"rate_limit"`
	AllowedOrigins []string        `yaml:"allowed_origins" env:"CORS_ALLOWED_ORIGINS" env-separator:","`
	Env            string          `yaml:"env" env:"APP_ENV" env-default:"development"`
	IsProduction   bool
}

// Load reads configuration from a YAML file (if provided or found at default paths)
// and overrides values from environment variables.
func Load(configPath ...string) Config {
	var cfg Config

	targetPath := findConfigPath(configPath...)
	if targetPath != "" {
		_ = cleanenv.ReadConfig(targetPath, &cfg)
	} else {
		_ = cleanenv.ReadEnv(&cfg)
	}

	if cfg.ServerCfg.Port != "" {
		cfg.SvPort = cfg.ServerCfg.Port
	}
	if cfg.ServerCfg.Env != "" {
		cfg.Env = cfg.ServerCfg.Env
	}
	if len(cfg.ServerCfg.AllowedOrigins) > 0 {
		cfg.AllowedOrigins = cfg.ServerCfg.AllowedOrigins
	}

	// Post-processing durations & environment flags
	cfg.IsProduction = strings.EqualFold(cfg.Env, "production")
	cfg.AuthCfg.ChallengeTTL = time.Duration(cfg.AuthCfg.ChallengeTTLMinutes) * time.Minute
	cfg.AuthCfg.SessionTTL = time.Duration(cfg.AuthCfg.SessionTTLDays) * 24 * time.Hour

	cfg.RateLimitCfg.AuthStartWindow = time.Duration(cfg.RateLimitCfg.AuthStartWindowMin) * time.Minute
	cfg.RateLimitCfg.AuthResendWindow = time.Duration(cfg.RateLimitCfg.AuthResendWindowMin) * time.Minute
	cfg.RateLimitCfg.AuthVerifyWindow = time.Duration(cfg.RateLimitCfg.AuthVerifyWindowMin) * time.Minute

	if len(cfg.AllowedOrigins) == 0 {
		cfg.AllowedOrigins = []string{"http://localhost:3000"}
	}

	return cfg
}

func findConfigPath(explicitPaths ...string) string {
	for _, p := range explicitPaths {
		if p != "" && fileExists(p) {
			return p
		}
	}

	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" && fileExists(envPath) {
		return envPath
	}

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

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
