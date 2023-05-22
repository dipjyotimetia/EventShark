package config

import (
	"errors"
	"fmt"
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Kafka Kafka `envconfig:"KAFKA"`
}

// Kafka API config for connecting to the target cluster.
type Kafka struct {
	Brokers []string `envconfig:"BROKERS"`
	Topics  string   `envconfig:"TOPICS"`
	// TLS            KafkaTLS       `envconfig:"TLS"`
	SchemaRegistry SchemaRegistry `envconfig:"SCHEMAREGISTRY"`
}

type SchemaRegistry struct {
	Enabled bool   `envconfig:"ENABLED"`
	Url     string `envconfig:"URL"`
}

type KafkaTLS struct {
	Enabled               bool   `envconfig:"ENABLED"`
	CaFilepath            string `envconfig:"CAFILEPATH"`
	CertFilepath          string `envconfig:"CERTFILEPATH"`
	KeyFilepath           string `envconfig:"KEYFILEPATH"`
	InsecureSkipTLSVerify bool   `envconfig:"INSECURESKIPTILSVERIFY"`
}

func NewConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("error processing environment variables: %w", err)
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	log.Printf("Loaded configuration: %+v\n", cfg)

	return &cfg, nil
}

func validateConfig(cfg Config) error {
	if len(cfg.Kafka.Brokers) == 0 {
		return errors.New("KAFKA_BROKERS is required")
	}
	if cfg.Kafka.Topics == "" {
		return errors.New("KAFKA_TOPICS is required")
	}
	if cfg.Kafka.SchemaRegistry.Enabled && cfg.Kafka.SchemaRegistry.Url == "" {
		return errors.New("SCHEMAREGISTRY_URL is required when SCHEMAREGISTRY_ENABLED is true")
	}
	return nil
}
