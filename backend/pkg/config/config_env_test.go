package config

import (
	"testing"
	"time"
)

// setProductionEnv sets a secure production baseline. Individual tests override
// one variable to prove that each insecure value is rejected.
func setProductionEnv(t *testing.T) {
	t.Helper()
	t.Setenv("APP_ENV", "production")
	t.Setenv("OTP_HASH_PEPPER", "test-pepper-not-the-default")
	t.Setenv("DATABASE_URL", "postgresql://user:pass@localhost:5432/db")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("MINIO_SECRET_KEY", "secure-minio-secret")
	t.Setenv("CORS_ALLOWED_ORIGINS", "https://app.example.com")
	t.Setenv("CLIENT_BASE_URL", "https://app.example.com")
	t.Setenv("AUTH_CHALLENGE_TTL", "5m")
	t.Setenv("AUTH_SESSION_TTL", "720h")
	t.Setenv("RL_AUTH_START_WINDOW", "1m")
	t.Setenv("RL_AUTH_RESEND_WINDOW", "1m")
	t.Setenv("RL_AUTH_VERIFY_WINDOW", "1m")
}

func TestLoadEnvOnlyProduction(t *testing.T) {
	setProductionEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected secure production config to load, got: %v", err)
	}
	if !cfg.IsProduction {
		t.Fatal("expected IsProduction to be true")
	}
	if cfg.ServerCfg.Port != "8080" {
		t.Fatalf("expected default HTTP port 8080, got %q", cfg.ServerCfg.Port)
	}
	if cfg.DBCfg.URL != "postgresql://user:pass@localhost:5432/db" {
		t.Fatal("environment value for DATABASE_URL was not applied")
	}
	if cfg.RedisCfg.URL != "redis://localhost:6379" {
		t.Fatal("environment value for REDIS_URL was not applied")
	}
	if cfg.AuthCfg.ChallengeTTL != 5*time.Minute {
		t.Fatalf("challenge TTL = %v, want 5m", cfg.AuthCfg.ChallengeTTL)
	}
	if cfg.AuthCfg.SessionTTL != 720*time.Hour {
		t.Fatalf("session TTL = %v, want 720h", cfg.AuthCfg.SessionTTL)
	}
	if cfg.RateLimitCfg.AuthStartWindow != time.Minute {
		t.Fatalf("auth start window = %v, want 1m", cfg.RateLimitCfg.AuthStartWindow)
	}
}
