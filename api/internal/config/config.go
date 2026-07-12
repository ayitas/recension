package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultAPIKey        = "dev-api-key"
	defaultPassword      = "dev-password"
	defaultSessionSecret = "recension-dev-secret-change-me"
)

type Config struct {
	Env               string
	Port              string
	BootstrapAPIKey   string
	BootstrapPassword string
	Bootstrap         bool
	AllowSignup       bool
	DatabaseURL       string
	SessionSecret     string
	SessionTTL        time.Duration
	CORSOrigins       []string
	MaxBodyBytes      int64
	// Object storage (MinIO / S3). Empty endpoint disables blob APIs.
	S3Endpoint    string
	S3AccessKey   string
	S3SecretKey   string
	S3Bucket      string
	S3UseSSL      bool
	MaxBlobBytes  int64
}

func Load() (Config, error) {
	env := strings.ToLower(envOr("RECENSION_ENV", "development"))
	isProd := env == "production"

	cfg := Config{
		Env:               env,
		Port:              envOr("RECENSION_PORT", "8080"),
		BootstrapAPIKey:   envOr("RECENSION_API_KEY", defaultAPIKey),
		BootstrapPassword: envOr("RECENSION_BOOTSTRAP_PASSWORD", defaultPassword),
		Bootstrap:         envBool("RECENSION_BOOTSTRAP", !isProd),
		AllowSignup:       envBool("RECENSION_ALLOW_SIGNUP", !isProd),
		DatabaseURL: envOr(
			"RECENSION_DATABASE_URL",
			"postgres://recension:recension@localhost:5432/recension?sslmode=disable",
		),
		SessionSecret: envOr("RECENSION_SESSION_SECRET", defaultSessionSecret),
		SessionTTL:    envDuration("RECENSION_SESSION_TTL", 7*24*time.Hour),
		CORSOrigins:   envList("RECENSION_CORS_ORIGINS", defaultCORS(isProd)),
		MaxBodyBytes:  envInt64("RECENSION_MAX_BODY_BYTES", 8<<20),
		S3Endpoint:    envOr("RECENSION_S3_ENDPOINT", "localhost:9000"),
		S3AccessKey:   envOr("RECENSION_S3_ACCESS_KEY", "recension"),
		S3SecretKey:   envOr("RECENSION_S3_SECRET_KEY", "recensionsecret"),
		S3Bucket:      envOr("RECENSION_S3_BUCKET", "recension"),
		S3UseSSL:      envBool("RECENSION_S3_USE_SSL", false),
		MaxBlobBytes:  envInt64("RECENSION_MAX_BLOB_BYTES", 64<<20),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) BlobsEnabled() bool {
	return c.S3Endpoint != ""
}

func (c Config) Validate() error {
	if c.Env != "production" {
		return nil
	}
	if c.BootstrapAPIKey == "" || c.BootstrapAPIKey == defaultAPIKey {
		return fmt.Errorf("production requires a non-default RECENSION_API_KEY")
	}
	if c.BootstrapPassword == "" || c.BootstrapPassword == defaultPassword {
		return fmt.Errorf("production requires a non-default RECENSION_BOOTSTRAP_PASSWORD")
	}
	if c.SessionSecret == "" || c.SessionSecret == defaultSessionSecret {
		return fmt.Errorf("production requires a non-default RECENSION_SESSION_SECRET")
	}
	if c.Bootstrap {
		return fmt.Errorf("production refuses RECENSION_BOOTSTRAP=true; create users via signup or ops")
	}
	return nil
}

func (c Config) IsProduction() bool {
	return c.Env == "production"
}

func defaultCORS(isProd bool) []string {
	if isProd {
		return nil
	}
	return []string{"*"}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envInt64(key string, fallback int64) int64 {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

func envList(key string, fallback []string) []string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}
