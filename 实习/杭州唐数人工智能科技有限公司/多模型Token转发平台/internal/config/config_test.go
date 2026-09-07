package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain removes any APP_* variables from the process environment so test
// outcomes do not depend on the shell or container the suite runs in
// (docker-compose.test.yml sets APP_DATABASE__URL etc. for integration tests).
func TestMain(m *testing.M) {
	for _, env := range os.Environ() {
		if key, _, ok := strings.Cut(env, "="); ok && strings.HasPrefix(key, "APP_") {
			os.Unsetenv(key)
		}
	}
	os.Exit(m.Run())
}

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("APP_DATABASE__URL", "postgres://localhost/db")
	t.Setenv("APP_REDIS__URL", "redis://localhost:6379/0")
	t.Setenv("APP_JWT__SECRET", "test-secret-must-be-at-least-32-characters")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Server.Env != "development" {
		t.Errorf("expected default env development, got %s", cfg.Server.Env)
	}
	if cfg.Server.LogLevel != "info" {
		t.Errorf("expected default log level info, got %s", cfg.Server.LogLevel)
	}
	if cfg.JWT.ExpireHours != 24 {
		t.Errorf("expected default jwt expire hours 24, got %d", cfg.JWT.ExpireHours)
	}
	if cfg.Relay.MaxRetries != 2 {
		t.Errorf("expected default relay max retries 2, got %d", cfg.Relay.MaxRetries)
	}
	if cfg.Billing.DefaultMaxTokens != 8192 {
		t.Errorf("expected default billing default max tokens 8192, got %d", cfg.Billing.DefaultMaxTokens)
	}
}

func TestLoad_FromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 9090
  env: test
  log_level: debug
database:
  url: "postgres://yaml/db"
redis:
  url: "redis://yaml:6379/1"
jwt:
  secret: "yaml-secret-key-must-be-32-char-long"
  expire_hours: 12
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.Env != "test" {
		t.Errorf("expected env test, got %s", cfg.Server.Env)
	}
	if cfg.Server.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.Server.LogLevel)
	}
	if cfg.Database.URL != "postgres://yaml/db" {
		t.Errorf("expected database url from yaml, got %s", cfg.Database.URL)
	}
	if cfg.Redis.URL != "redis://yaml:6379/1" {
		t.Errorf("expected redis url from yaml, got %s", cfg.Redis.URL)
	}
	if cfg.JWT.Secret != "yaml-secret-key-must-be-32-char-long" {
		t.Errorf("expected jwt secret from yaml, got %s", cfg.JWT.Secret)
	}
	if cfg.JWT.ExpireHours != 12 {
		t.Errorf("expected jwt expire hours 12, got %d", cfg.JWT.ExpireHours)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := `
server:
  port: 9090
database:
  url: "postgres://yaml/db"
redis:
  url: "redis://yaml:6379/1"
jwt:
  secret: "yaml-secret-key-must-be-32-char-long"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	t.Setenv("APP_SERVER__PORT", "7777")
	t.Setenv("APP_DATABASE__URL", "postgres://env/db")
	t.Setenv("APP_REDIS__URL", "redis://env:6379/2")
	t.Setenv("APP_JWT__SECRET", "env-secret-key-must-be-32-char-long")
	t.Setenv("APP_JWT__EXPIRE_HOURS", "48")
	t.Setenv("APP_SERVER__LOG_LEVEL", "warn")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Server.Port != 7777 {
		t.Errorf("expected env override port 7777, got %d", cfg.Server.Port)
	}
	if cfg.Database.URL != "postgres://env/db" {
		t.Errorf("expected env override database url, got %s", cfg.Database.URL)
	}
	if cfg.Redis.URL != "redis://env:6379/2" {
		t.Errorf("expected env override redis url, got %s", cfg.Redis.URL)
	}
	if cfg.JWT.Secret != "env-secret-key-must-be-32-char-long" {
		t.Errorf("expected env override jwt secret, got %s", cfg.JWT.Secret)
	}
	if cfg.JWT.ExpireHours != 48 {
		t.Errorf("expected env override jwt expire hours 48, got %d", cfg.JWT.ExpireHours)
	}
	if cfg.Server.LogLevel != "warn" {
		t.Errorf("expected env override log level warn, got %s", cfg.Server.LogLevel)
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("APP_REDIS__URL", "redis://localhost:6379/0")
	t.Setenv("APP_JWT__SECRET", "test-secret-must-be-at-least-32-characters")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for missing database url")
	}
}

func TestLoad_MissingRedisURL(t *testing.T) {
	t.Setenv("APP_DATABASE__URL", "postgres://localhost/db")
	t.Setenv("APP_JWT__SECRET", "test-secret-must-be-at-least-32-characters")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for missing redis url")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	t.Setenv("APP_DATABASE__URL", "postgres://localhost/db")
	t.Setenv("APP_REDIS__URL", "redis://localhost:6379/0")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for missing jwt secret")
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("APP_SERVER__PORT", "70000")
	t.Setenv("APP_DATABASE__URL", "postgres://localhost/db")
	t.Setenv("APP_REDIS__URL", "redis://localhost:6379/0")
	t.Setenv("APP_JWT__SECRET", "test-secret-must-be-at-least-32-characters")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}

func TestLoad_JWTSecretTooShort(t *testing.T) {
	t.Setenv("APP_DATABASE__URL", "postgres://localhost/db")
	t.Setenv("APP_REDIS__URL", "redis://localhost:6379/0")
	t.Setenv("APP_JWT__SECRET", "short")

	_, err := Load("")
	if err == nil {
		t.Fatal("expected error for short jwt secret")
	}
}

func TestLoad_MissingConfigFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}

func TestLoad_NumericJWTSecretPreserved(t *testing.T) {
	t.Setenv("APP_DATABASE__URL", "postgres://localhost/db")
	t.Setenv("APP_REDIS__URL", "redis://localhost:6379/0")
	t.Setenv("APP_JWT__SECRET", "12345678901234567890123456789012")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.JWT.Secret != "12345678901234567890123456789012" {
		t.Errorf("expected numeric jwt secret to remain string, got %q", cfg.JWT.Secret)
	}
}
