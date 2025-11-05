package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// Check defaults
	if cfg.Server.Port != 8083 {
		t.Errorf("Expected default port 8083, got %d", cfg.Server.Port)
	}

	if cfg.Kafka.Brokers == "" {
		t.Error("Expected default broker address")
	}

	if len(cfg.Kafka.Topics) == 0 {
		t.Error("Expected default topics")
	}
}

func TestConfigValidation(t *testing.T) {
	t.Run("Valid config", func(t *testing.T) {
		cfg := &Config{
			Kafka: KafkaConfig{
				Brokers: "localhost:9092",
				Topics:  []string{"test-topic"},
			},
			SchemaRegistry: "localhost:8081",
			Compression: CompressionConfig{
				Enabled: true,
				Codec:   "snappy",
			},
			Serialization: SerializationConfig{
				DefaultFormat: "avro",
			},
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("Expected valid config, got error: %v", err)
		}
	})

	t.Run("Empty brokers", func(t *testing.T) {
		cfg := &Config{
			Kafka: KafkaConfig{
				Brokers: "",
				Topics:  []string{"test-topic"},
			},
			SchemaRegistry: "localhost:8081",
			Compression: CompressionConfig{
				Codec: "snappy",
			},
			Serialization: SerializationConfig{
				DefaultFormat: "avro",
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected error for empty brokers")
		}
	})

	t.Run("Empty topics", func(t *testing.T) {
		cfg := &Config{
			Kafka: KafkaConfig{
				Brokers: "localhost:9092",
				Topics:  []string{},
			},
			SchemaRegistry: "localhost:8081",
			Compression: CompressionConfig{
				Codec: "snappy",
			},
			Serialization: SerializationConfig{
				DefaultFormat: "avro",
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected error for empty topics")
		}
	})

	t.Run("Empty schema registry", func(t *testing.T) {
		cfg := &Config{
			Kafka: KafkaConfig{
				Brokers: "localhost:9092",
				Topics:  []string{"test-topic"},
			},
			SchemaRegistry: "",
			Compression: CompressionConfig{
				Codec: "snappy",
			},
			Serialization: SerializationConfig{
				DefaultFormat: "avro",
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected error for empty schema registry")
		}
	})

	t.Run("Invalid compression codec", func(t *testing.T) {
		cfg := &Config{
			Kafka: KafkaConfig{
				Brokers: "localhost:9092",
				Topics:  []string{"test-topic"},
			},
			SchemaRegistry: "localhost:8081",
			Compression: CompressionConfig{
				Codec: "invalid",
			},
			Serialization: SerializationConfig{
				DefaultFormat: "avro",
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected error for invalid compression codec")
		}
	})

	t.Run("Invalid serialization format", func(t *testing.T) {
		cfg := &Config{
			Kafka: KafkaConfig{
				Brokers: "localhost:9092",
				Topics:  []string{"test-topic"},
			},
			SchemaRegistry: "localhost:8081",
			Compression: CompressionConfig{
				Codec: "snappy",
			},
			Serialization: SerializationConfig{
				DefaultFormat: "invalid",
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected error for invalid serialization format")
		}
	})

	t.Run("TLS enabled without certificates", func(t *testing.T) {
		cfg := &Config{
			Kafka: KafkaConfig{
				Brokers: "localhost:9092",
				Topics:  []string{"test-topic"},
				TLS: KafkaTLS{
					Enabled:      true,
					CertFilepath: "",
					KeyFilepath:  "",
				},
			},
			SchemaRegistry: "localhost:8081",
			Compression: CompressionConfig{
				Codec: "snappy",
			},
			Serialization: SerializationConfig{
				DefaultFormat: "avro",
			},
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("Expected error for TLS enabled without certificates")
		}
	})
}

func TestCompressionCodecs(t *testing.T) {
	validCodecs := []string{"none", "gzip", "snappy", "lz4", "zstd"}

	for _, codec := range validCodecs {
		t.Run(codec, func(t *testing.T) {
			cfg := &Config{
				Kafka: KafkaConfig{
					Brokers: "localhost:9092",
					Topics:  []string{"test-topic"},
				},
				SchemaRegistry: "localhost:8081",
				Compression: CompressionConfig{
					Enabled: true,
					Codec:   codec,
				},
				Serialization: SerializationConfig{
					DefaultFormat: "avro",
				},
			}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Expected codec %s to be valid, got error: %v", codec, err)
			}
		})
	}
}

func TestSerializationFormats(t *testing.T) {
	validFormats := []string{"avro", "json", "protobuf"}

	for _, format := range validFormats {
		t.Run(format, func(t *testing.T) {
			cfg := &Config{
				Kafka: KafkaConfig{
					Brokers: "localhost:9092",
					Topics:  []string{"test-topic"},
				},
				SchemaRegistry: "localhost:8081",
				Compression: CompressionConfig{
					Codec: "snappy",
				},
				Serialization: SerializationConfig{
					DefaultFormat: format,
				},
			}

			err := cfg.Validate()
			if err != nil {
				t.Errorf("Expected format %s to be valid, got error: %v", format, err)
			}
		})
	}
}

func TestGetDLQTopic(t *testing.T) {
	cfg := &Config{
		DLQ: DLQConfig{
			TopicSuffix: "-dlq",
		},
	}

	dlqTopic := cfg.GetDLQTopic("expense-topic")
	expected := "expense-topic-dlq"

	if dlqTopic != expected {
		t.Errorf("Expected DLQ topic %s, got %s", expected, dlqTopic)
	}
}

func TestEnvironmentVariableOverride(t *testing.T) {
	// Set environment variables
	os.Setenv("BROKERS", "test-broker:9092")
	os.Setenv("SCHEMAREGISTRY", "test-registry:8081")
	os.Setenv("COMPRESSION_CODEC", "gzip")
	defer func() {
		os.Unsetenv("BROKERS")
		os.Unsetenv("SCHEMAREGISTRY")
		os.Unsetenv("COMPRESSION_CODEC")
	}()

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.Kafka.Brokers != "test-broker:9092" {
		t.Errorf("Expected brokers to be overridden, got %s", cfg.Kafka.Brokers)
	}

	if cfg.SchemaRegistry != "test-registry:8081" {
		t.Errorf("Expected schema registry to be overridden, got %s", cfg.SchemaRegistry)
	}

	if cfg.Compression.Codec != "gzip" {
		t.Errorf("Expected compression codec to be overridden, got %s", cfg.Compression.Codec)
	}
}

func TestDefaultValues(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Server defaults
	if cfg.Server.Port != 8083 {
		t.Error("Unexpected default server port")
	}
	if cfg.Server.IdleTimeout != 5*time.Second {
		t.Error("Unexpected default idle timeout")
	}

	// Kafka defaults
	if cfg.Kafka.RequestTimeout != 30*time.Second {
		t.Error("Unexpected default request timeout")
	}

	// Async defaults
	if !cfg.Async.Enabled {
		t.Error("Expected async to be enabled by default")
	}
	if cfg.Async.WorkerCount != 10 {
		t.Error("Unexpected default worker count")
	}

	// Compression defaults
	if !cfg.Compression.Enabled {
		t.Error("Expected compression to be enabled by default")
	}
	if cfg.Compression.Codec != "snappy" {
		t.Error("Unexpected default compression codec")
	}

	// Idempotency defaults
	if !cfg.Idempotency.Enabled {
		t.Error("Expected idempotency to be enabled by default")
	}
	if cfg.Idempotency.CacheTTL != 24*time.Hour {
		t.Error("Unexpected default cache TTL")
	}

	// Circuit breaker defaults
	if !cfg.CircuitBreaker.Enabled {
		t.Error("Expected circuit breaker to be enabled by default")
	}

	// DLQ defaults
	if !cfg.DLQ.Enabled {
		t.Error("Expected DLQ to be enabled by default")
	}
	if cfg.DLQ.MaxRetries != 3 {
		t.Error("Unexpected default max retries")
	}
}
