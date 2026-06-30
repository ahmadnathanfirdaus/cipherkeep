package config

import (
	"strings"
	"testing"
)

func baseValid() *Config {
	return &Config{
		DatabaseURL:    "postgres://u:p@h:5432/db",
		JWTSecret:      strings.Repeat("a", 32),
		MasterPassword: strings.Repeat("b", 16),
	}
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(c *Config)
		wantErr string
	}{
		{"valid", func(c *Config) {}, ""},
		{"missing db", func(c *Config) { c.DatabaseURL = "" }, "DATABASE_URL"},
		{"short jwt", func(c *Config) { c.JWTSecret = "tooshort" }, "JWT_SECRET must be at least"},
		{"short master", func(c *Config) { c.MasterPassword = "short" }, "MASTER_PASSWORD must be at least"},
		{"placeholder jwt", func(c *Config) { c.JWTSecret = "change-me-change-me-change-me-1234" }, "JWT_SECRET looks like a placeholder"},
		{"placeholder master", func(c *Config) { c.MasterPassword = "placeholder-value-x" }, "MASTER_PASSWORD looks like a placeholder"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := baseValid()
			tc.mutate(c)
			err := c.validate()
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}
