// Package events provides functionality for producing messages to Kafka topics.

package events

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
)

// SchemaCache represents a cached schema with its serde
type SchemaCache struct {
	Schema     sr.SubjectSchema
	AvroSchema avro.Schema
	Serde      *sr.Serde
}

// KafkaClient wraps a kgo.Client to provide Kafka producer functionality.
type KafkaClient struct {
	*kgo.Client
	schemaCache map[string]*SchemaCache
	mutex       sync.RWMutex
	srClient    *sr.Client
}

// Produce defines the interface for producing messages to Kafka.
type Produce interface {
	Producer(ctx context.Context, record *kgo.Record)
}

// NewKafkaClient creates a new KafkaClient based on the provided configuration.
// It initializes a Kafka producer client and returns a KafkaClient instance.
func NewKafkaClient(cfg *config.Config) *KafkaClient {
	seeds := []string{cfg.Brokers}
	
	// Create Schema Registry client
	srClient, err := sr.NewClient(sr.URLs(cfg.SchemaRegistry))
	if err != nil {
		fmt.Printf("error initializing Schema Registry client: %v\n", err)
		return &KafkaClient{}
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.ProducerBatchCompression(kgo.GzipCompression()),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.RecordRetries(3),
		kgo.ProducerBatchMaxBytes(1000000), // 1MB
		kgo.ProducerLinger(5*time.Millisecond),
	)
	if err != nil {
		fmt.Printf("error initializing Kafka producer client: %v\n", err)
		return &KafkaClient{}
	}

	return &KafkaClient{
		Client:      client,
		schemaCache: make(map[string]*SchemaCache),
		srClient:    srClient,
	}
}

// Producer sends a Kafka record synchronously and prints the result.
func (c *KafkaClient) Producer(ctx context.Context, record *kgo.Record) error {
	results := c.Client.ProduceSync(ctx, record)
	for _, pr := range results {
		if pr.Err != nil {
			return fmt.Errorf("error sending synchronous message: %v", pr.Err)
		} else {
			fmt.Printf("Message sent: topic: %s, offset: %d, partition: %d \n",
				pr.Record.Topic, pr.Record.Offset, pr.Record.Partition)
		}
	}
	return nil
}

// Close closes the Kafka client and cleans up resources
func (c *KafkaClient) Close() {
	if c.Client != nil {
		c.Client.Close()
	}
}

// getOrCreateSchemaCache retrieves or creates a cached schema for the specified subject
func (c *KafkaClient) getOrCreateSchemaCache(subject string) (*SchemaCache, error) {
	c.mutex.RLock()
	cached, exists := c.schemaCache[subject]
	c.mutex.RUnlock()

	if exists {
		return cached, nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Double-check pattern
	if cached, exists := c.schemaCache[subject]; exists {
		return cached, nil
	}

	schemaSubject, err := c.srClient.SchemaByVersion(context.Background(), subject, -1)
	if err != nil {
		return nil, fmt.Errorf("unable to get schema: %w", err)
	}

	avroSchema, err := avro.Parse(schemaSubject.Schema.Schema)
	if err != nil {
		return nil, fmt.Errorf("unable to parse avro schema: %w", err)
	}

	serde := &sr.Serde{}

	cached = &SchemaCache{
		Schema:     schemaSubject,
		AvroSchema: avroSchema,
		Serde:      serde,
	}

	c.schemaCache[subject] = cached
	return cached, nil
}

// generateRecordKey generates a consistent key for the record based on content
func generateRecordKey(data any) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%v", data)))
	return fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes as key
}

// SetRecord encodes the provided data using Avro and creates a Kafka record with the encoded value.
func (c *KafkaClient) SetRecord(cfg *config.Config, ts any, topic string, schemaType any) (*kgo.Record, error) {
	subject := topic + "-value"
	
	cached, err := c.getOrCreateSchemaCache(subject)
	if err != nil {
		return nil, err
	}

	cached.Serde.Register(
		cached.Schema.ID,
		schemaType,
		sr.EncodeFn(func(v any) ([]byte, error) {
			return avro.Marshal(cached.AvroSchema, v)
		}),
		sr.DecodeFn(func(b []byte, v any) error {
			return avro.Unmarshal(cached.AvroSchema, b, v)
		}),
	)

	encodedValue := cached.Serde.MustEncode(ts)
	recordKey := generateRecordKey(ts)

	record := kgo.Record{
		Key:   []byte(recordKey),
		Value: encodedValue,
		Topic: topic,
	}
	return &record, nil
}

// Ptr returns a pointer to the provided value.
func Ptr[T any](val T) *T {
	return &val
}
