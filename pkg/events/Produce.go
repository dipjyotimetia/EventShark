// Package events provides functionality for producing messages to Kafka topics.

package events

import (
	"context"
	"fmt"

	"github.com/dipjyotimetia/event-stream/gen"
	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
)

// KafkaClient wraps a kgo.Client to provide Kafka producer functionality.
type KafkaClient struct {
	*kgo.Client
}

// Produce defines the interface for producing messages to Kafka.
type Produce interface {
	Producer(ctx context.Context, record *kgo.Record)
}

// NewKafkaClient creates a new KafkaClient based on the provided configuration.
// It initializes a Kafka producer client and returns a KafkaClient instance.
func NewKafkaClient(cfg *config.Config) *KafkaClient {
	seeds := []string{cfg.Brokers}
	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.DefaultProduceTopic("expense-topic"),
	)
	if err != nil {
		fmt.Printf("error initializing Kafka producer client: %v\n", err)
		return &KafkaClient{}
	}
	return &KafkaClient{client}
}

// Producer sends a Kafka record synchronously and prints the result.
func (c KafkaClient) Producer(ctx context.Context, record *kgo.Record) error {
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

// getSchema retrieves the Avro schema for the specified subject from the schema registry.
func getSchema(cfg config.Config, subject string) sr.SubjectSchema {
	rcl, err := sr.NewClient(sr.URLs(cfg.SchemaRegistry))
	if err != nil {
		_ = fmt.Errorf("unable to create schema registry client")
	}
	schemaSubject, err := rcl.SchemaByVersion(context.Background(), subject, -1)
	if err != nil {
		_ = fmt.Errorf("unable to get schema registry client")
	}
	return schemaSubject
}

// SetExpenseRecord encodes the provided data using Avro and creates a Kafka record with the encoded value.
func (c KafkaClient) SetExpenseRecord(cfg *config.Config, ts interface{}) *kgo.Record {
	schemaSubject := getSchema(*cfg, "expense-topic-value")
	avroSchema, err := avro.Parse(schemaSubject.Schema.Schema)
	if err != nil {
		_ = fmt.Errorf("unable to parse avro schema")
	}

	var serde sr.Serde
	serde.Register(
		schemaSubject.ID,
		gen.Expense{},
		sr.EncodeFn(func(v interface{}) ([]byte, error) {
			return avro.Marshal(avroSchema, v)
		}),
		sr.DecodeFn(func(b []byte, v interface{}) error {
			return avro.Unmarshal(avroSchema, b, v)
		}),
	)
	tt := serde.MustEncode(ts)
	record := kgo.Record{
		Value: tt,
	}
	return &record
}

// Ptr returns a pointer to the provided value.
func Ptr[T any](val T) *T {
	return &val
}
