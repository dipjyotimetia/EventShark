// Package replay provides event replay capabilities from Kafka
package replay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

// ReplayRequest represents a replay request
type ReplayRequest struct {
	Topic         string
	StartOffset   int64
	EndOffset     int64
	StartTime     *time.Time
	EndTime       *time.Time
	TargetTopic   string
	FilterPattern string
	MaxMessages   int
}

// ReplayResult represents the result of a replay operation
type ReplayResult struct {
	MessagesReplayed int
	MessagesFiltered int
	StartTime        time.Time
	EndTime          time.Time
	Errors           []error
}

// ReplayManager manages event replay operations
type ReplayManager struct {
	client *kgo.Client
}

// NewReplayManager creates a new replay manager
func NewReplayManager(client *kgo.Client) *ReplayManager {
	return &ReplayManager{
		client: client,
	}
}

// ReplayByOffset replays messages from a topic by offset range
func (rm *ReplayManager) ReplayByOffset(ctx context.Context, req *ReplayRequest) (*ReplayResult, error) {
	if req.Topic == "" {
		return nil, fmt.Errorf("source topic is required")
	}

	if req.TargetTopic == "" {
		req.TargetTopic = req.Topic
	}

	result := &ReplayResult{
		StartTime: time.Now(),
	}

	// Create a new consumer for replay
	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(rm.client.OptValue(kgo.SeedBrokers).([]string)...),
		kgo.ConsumeTopics(req.Topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().At(req.StartOffset)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create replay consumer: %w", err)
	}
	defer consumer.Close()

	log.Printf("Starting replay from topic %s (offset %d to %d)", req.Topic, req.StartOffset, req.EndOffset)

	replayedCount := 0
	filteredCount := 0

	for {
		fetches := consumer.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			result.Errors = append(result.Errors, err)
			break
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			// Check if we've reached the end offset
			if req.EndOffset > 0 && record.Offset > req.EndOffset {
				goto done
			}

			// Check if we've hit max messages
			if req.MaxMessages > 0 && replayedCount >= req.MaxMessages {
				goto done
			}

			// Apply filter if specified
			if req.FilterPattern != "" {
				// TODO: Implement filter pattern matching
				filteredCount++
				continue
			}

			// Create new record for replay
			replayRecord := &kgo.Record{
				Topic: req.TargetTopic,
				Key:   record.Key,
				Value: record.Value,
				Headers: []kgo.RecordHeader{
					{Key: "x-replayed-from", Value: []byte(req.Topic)},
					{Key: "x-original-offset", Value: []byte(fmt.Sprintf("%d", record.Offset))},
					{Key: "x-original-partition", Value: []byte(fmt.Sprintf("%d", record.Partition))},
					{Key: "x-replay-time", Value: []byte(time.Now().Format(time.RFC3339))},
				},
			}

			// Append original headers
			for _, h := range record.Headers {
				replayRecord.Headers = append(replayRecord.Headers, h)
			}

			// Publish replayed message
			results := rm.client.ProduceSync(ctx, replayRecord)
			if err := results.FirstErr(); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to replay message at offset %d: %w", record.Offset, err))
				continue
			}

			replayedCount++
		}
	}

done:
	result.MessagesReplayed = replayedCount
	result.MessagesFiltered = filteredCount
	result.EndTime = time.Now()

	log.Printf("Replay completed: %d messages replayed, %d filtered, duration: %s",
		result.MessagesReplayed, result.MessagesFiltered, result.EndTime.Sub(result.StartTime))

	return result, nil
}

// ReplayByTime replays messages from a topic by time range
func (rm *ReplayManager) ReplayByTime(ctx context.Context, req *ReplayRequest) (*ReplayResult, error) {
	if req.StartTime == nil {
		return nil, fmt.Errorf("start time is required for time-based replay")
	}

	if req.Topic == "" {
		return nil, fmt.Errorf("source topic is required")
	}

	if req.TargetTopic == "" {
		req.TargetTopic = req.Topic
	}

	result := &ReplayResult{
		StartTime: time.Now(),
	}

	// Create a new consumer for replay with time-based offset
	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(rm.client.OptValue(kgo.SeedBrokers).([]string)...),
		kgo.ConsumeTopics(req.Topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create replay consumer: %w", err)
	}
	defer consumer.Close()

	log.Printf("Starting time-based replay from topic %s (from %s)", req.Topic, req.StartTime.Format(time.RFC3339))

	replayedCount := 0
	filteredCount := 0

	for {
		fetches := consumer.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			result.Errors = append(result.Errors, err)
			break
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			// Check time constraints
			recordTime := record.Timestamp
			if recordTime.Before(*req.StartTime) {
				continue
			}

			if req.EndTime != nil && recordTime.After(*req.EndTime) {
				goto done
			}

			// Check max messages
			if req.MaxMessages > 0 && replayedCount >= req.MaxMessages {
				goto done
			}

			// Create new record for replay
			replayRecord := &kgo.Record{
				Topic: req.TargetTopic,
				Key:   record.Key,
				Value: record.Value,
				Headers: []kgo.RecordHeader{
					{Key: "x-replayed-from", Value: []byte(req.Topic)},
					{Key: "x-original-timestamp", Value: []byte(record.Timestamp.Format(time.RFC3339))},
					{Key: "x-replay-time", Value: []byte(time.Now().Format(time.RFC3339))},
				},
			}

			// Append original headers
			for _, h := range record.Headers {
				replayRecord.Headers = append(replayRecord.Headers, h)
			}

			// Publish replayed message
			results := rm.client.ProduceSync(ctx, replayRecord)
			if err := results.FirstErr(); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to replay message: %w", err))
				continue
			}

			replayedCount++
		}
	}

done:
	result.MessagesReplayed = replayedCount
	result.MessagesFiltered = filteredCount
	result.EndTime = time.Now()

	log.Printf("Time-based replay completed: %d messages replayed, duration: %s",
		result.MessagesReplayed, result.EndTime.Sub(result.StartTime))

	return result, nil
}
