package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	// Store original env vars
	originalEnvs := make(map[string]string)
	envVars := []string{"SERVER_PORT", "BROKERS", "SCHEMAREGISTRY", "ENVIRONMENT", "LOG_LEVEL"}

	for _, env := range envVars {
		originalEnvs[env] = os.Getenv(env)
	}

	// Clean up after test
	defer func() {
		for env, val := range originalEnvs {
			if val == "" {
				os.Unsetenv(env)
			} else {
				os.Setenv(env, val)
			}
		}
	}()

	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "valid config with defaults",
			envVars: map[string]string{
				"ENVIRONMENT": "development",
			},
			expectError: false,
		},
		{
			name: "valid config with custom values",
			envVars: map[string]string{
				"SERVER_PORT":    "8080",
				"BROKERS":        "localhost:9092,localhost:9093",
				"SCHEMAREGISTRY": "localhost:8081",
				"ENVIRONMENT":    "production",
				"LOG_LEVEL":      "info",
			},
			expectError: false,
		},
		{
			name: "invalid environment",
			envVars: map[string]string{
				"ENVIRONMENT": "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			envVars: map[string]string{
				"LOG_LEVEL": "invalid",
			},
			expectError: true,
		},
		{
			name: "empty brokers",
			envVars: map[string]string{
				"BROKERS": " ",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			cfg, err := NewConfig()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
				if cfg == nil {
					t.Errorf("expected config but got nil")
				}
			}

			// Clean up environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				ServerPort:      "8083",
				ServerTimeout:   30 * time.Second,
				ShutdownTimeout: 30 * time.Second,
				Brokers:         "localhost:9092",
				Topics:          []string{"test-topic"},
				SchemaRegistry:  "localhost:8081",
				Environment:     "development",
				LogLevel:        "info",
			},
			expectError: false,
		},
		{
			name: "empty server port",
			config: Config{
				ServerPort: "",
			},
			expectError: true,
		},
		{
			name: "empty brokers",
			config: Config{
				ServerPort: "8083",
				Brokers:    "",
			},
			expectError: true,
		},
		{
			name: "invalid broker format",
			config: Config{
				ServerPort: "8083",
				Brokers:    "invalid-broker",
			},
			expectError: true,
		},
		{
			name: "empty topics",
			config: Config{
				ServerPort: "8083",
				Brokers:    "localhost:9092",
				Topics:     []string{},
			},
			expectError: true,
		},
		{
			name: "invalid environment",
			config: Config{
				ServerPort:  "8083",
				Brokers:     "localhost:9092",
				Topics:      []string{"test-topic"},
				Environment: "invalid",
			},
			expectError: true,
		},
		{
			name: "invalid log level",
			config: Config{
				ServerPort:  "8083",
				Brokers:     "localhost:9092",
				Topics:      []string{"test-topic"},
				Environment: "development",
				LogLevel:    "invalid",
			},
			expectError: true,
		},
		{
			name: "TLS enabled but missing cert",
			config: Config{
				ServerPort:  "8083",
				Brokers:     "localhost:9092",
				Topics:      []string{"test-topic"},
				Environment: "development",
				LogLevel:    "info",
				TLS: KafkaTLS{
					Enabled: true,
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"production environment", "production", true},
		{"development environment", "development", false},
		{"staging environment", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Environment: tt.environment}
			result := cfg.IsProduction()

			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		expected    bool
	}{
		{"development environment", "development", true},
		{"production environment", "production", false},
		{"staging environment", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Environment: tt.environment}
			result := cfg.IsDevelopment()

			if result != tt.expected {
				t.Errorf("expected %v but got %v", tt.expected, result)
			}
		})
	}
}
