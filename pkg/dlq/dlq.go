// Package dlq provides Dead Letter Queue functionality for failed messages
package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

// FailedMessage represents a message that failed to process
type FailedMessage struct {
	OriginalTopic string    `json:"original_topic"`
	OriginalKey   []byte    `json:"original_key"`
	OriginalValue []byte    `json:"original_value"`
	Headers       []Header  `json:"headers"`
	FailureReason string    `json:"failure_reason"`
	FailureTime   time.Time `json:"failure_time"`
	RetryCount    int       `json:"retry_count"`
}

// Header represents a message header
type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// DLQManager manages the Dead Letter Queue
type DLQManager struct {
	client     *kgo.Client
	config     *config.DLQConfig
	kafkaConfig *config.KafkaConfig
}

// NewDLQManager creates a new DLQ manager
func NewDLQManager(cfg *config.Config, client *kgo.Client) *DLQManager {
	return &DLQManager{
		client:      client,
		config:      &cfg.DLQ,
		kafkaConfig: &cfg.Kafka,
	}
}

// SendToD LQ sends a failed message to the Dead Letter Queue
func (d *DLQManager) SendToDLQ(ctx context.Context, record *kgo.Record, failureReason string, retryCount int) error {
	if !d.config.Enabled {
		log.Printf("DLQ is disabled, skipping failed message: %s", failureReason)
		return nil
	}

	// Create failed message metadata
	failedMsg := FailedMessage{
		OriginalTopic: record.Topic,
		OriginalKey:   record.Key,
		OriginalValue: record.Value,
		Headers:       convertHeaders(record.Headers),
		FailureReason: failureReason,
		FailureTime:   time.Now(),
		RetryCount:    retryCount,
	}

	// Serialize failed message metadata
	metadata, err := json.Marshal(failedMsg)
	if err != nil {
		return fmt.Errorf("failed to serialize DLQ metadata: %w", err)
	}

	// Determine DLQ topic
	dlqTopic := record.Topic + d.config.TopicSuffix

	// Create DLQ record
	dlqRecord := &kgo.Record{
		Topic: dlqTopic,
		Key:   record.Key,
		Value: record.Value,
		Headers: []kgo.RecordHeader{
			{Key: "x-dlq-metadata", Value: metadata},
			{Key: "x-original-topic", Value: []byte(record.Topic)},
			{Key: "x-failure-reason", Value: []byte(failureReason)},
			{Key: "x-failure-time", Value: []byte(failedMsg.FailureTime.Format(time.RFC3339))},
			{Key: "x-retry-count", Value: []byte(fmt.Sprintf("%d", retryCount))},
		},
	}

	// Append original headers
	for _, header := range record.Headers {
		dlqRecord.Headers = append(dlqRecord.Headers, kgo.RecordHeader{
			Key:   "x-original-" + header.Key,
			Value: header.Value,
		})
	}

	// Publish to DLQ
	results := d.client.ProduceSync(ctx, dlqRecord)
	if err := results.FirstErr(); err != nil {
		return fmt.Errorf("failed to publish to DLQ: %w", err)
	}

	log.Printf("Message sent to DLQ: topic=%s, key=%s, reason=%s", dlqTopic, string(record.Key), failureReason)
	return nil
}

// ShouldRetry determines if a message should be retried or sent to DLQ
func (d *DLQManager) ShouldRetry(retryCount int) bool {
	return retryCount < d.config.MaxRetries
}

// GetDLQTopic returns the DLQ topic name for a given topic
func (d *DLQManager) GetDLQTopic(topic string) string {
	return topic + d.config.TopicSuffix
}

// convertHeaders converts kgo.RecordHeader to Header
func convertHeaders(headers []kgo.RecordHeader) []Header {
	result := make([]Header, len(headers))
	for i, h := range headers {
		result[i] = Header{
			Key:   h.Key,
			Value: string(h.Value),
		}
	}
	return result
}

// RetryFromDLQ retrieves and retries a message from the DLQ
func (d *DLQManager) RetryFromDLQ(ctx context.Context, dlqTopic string, partition int32, offset int64) error {
	if !d.config.Enabled {
		return fmt.Errorf("DLQ is not enabled")
	}

	// This would require additional consumer logic to read from DLQ and retry
	// For now, this is a placeholder for the retry functionality
	log.Printf("Retry from DLQ requested: topic=%s, partition=%d, offset=%d", dlqTopic, partition, offset)

	// TODO: Implement actual retry logic
	// 1. Fetch message from DLQ
	// 2. Parse metadata
	// 3. Reconstruct original message
	// 4. Publish to original topic

	return fmt.Errorf("retry from DLQ not yet implemented")
}
