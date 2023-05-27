package events

import (
	"context"
	"fmt"

	"github.com/dipjyotimetia/event-stream/gen/expense"
	"github.com/dipjyotimetia/event-stream/pkg/config"
	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
)

type KafkaClient struct {
	*kgo.Client
}

type Produce interface {
	Producer(ctx context.Context, record *kgo.Record)
}

func NewKafkaClient(cfg *config.Config) KafkaClient {
	seeds := []string{cfg.Brokers}
	client, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.DefaultProduceTopic("expense-topic"),
	)
	if err != nil {
		fmt.Printf("error initializing Kafka producer client: %v\n", err)
		return KafkaClient{}
	}
	return KafkaClient{client}
}

func (c KafkaClient) Producer(ctx context.Context, record *kgo.Record) {
	results := c.Client.ProduceSync(ctx, record)
	for _, pr := range results {
		if pr.Err != nil {
			fmt.Printf("Error sending synchronous message: %v \n", pr.Err)
		} else {
			fmt.Printf("Message sent: topic: %s, offset: %d, partition: %d \n",
				pr.Record.Topic, pr.Record.Offset, pr.Record.Partition)
		}
	}
}

func getSchema(cfg config.Config, subject string) sr.SubjectSchema {
	rcl, err := sr.NewClient(sr.URLs(cfg.SchemaRegistry))
	if err != nil {
		_ = fmt.Errorf("unable to create schema registry client")
	}
	schemaSubject, err := rcl.SchemaByVersion(context.Background(), subject, -1, false)
	if err != nil {
		_ = fmt.Errorf("unable to get schema registry client")
	}
	return schemaSubject
}

func (c KafkaClient) SetExpenseRecord(cfg *config.Config, ts any) *kgo.Record {
	schemaSubject := getSchema(*cfg, "expense-topic-value")
	avroSchema, err := avro.Parse(schemaSubject.Schema.Schema)
	if err != nil {
		_ = fmt.Errorf("unable to parse avro schema")
	}

	var serde sr.Serde
	serde.Register(
		schemaSubject.ID,
		expense.Expense{},
		sr.EncodeFn(func(v any) ([]byte, error) {
			return avro.Marshal(avroSchema, v)
		}),
		sr.DecodeFn(func(b []byte, v any) error {
			return avro.Unmarshal(avroSchema, b, v)
		}),
	)
	tt := serde.MustEncode(ts)
	record := kgo.Record{
		Value: tt,
	}
	return &record
}

func Ptr[T any](val T) *T {
	return &val
}
