package config

import "github.com/m-mahmoud-alsaid/prim-backend/pkg/utils"

type DatabaseConfig struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type Secrets struct {
	JwtAccessTokenSecretKey    string
	JwtRefreshTokenSecretKey   string
	JwtResetPassTokenSecretKey string
}

type ClientConfig struct {
	BaseURL string
}

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	PublicURL string
}

type Config struct {
	ClientCfg *ClientConfig
	DBCfg     *DatabaseConfig
	RedisCfg  *RedisConfig
	SMTPCfg   *SMTPConfig
	KeysCfg   *Secrets
	SvPort    string
	MinioCfg  *MinioConfig
}

func Load() *Config {
	return &Config{
		MinioCfg: &MinioConfig{
			Endpoint:  utils.GetEnvOrDefault("MINIO_ENDPOINT", "minio:9000"),
			PublicURL: utils.GetEnvOrDefault("MINIO_PUBLIC_URL", "http://localhost:9000"),
			AccessKey: utils.GetEnvOrDefault("MINIO_ACCESS_KEY", "admin"),
			SecretKey: utils.GetEnvOrDefault("MINIO_SECRET_KEY", "supersecret"),
		},
		ClientCfg: &ClientConfig{
			BaseURL: utils.GetEnvOrDefault("CLIENT_BASE_URL", "http://localhost:8080"),
		},
		DBCfg: &DatabaseConfig{
			DBHost:     utils.GetEnvOrDefault("DB_HOST", "localhost"),
			DBPort:     utils.GetEnvOrDefault("DB_PORT", "5432"),
			DBUser:     utils.GetEnvOrDefault("DB_USER", "prim"),
			DBPassword: utils.GetEnvOrDefault("DB_PASSWORD", "prim"),
			DBName:     utils.GetEnvOrDefault("DB_NAME", "prim"),
		},
		RedisCfg: &RedisConfig{
			Host:     utils.GetEnvOrDefault("REDIS_HOST", "localhost"),
			Port:     utils.GetEnvAsInt("REDIS_PORT", 6379),
			Password: utils.GetEnvOrDefault("REDIS_PASSWORD", ""),
			DB:       utils.GetEnvAsInt("REDIS_DB", 0),
		},
		SMTPCfg: &SMTPConfig{
			Host:     utils.GetEnvOrDefault("SMTP_HOST", ""),
			Port:     utils.GetEnvAsInt("SMTP_PORT", 0),
			Username: utils.GetEnvOrDefault("SMTP_USERNAME", ""),
			Password: utils.GetEnvOrDefault("SMTP_PASSWORD", ""),
		},
		KeysCfg: &Secrets{
			JwtAccessTokenSecretKey:    utils.GetEnvOrDefault("JWT_ACCESS_SECRET", "jwt-access-secret-key"),
			JwtRefreshTokenSecretKey:   utils.GetEnvOrDefault("JWT_REFRESH_SECRET", "jwt-refresh-secret-key"),
			JwtResetPassTokenSecretKey: utils.GetEnvOrDefault("JWT_RESET_PASS_SECRET", "jwt-reset-pass-secret-key"),
		},
		SvPort: utils.GetEnvOrDefault("HTTP_PORT", "8080"),
	}
}
