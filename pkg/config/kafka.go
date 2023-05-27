// Package config provides functionality for loading configuration settings.

package config

import (
	"fmt"
	"log"

	"github.com/caarlos0/env/v8"
)

// Config represents the application configuration.
type Config struct {
	Brokers        string `env:"BROKERS" envDefault:"localhost:9092"`
	Topics         string `env:"TOPICS" envDefault:"expense-topic"`
	SchemaRegistry string `env:"SCHEMAREGISTRY" envDefault:"localhost:8081"`
}

// KafkaTLS represents the configuration for Kafka TLS settings.
type KafkaTLS struct {
	Enabled               bool   `env:"ENABLED"`
	CaFilepath            string `env:"CAFILEPATH"`
	CertFilepath          string `env:"CERTFILEPATH"`
	KeyFilepath           string `env:"KEYFILEPATH"`
	InsecureSkipTLSVerify bool   `env:"INSECURESKIPTLSVERIFY"`
}

// NewConfig creates a new Config instance by parsing environment variables.
// It returns a pointer to the Config and an error if there was a problem parsing the environment variables.
func NewConfig() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("error processing environment variables: %w", err)
	}

	log.Printf("Loaded configuration: %+v\n", cfg)

	return &cfg, nil
}
