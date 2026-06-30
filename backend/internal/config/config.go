package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	AppEnv             string
	HTTPPort           string
	DatabaseURL        string
	MasterPassword     string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	CORSAllowedOrigins []string
	TrustedProxies     []string
	MaxBodyBytes       int64
	LogLevel           string
}

// Minimum lengths for the secrets that protect the whole system.
const (
	minJWTSecretLen      = 32
	minMasterPasswordLen = 16
	defaultMaxBodyBytes  = 1 << 20 // 1 MiB
)

// IsProduction reports whether the app runs in a production environment.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production")
}

// Load reads configuration from the environment and validates required values.
// It fails fast (returns an error) if any required secret is missing.
func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:         getEnv("APP_ENV", "development"),
		HTTPPort:       getEnv("HTTP_PORT", "8080"),
		DatabaseURL:    os.Getenv("DATABASE_URL"),
		MasterPassword: os.Getenv("MASTER_PASSWORD"),
		JWTSecret:      os.Getenv("JWT_SECRET"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}

	accessTTL, err := parseDuration("ACCESS_TOKEN_TTL", "15m")
	if err != nil {
		return nil, err
	}
	cfg.AccessTokenTTL = accessTTL

	refreshTTL, err := parseDuration("REFRESH_TOKEN_TTL", "168h")
	if err != nil {
		return nil, err
	}
	cfg.RefreshTokenTTL = refreshTTL

	origins := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173")
	cfg.CORSAllowedOrigins = splitAndTrim(origins)

	// Reverse proxies whose X-Forwarded-For we trust. Empty (default) means trust
	// none: the client IP is taken from the direct connection, so a client cannot
	// spoof its IP to evade rate limiting.
	cfg.TrustedProxies = splitAndTrim(getEnv("TRUSTED_PROXIES", ""))

	maxBody, err := parseInt64("MAX_BODY_BYTES", defaultMaxBodyBytes)
	if err != nil {
		return nil, err
	}
	cfg.MaxBodyBytes = maxBody

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// validate enforces presence of required secrets.
func (c *Config) validate() error {
	var missing []string
	if c.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if c.MasterPassword == "" {
		missing = append(missing, "MASTER_PASSWORD")
	}
	if c.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	// Reject weak or placeholder secrets: a guessable JWT_SECRET allows token
	// forgery, and a weak MASTER_PASSWORD makes the wrapped DEK brute-forceable.
	if len(c.JWTSecret) < minJWTSecretLen {
		return fmt.Errorf("JWT_SECRET must be at least %d characters", minJWTSecretLen)
	}
	if len(c.MasterPassword) < minMasterPasswordLen {
		return fmt.Errorf("MASTER_PASSWORD must be at least %d characters", minMasterPasswordLen)
	}
	if isPlaceholderSecret(c.JWTSecret) {
		return fmt.Errorf("JWT_SECRET looks like a placeholder; set a strong random value")
	}
	if isPlaceholderSecret(c.MasterPassword) {
		return fmt.Errorf("MASTER_PASSWORD looks like a placeholder; set a strong random value")
	}
	return nil
}

// isPlaceholderSecret flags obvious non-secret defaults so they never ship.
func isPlaceholderSecret(s string) bool {
	lower := strings.ToLower(s)
	for _, bad := range []string{"change-me", "changeme", "change_me", "placeholder", "example", "your-secret", "yoursecret"} {
		if strings.Contains(lower, bad) {
			return true
		}
	}
	return false
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(key, fallback string) (time.Duration, error) {
	raw := getEnv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s=%q: %w", key, raw, err)
	}
	return d, nil
}

func parseInt64(key string, fallback int64) (int64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid value for %s=%q: must be a positive integer", key, raw)
	}
	return v, nil
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
