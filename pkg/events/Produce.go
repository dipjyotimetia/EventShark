// Package events provides functionality for producing messages to Kafka topics.

package events

import (
	"context"
	"fmt"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/logger"
	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
)

// Producer defines the interface for producing messages to Kafka.
type Producer interface {
	ProduceSync(ctx context.Context, record *kgo.Record) error
	ProduceAsync(ctx context.Context, record *kgo.Record) error
	SetRecord(cfg *config.Config, data interface{}, topic string, schemaType interface{}) (*kgo.Record, error)
	Close() error
}

// KafkaClient wraps a kgo.Client to provide Kafka producer functionality.
type KafkaClient struct {
	client *kgo.Client
	logger *logger.Logger
}

// NewKafkaClient creates a new KafkaClient based on the provided configuration.
// It initializes a Kafka producer client and returns a KafkaClient instance.
func NewKafkaClient(cfg *config.Config) (*KafkaClient, error) {
	seeds := []string{cfg.Brokers}

	opts := []kgo.Opt{
		kgo.SeedBrokers(seeds...),
		kgo.ClientID("event-shark-producer"),
		kgo.ProducerBatchCompression(kgo.GzipCompression()),
		kgo.ProducerBatchMaxBytes(1000000),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerLinger(5 * time.Millisecond),
	}

	// Add TLS configuration if enabled
	if cfg.TLS.Enabled {
		// TLS configuration would be added here
		// This is a placeholder for TLS implementation
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("error initializing Kafka producer client: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Kafka brokers: %w", err)
	}

	return &KafkaClient{
		client: client,
		logger: logger.New(),
	}, nil
}

// ProduceSync sends a Kafka record synchronously and returns an error if any occurred.
func (c *KafkaClient) ProduceSync(ctx context.Context, record *kgo.Record) error {
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	results := c.client.ProduceSync(ctx, record)

	for _, pr := range results {
		if pr.Err != nil {
			c.logger.LogError(ctx, pr.Err, "failed to produce message synchronously",
				"topic", pr.Record.Topic,
				"partition", pr.Record.Partition,
			)
			return fmt.Errorf("error sending synchronous message to topic %s: %w", pr.Record.Topic, pr.Err)
		}

		c.logger.LogKafkaEvent(ctx, pr.Record.Topic, pr.Record.Partition, pr.Record.Offset, "message produced successfully")
	}

	return nil
}

// ProduceAsync sends a Kafka record asynchronously.
func (c *KafkaClient) ProduceAsync(ctx context.Context, record *kgo.Record) error {
	if record == nil {
		return fmt.Errorf("record cannot be nil")
	}

	c.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			c.logger.LogError(ctx, err, "failed to produce message asynchronously",
				"topic", r.Topic,
				"partition", r.Partition,
			)
		} else {
			c.logger.LogKafkaEvent(ctx, r.Topic, r.Partition, r.Offset, "message produced successfully (async)")
		}
	})

	return nil
}

// Producer maintains backward compatibility - delegates to ProduceSync.
func (c *KafkaClient) Producer(ctx context.Context, record *kgo.Record) error {
	return c.ProduceSync(ctx, record)
}

// Close closes the Kafka client connection.
func (c *KafkaClient) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	return nil
}

// getSchema retrieves the Avro schema for the specified subject from the schema registry.
func getSchema(cfg config.Config, subject string) (sr.SubjectSchema, error) {
	if cfg.SchemaRegistry == "" {
		return sr.SubjectSchema{}, fmt.Errorf("schema registry URL is not configured")
	}

	rcl, err := sr.NewClient(sr.URLs("http://" + cfg.SchemaRegistry))
	if err != nil {
		return sr.SubjectSchema{}, fmt.Errorf("unable to create schema registry client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	schemaSubject, err := rcl.SchemaByVersion(ctx, subject, -1)
	if err != nil {
		return sr.SubjectSchema{}, fmt.Errorf("unable to get schema from registry for subject %s: %w", subject, err)
	}

	return schemaSubject, nil
}

// SetRecord encodes the provided data using Avro and creates a Kafka record with the encoded value.
func (c *KafkaClient) SetRecord(cfg *config.Config, data interface{}, topic string, schemaType interface{}) (*kgo.Record, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}
	if topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	schemaSubject, err := getSchema(*cfg, topic+"-value")
	if err != nil {
		return nil, fmt.Errorf("failed to get schema for topic %s: %w", topic, err)
	}

	avroSchema, err := avro.Parse(schemaSubject.Schema.Schema)
	if err != nil {
		return nil, fmt.Errorf("unable to parse avro schema for topic %s: %w", topic, err)
	}

	var serde sr.Serde
	serde.Register(
		schemaSubject.ID,
		schemaType,
		sr.EncodeFn(func(v interface{}) ([]byte, error) {
			return avro.Marshal(avroSchema, v)
		}),
		sr.DecodeFn(func(b []byte, v interface{}) error {
			return avro.Unmarshal(avroSchema, b, v)
		}),
	)

	encodedData, err := serde.Encode(data)
	if err != nil {
		return nil, fmt.Errorf("failed to encode data with Avro schema: %w", err)
	}

	record := &kgo.Record{
		Value:     encodedData,
		Topic:     topic,
		Timestamp: time.Now(),
	}

	return record, nil
}

// Ptr returns a pointer to the provided value.
func Ptr[T any](val T) *T {
	return &val
}
