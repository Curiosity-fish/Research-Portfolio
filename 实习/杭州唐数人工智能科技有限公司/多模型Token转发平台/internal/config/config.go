package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// AppConfig holds the complete application configuration.
type AppConfig struct {
	Server     ServerConfig     `koanf:"server"`
	Database   DatabaseConfig   `koanf:"database"`
	Redis      RedisConfig      `koanf:"redis"`
	JWT        JWTConfig        `koanf:"jwt"`
	Encryption EncryptionConfig `koanf:"encryption"`
	Relay      RelayConfig      `koanf:"relay"`
	Billing    BillingConfig    `koanf:"billing"`
	Retention  RetentionConfig  `koanf:"retention"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port     int    `koanf:"port"`
	Env      string `koanf:"env"`
	LogLevel string `koanf:"log_level"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	URL string `koanf:"url"`
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	URL string `koanf:"url"`
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret      string `koanf:"secret"`
	ExpireHours int    `koanf:"expire_hours"`
}

// EncryptionConfig holds the symmetric key used to encrypt secrets at rest
// (e.g. upstream API keys stored in the accounts table).
// Key must be a base64-encoded 32-byte value (AES-256).
type EncryptionConfig struct {
	Key string `koanf:"key"`
}

// RelayConfig holds settings for AI upstream forwarding.
type RelayConfig struct {
	Timeout      time.Duration `koanf:"timeout"`
	MaxBodyBytes int64         `koanf:"max_body_bytes"`
	MaxRetries   int           `koanf:"max_retries"`
}

// BillingConfig holds settings for quota and balance calculation.
type BillingConfig struct {
	DefaultMaxTokens int64 `koanf:"default_max_tokens"`
}

// RetentionConfig holds data-retention settings. CallLogDays is the number of
// days call logs are kept (customer requirement #7: 30 days); a value <= 0
// disables sweeping entirely. SweepInterval is how often the background sweep
// runs.
type RetentionConfig struct {
	CallLogDays   int           `koanf:"call_log_days"`
	SweepInterval time.Duration `koanf:"sweep_interval"`
}

// Load reads configuration from the provided YAML file and environment variables.
// Environment variables take precedence and use the prefix "APP_". Double underscores
// separate nested levels, single underscores are part of the key name.
// Example: APP_SERVER__PORT, APP_JWT__EXPIRE_HOURS.
func Load(path string) (*AppConfig, error) {
	k := koanf.NewWithConf(koanf.Conf{
		Delim: ".",
	})

	// Load safe defaults first.
	defaults := map[string]interface{}{
		"server.port":          8080,
		"server.env":           "development",
		"server.log_level":     "info",
		"database.url":         "",
		"redis.url":            "",
		"jwt.secret":           "",
		"jwt.expire_hours":     24,
		"encryption.key":       "",
		"relay.timeout":        120 * time.Second,
		"relay.max_body_bytes": 50 * 1024 * 1024,
		"relay.max_retries":    2,
		"billing.default_max_tokens": 8192,
		"retention.call_log_days":   30,
		"retention.sweep_interval":  24 * time.Hour,
	}
	if err := k.Load(confmap.Provider(defaults, "."), nil); err != nil {
		return nil, fmt.Errorf("load defaults: %w", err)
	}

	// Load YAML file if provided. If path is explicitly provided but missing, fail.
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("load config file %s: %w", path, err)
			}
		} else if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s", path)
		} else {
			return nil, fmt.Errorf("check config file %s: %w", path, err)
		}
	}

	// Load environment variables with APP_ prefix.
	if err := k.Load(env.ProviderWithValue("APP_", ".", envVarMapper), nil); err != nil {
		return nil, fmt.Errorf("load env vars: %w", err)
	}

	var cfg AppConfig
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// envVarMapper converts environment variable keys to koanf keys.
// Double underscores are used as level separators, single underscores are preserved.
// Example: APP_SERVER__PORT -> server.port; APP_JWT__EXPIRE_HOURS -> jwt.expire_hours.
func envVarMapper(key string, value string) (string, interface{}) {
	key = strings.TrimPrefix(key, "APP_")
	key = strings.ToLower(key)
	key = strings.ReplaceAll(key, "__", ".")
	return key, value
}

// Validate checks that the configuration is usable.
func (c *AppConfig) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535, got %d", c.Server.Port)
	}

	if c.Server.Env == "" {
		return errors.New("server.env is required")
	}

	if c.Server.LogLevel == "" {
		return errors.New("server.log_level is required")
	}

	if c.Database.URL == "" {
		return errors.New("database.url is required")
	}

	if c.Redis.URL == "" {
		return errors.New("redis.url is required")
	}

	if c.JWT.Secret == "" {
		return errors.New("jwt.secret is required")
	}
	if len(c.JWT.Secret) < 32 {
		return errors.New("jwt.secret must be at least 32 characters long for HS256")
	}

	if c.JWT.ExpireHours < 1 {
		return fmt.Errorf("jwt.expire_hours must be at least 1, got %d", c.JWT.ExpireHours)
	}

	if c.Relay.Timeout <= 0 {
		return errors.New("relay.timeout must be positive")
	}
	if c.Relay.MaxBodyBytes <= 0 {
		return errors.New("relay.max_body_bytes must be positive")
	}
	if c.Relay.MaxRetries < 0 {
		return errors.New("relay.max_retries must be non-negative")
	}
	if c.Billing.DefaultMaxTokens <= 0 {
		return errors.New("billing.default_max_tokens must be positive")
	}
	// retention.call_log_days may be 0/negative to disable sweeping; only a
	// positive sweep interval paired with sweeping enabled is usable.
	if c.Retention.CallLogDays > 0 && c.Retention.SweepInterval <= 0 {
		return errors.New("retention.sweep_interval must be positive when call-log retention is enabled")
	}

	return nil
}
