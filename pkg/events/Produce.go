// Package events provides functionality for producing messages to Kafka topics.

package events

import (
	"context"
	"fmt"

	"github.com/dipjyotimetia/event-shark/pkg/config"
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
// getSchema retrieves the Avro schema for the specified subject from the schema registry.
func getSchema(cfg config.Config, subject string) (sr.SubjectSchema, error) {
	rcl, err := sr.NewClient(sr.URLs(cfg.SchemaRegistry))
	if err != nil {
		return sr.SubjectSchema{}, fmt.Errorf("unable to create schema registry client: %w", err)
	}
	schemaSubject, err := rcl.SchemaByVersion(context.Background(), subject, -1)
	if err != nil {
		return sr.SubjectSchema{}, fmt.Errorf("unable to get schema registry client: %w", err)
	}
	return schemaSubject, nil
}

// SetRecord encodes the provided data using Avro and creates a Kafka record with the encoded value.
func (c KafkaClient) SetRecord(cfg *config.Config, ts interface{}, topic string, schemaType interface{}) (*kgo.Record, error) {
	schemaSubject, err := getSchema(*cfg, topic+"-value")
	if err != nil {
		return nil, err
	}
	avroSchema, err := avro.Parse(schemaSubject.Schema.Schema)
	if err != nil {
		return nil, fmt.Errorf("unable to parse avro schema: %w", err)
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
	tt := serde.MustEncode(ts)
	record := kgo.Record{
		Value: tt,
		Topic: topic,
	}
	return &record, nil
}

// Ptr returns a pointer to the provided value.
func Ptr[T any](val T) *T {
	return &val
}
