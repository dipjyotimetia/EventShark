// Package config provides functionality for loading configuration settings.

package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Server         ServerConfig         `yaml:"server"`
	Kafka          KafkaConfig          `yaml:"kafka"`
	SchemaRegistry string               `env:"SCHEMAREGISTRY" envDefault:"localhost:8081" yaml:"schema_registry"`
	Async          AsyncConfig          `yaml:"async"`
	Compression    CompressionConfig    `yaml:"compression"`
	Idempotency    IdempotencyConfig    `yaml:"idempotency"`
	CircuitBreaker CircuitBreakerConfig `yaml:"circuit_breaker"`
	DLQ            DLQConfig            `yaml:"dlq"`
	Serialization  SerializationConfig  `yaml:"serialization"`
}

// ServerConfig represents HTTP server configuration
type ServerConfig struct {
	Port         int           `env:"SERVER_PORT" envDefault:"8083" yaml:"port"`
	ReadTimeout  time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"30s" yaml:"read_timeout"`
	WriteTimeout time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"30s" yaml:"write_timeout"`
	IdleTimeout  time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"5s" yaml:"idle_timeout"`
}

// KafkaConfig represents Kafka configuration
type KafkaConfig struct {
	Brokers        string        `env:"BROKERS" envDefault:"localhost:9092" yaml:"brokers"`
	Topics         []string      `env:"TOPICS" envDefault:"expense-topic,payment-topic,transaction-topic" yaml:"topics"`
	TLS            KafkaTLS      `yaml:"tls"`
	RequestTimeout time.Duration `env:"KAFKA_REQUEST_TIMEOUT" envDefault:"30s" yaml:"request_timeout"`
}

// KafkaTLS represents the configuration for Kafka TLS settings.
type KafkaTLS struct {
	Enabled               bool   `env:"TLS_ENABLED" envDefault:"false" yaml:"enabled"`
	CaFilepath            string `env:"TLS_CA_FILEPATH" yaml:"ca_filepath"`
	CertFilepath          string `env:"TLS_CERT_FILEPATH" yaml:"cert_filepath"`
	KeyFilepath           string `env:"TLS_KEY_FILEPATH" yaml:"key_filepath"`
	InsecureSkipTLSVerify bool   `env:"TLS_INSECURE_SKIP_VERIFY" envDefault:"false" yaml:"insecure_skip_verify"`
}

// AsyncConfig represents asynchronous publishing configuration
type AsyncConfig struct {
	Enabled        bool          `env:"ASYNC_ENABLED" envDefault:"true" yaml:"enabled"`
	MaxQueueSize   int           `env:"ASYNC_MAX_QUEUE_SIZE" envDefault:"10000" yaml:"max_queue_size"`
	WorkerCount    int           `env:"ASYNC_WORKER_COUNT" envDefault:"10" yaml:"worker_count"`
	JobTimeout     time.Duration `env:"ASYNC_JOB_TIMEOUT" envDefault:"5m" yaml:"job_timeout"`
	CleanupInterval time.Duration `env:"ASYNC_CLEANUP_INTERVAL" envDefault:"1h" yaml:"cleanup_interval"`
}

// CompressionConfig represents message compression configuration
type CompressionConfig struct {
	Enabled bool   `env:"COMPRESSION_ENABLED" envDefault:"true" yaml:"enabled"`
	Codec   string `env:"COMPRESSION_CODEC" envDefault:"snappy" yaml:"codec"` // none, gzip, snappy, lz4, zstd
}

// IdempotencyConfig represents idempotency configuration
type IdempotencyConfig struct {
	Enabled    bool          `env:"IDEMPOTENCY_ENABLED" envDefault:"true" yaml:"enabled"`
	CacheTTL   time.Duration `env:"IDEMPOTENCY_CACHE_TTL" envDefault:"24h" yaml:"cache_ttl"`
	MaxEntries int           `env:"IDEMPOTENCY_MAX_ENTRIES" envDefault:"100000" yaml:"max_entries"`
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled           bool          `env:"CIRCUIT_BREAKER_ENABLED" envDefault:"true" yaml:"enabled"`
	MaxRequests       uint32        `env:"CIRCUIT_BREAKER_MAX_REQUESTS" envDefault:"3" yaml:"max_requests"`
	Interval          time.Duration `env:"CIRCUIT_BREAKER_INTERVAL" envDefault:"60s" yaml:"interval"`
	Timeout           time.Duration `env:"CIRCUIT_BREAKER_TIMEOUT" envDefault:"30s" yaml:"timeout"`
	FailureThreshold  uint32        `env:"CIRCUIT_BREAKER_FAILURE_THRESHOLD" envDefault:"5" yaml:"failure_threshold"`
}

// DLQConfig represents Dead Letter Queue configuration
type DLQConfig struct {
	Enabled     bool   `env:"DLQ_ENABLED" envDefault:"true" yaml:"enabled"`
	TopicSuffix string `env:"DLQ_TOPIC_SUFFIX" envDefault:"-dlq" yaml:"topic_suffix"`
	MaxRetries  int    `env:"DLQ_MAX_RETRIES" envDefault:"3" yaml:"max_retries"`
}

// SerializationConfig represents serialization format configuration
type SerializationConfig struct {
	DefaultFormat string   `env:"SERIALIZATION_DEFAULT_FORMAT" envDefault:"avro" yaml:"default_format"` // avro, json, protobuf
	SupportedFormats []string `yaml:"supported_formats"`
}

// NewConfig creates a new Config instance by parsing environment variables.
// It returns a pointer to the Config and an error if there was a problem parsing the environment variables.
func NewConfig() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Port:         8083,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  5 * time.Second,
		},
		Kafka: KafkaConfig{
			Brokers:        "localhost:9092",
			Topics:         []string{"expense-topic", "payment-topic", "transaction-topic"},
			RequestTimeout: 30 * time.Second,
		},
		SchemaRegistry: "localhost:8081",
		Async: AsyncConfig{
			Enabled:         true,
			MaxQueueSize:    10000,
			WorkerCount:     10,
			JobTimeout:      5 * time.Minute,
			CleanupInterval: 1 * time.Hour,
		},
		Compression: CompressionConfig{
			Enabled: true,
			Codec:   "snappy",
		},
		Idempotency: IdempotencyConfig{
			Enabled:    true,
			CacheTTL:   24 * time.Hour,
			MaxEntries: 100000,
		},
		CircuitBreaker: CircuitBreakerConfig{
			Enabled:          true,
			MaxRequests:      3,
			Interval:         60 * time.Second,
			Timeout:          30 * time.Second,
			FailureThreshold: 5,
		},
		DLQ: DLQConfig{
			Enabled:     true,
			TopicSuffix: "-dlq",
			MaxRetries:  3,
		},
		Serialization: SerializationConfig{
			DefaultFormat:    "avro",
			SupportedFormats: []string{"avro", "json"},
		},
	}

	// Try to load from YAML file first
	configPath := os.Getenv("CONFIG_FILE")
	if configPath != "" {
		if err := LoadFromYAML(cfg, configPath); err != nil {
			log.Printf("Warning: Failed to load config from YAML: %v", err)
		} else {
			log.Printf("Loaded configuration from YAML: %s", configPath)
		}
	}

	// Override with environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("error processing environment variables: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	log.Printf("Loaded configuration successfully")

	return cfg, nil
}

// LoadFromYAML loads configuration from a YAML file
func LoadFromYAML(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Kafka.Brokers == "" {
		return fmt.Errorf("kafka brokers cannot be empty")
	}

	if len(c.Kafka.Topics) == 0 {
		return fmt.Errorf("at least one topic must be configured")
	}

	if c.SchemaRegistry == "" {
		return fmt.Errorf("schema registry URL cannot be empty")
	}

	// Validate compression codec
	validCodecs := map[string]bool{"none": true, "gzip": true, "snappy": true, "lz4": true, "zstd": true}
	if !validCodecs[c.Compression.Codec] {
		return fmt.Errorf("invalid compression codec: %s", c.Compression.Codec)
	}

	// Validate serialization format
	validFormats := map[string]bool{"avro": true, "json": true, "protobuf": true}
	if !validFormats[c.Serialization.DefaultFormat] {
		return fmt.Errorf("invalid default serialization format: %s", c.Serialization.DefaultFormat)
	}

	// Validate TLS configuration if enabled
	if c.Kafka.TLS.Enabled {
		if c.Kafka.TLS.CertFilepath == "" || c.Kafka.TLS.KeyFilepath == "" {
			return fmt.Errorf("TLS enabled but certificate or key file not specified")
		}
	}

	return nil
}

// GetDLQTopic returns the DLQ topic name for a given topic
func (c *Config) GetDLQTopic(topic string) string {
	return topic + c.DLQ.TopicSuffix
}
