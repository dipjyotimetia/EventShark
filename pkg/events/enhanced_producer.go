// Package events provides enhanced Kafka producer functionality with TLS and compression
package events

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

// EnhancedProducer provides advanced Kafka producer capabilities
type EnhancedProducer struct {
	client *kgo.Client
	config *config.Config
}

// NewEnhancedProducer creates a new enhanced Kafka producer with TLS and compression support
func NewEnhancedProducer(cfg *config.Config) (*EnhancedProducer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Kafka.Brokers),
	}

	// Configure TLS if enabled
	if cfg.Kafka.TLS.Enabled {
		tlsConfig, err := createTLSConfig(&cfg.Kafka.TLS)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts = append(opts, kgo.DialTLSConfig(tlsConfig))
	}

	// Configure compression
	if cfg.Compression.Enabled {
		codec, err := getCompressionCodec(cfg.Compression.Codec)
		if err != nil {
			return nil, fmt.Errorf("failed to configure compression: %w", err)
		}
		opts = append(opts, kgo.ProducerBatchCompression(codec...))
	}

	// Create client
	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	return &EnhancedProducer{
		client: client,
		config: cfg,
	}, nil
}

// createTLSConfig creates TLS configuration from config
func createTLSConfig(tlsCfg *config.KafkaTLS) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: tlsCfg.InsecureSkipTLSVerify,
	}

	// Load CA certificate if specified
	if tlsCfg.CaFilepath != "" {
		caCert, err := os.ReadFile(tlsCfg.CaFilepath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if specified
	if tlsCfg.CertFilepath != "" && tlsCfg.KeyFilepath != "" {
		cert, err := tls.LoadX509KeyPair(tlsCfg.CertFilepath, tlsCfg.KeyFilepath)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}

// getCompressionCodec returns the appropriate compression codec
func getCompressionCodec(codec string) ([]kgo.CompressionCodec, error) {
	switch codec {
	case "none":
		return []kgo.CompressionCodec{kgo.NoCompression()}, nil
	case "gzip":
		return []kgo.CompressionCodec{kgo.GzipCompression()}, nil
	case "snappy":
		return []kgo.CompressionCodec{kgo.SnappyCompression()}, nil
	case "lz4":
		return []kgo.CompressionCodec{kgo.Lz4Compression()}, nil
	case "zstd":
		return []kgo.CompressionCodec{kgo.ZstdCompression()}, nil
	default:
		return nil, fmt.Errorf("unsupported compression codec: %s", codec)
	}
}

// Produce sends a message to Kafka
func (ep *EnhancedProducer) Produce(ctx context.Context, record *kgo.Record) error {
	results := ep.client.ProduceSync(ctx, record)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("failed to produce message: %w", err)
	}
	return nil
}

// ProduceAsync sends a message to Kafka asynchronously
func (ep *EnhancedProducer) ProduceAsync(record *kgo.Record, callback func(record *kgo.Record, err error)) {
	ep.client.Produce(context.Background(), record, callback)
}

// GetClient returns the underlying Kafka client
func (ep *EnhancedProducer) GetClient() *kgo.Client {
	return ep.client
}

// Close closes the producer
func (ep *EnhancedProducer) Close() {
	ep.client.Close()
}

// Flush flushes any pending messages
func (ep *EnhancedProducer) Flush(ctx context.Context) error {
	return ep.client.Flush(ctx)
}
