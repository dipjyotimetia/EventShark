// Package config provides functionality for loading configuration settings.

package config

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config represents the application configuration.
type Config struct {
	// Server configuration
	ServerPort      string        `env:"SERVER_PORT" envDefault:"8083"`
	ServerTimeout   time.Duration `env:"SERVER_TIMEOUT" envDefault:"30s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"30s"`

	// Kafka configuration
	Brokers        string   `env:"BROKERS" envDefault:"localhost:9092"`
	Topics         []string `env:"TOPICS" envDefault:"expense-topic,payment-topic,transaction-topic"`
	SchemaRegistry string   `env:"SCHEMAREGISTRY" envDefault:"localhost:8081"`

	// Environment and logging
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	// TLS configuration
	TLS KafkaTLS
}

// KafkaTLS represents the configuration for Kafka TLS settings.
type KafkaTLS struct {
	Enabled               bool   `env:"TLS_ENABLED" envDefault:"false"`
	CaFilepath            string `env:"TLS_CA_FILEPATH"`
	CertFilepath          string `env:"TLS_CERT_FILEPATH"`
	KeyFilepath           string `env:"TLS_KEY_FILEPATH"`
	InsecureSkipTLSVerify bool   `env:"TLS_INSECURE_SKIP_VERIFY" envDefault:"false"`
}

// NewConfig creates a new Config instance by parsing environment variables.
// It returns a pointer to the Config and an error if there was a problem parsing the environment variables.
func NewConfig() (*Config, error) {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error processing environment variables: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	log.Printf("Loaded configuration for environment: %s\n", cfg.Environment)

	return &cfg, nil
}

// Validate validates the configuration values.
func (c *Config) Validate() error {
	// Validate server port
	if c.ServerPort == "" {
		return fmt.Errorf("server port cannot be empty")
	}

	// Validate brokers
	if c.Brokers == "" || strings.TrimSpace(c.Brokers) == "" {
		return fmt.Errorf("Kafka brokers cannot be empty")
	}

	// Validate broker URLs
	brokers := strings.Split(c.Brokers, ",")
	for _, broker := range brokers {
		broker = strings.TrimSpace(broker)
		if broker == "" {
			continue
		}

		// Simple validation for host:port format
		if !strings.Contains(broker, ":") {
			return fmt.Errorf("invalid broker format: %s (expected host:port)", broker)
		}
	}

	// Validate schema registry URL
	if c.SchemaRegistry != "" {
		if _, err := url.Parse("http://" + c.SchemaRegistry); err != nil {
			return fmt.Errorf("invalid schema registry URL: %w", err)
		}
	}

	// Validate topics
	if len(c.Topics) == 0 {
		return fmt.Errorf("at least one topic must be configured")
	}

	for _, topic := range c.Topics {
		if topic == "" {
			return fmt.Errorf("topic name cannot be empty")
		}
	}

	// Validate environment
	validEnvs := []string{"development", "staging", "production"}
	isValidEnv := false
	for _, env := range validEnvs {
		if c.Environment == env {
			isValidEnv = true
			break
		}
	}
	if !isValidEnv {
		return fmt.Errorf("invalid environment: %s (must be one of: %s)", c.Environment, strings.Join(validEnvs, ", "))
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error"}
	isValidLogLevel := false
	for _, level := range validLogLevels {
		if c.LogLevel == level {
			isValidLogLevel = true
			break
		}
	}
	if !isValidLogLevel {
		return fmt.Errorf("invalid log level: %s (must be one of: %s)", c.LogLevel, strings.Join(validLogLevels, ", "))
	}

	// Validate TLS configuration
	if c.TLS.Enabled {
		if c.TLS.CertFilepath == "" || c.TLS.KeyFilepath == "" {
			return fmt.Errorf("TLS is enabled but cert or key filepath is missing")
		}
	}

	return nil
}

// IsProduction returns true if the environment is production.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsDevelopment returns true if the environment is development.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
