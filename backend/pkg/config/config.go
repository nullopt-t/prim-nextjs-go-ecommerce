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

type Config struct {
	ClientCfg *ClientConfig
	DBCfg     *DatabaseConfig
	RedisCfg  *RedisConfig
	SMTPCfg   *SMTPConfig
	KeysCfg   *Secrets
	SvPort    string
}

func Load() *Config {
	return &Config{
		ClientCfg: &ClientConfig{
			BaseURL: utils.GetEnvOrDefault("CLIENT_BASE_URL", "http://localhost:8080"),
		},
		DBCfg: &DatabaseConfig{
			DBHost:     utils.GetEnvOrDefault("DB_HOST", "localhost"),
			DBPort:     utils.GetEnvOrDefault("DB_PORT", "5432"),
			DBUser:     utils.GetEnvOrDefault("DB_USER", "bookshop"),
			DBPassword: utils.GetEnvOrDefault("DB_PASSWORD", "bookshop123"),
			DBName:     utils.GetEnvOrDefault("DB_NAME", "bookshop"),
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
