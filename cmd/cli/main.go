// Package main provides the EventShark CLI tool
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/dipjyotimetia/event-shark/pkg/replay"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	version = "1.0.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "publish":
		handlePublish()
	case "replay":
		handleReplay()
	case "health":
		handleHealth()
	case "version":
		fmt.Printf("EventShark CLI v%s\n", version)
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("EventShark CLI - Kafka Event Publishing Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  eventshark-cli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  publish   - Publish an event to Kafka")
	fmt.Println("  replay    - Replay events from a topic")
	fmt.Println("  health    - Check EventShark server health")
	fmt.Println("  version   - Show version information")
	fmt.Println("  help      - Show this help message")
	fmt.Println()
	fmt.Println("Publish Options:")
	fmt.Println("  --topic <topic>       - Kafka topic name (required)")
	fmt.Println("  --key <key>           - Message key")
	fmt.Println("  --data <json>         - Message data as JSON string")
	fmt.Println("  --file <path>         - Read message data from file")
	fmt.Println()
	fmt.Println("Replay Options:")
	fmt.Println("  --source <topic>      - Source topic to replay from (required)")
	fmt.Println("  --target <topic>      - Target topic (defaults to source)")
	fmt.Println("  --start-offset <n>    - Start offset")
	fmt.Println("  --end-offset <n>      - End offset")
	fmt.Println("  --max-messages <n>    - Maximum messages to replay")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  BROKERS               - Kafka broker addresses")
	fmt.Println("  SCHEMAREGISTRY        - Schema registry URL")
	fmt.Println("  CONFIG_FILE           - Path to config file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Publish a message")
	fmt.Println("  eventshark-cli publish --topic expense-topic --data '{\"amount\": 100}'")
	fmt.Println()
	fmt.Println("  # Replay messages")
	fmt.Println("  eventshark-cli replay --source expense-topic --start-offset 0 --max-messages 100")
	fmt.Println()
}

func handlePublish() {
	var topic, key, data, file string

	// Parse arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--topic":
			if i+1 < len(os.Args) {
				topic = os.Args[i+1]
				i++
			}
		case "--key":
			if i+1 < len(os.Args) {
				key = os.Args[i+1]
				i++
			}
		case "--data":
			if i+1 < len(os.Args) {
				data = os.Args[i+1]
				i++
			}
		case "--file":
			if i+1 < len(os.Args) {
				file = os.Args[i+1]
				i++
			}
		}
	}

	if topic == "" {
		fmt.Println("Error: --topic is required")
		os.Exit(1)
	}

	// Read data from file if specified
	if file != "" {
		fileData, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		data = string(fileData)
	}

	if data == "" {
		fmt.Println("Error: --data or --file is required")
		os.Exit(1)
	}

	// Validate JSON
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(data), &jsonData); err != nil {
		fmt.Printf("Error: Invalid JSON data: %v\n", err)
		os.Exit(1)
	}

	// Load config
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create producer
	producer, err := events.NewEnhancedProducer(cfg)
	if err != nil {
		fmt.Printf("Error creating producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	// Create record
	record := &kgo.Record{
		Topic: topic,
		Key:   []byte(key),
		Value: []byte(data),
	}

	// Publish
	ctx := context.Background()
	if err := producer.Produce(ctx, record); err != nil {
		fmt.Printf("Error publishing message: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Message published successfully")
}

func handleReplay() {
	var source, target string
	var startOffset, endOffset int64 = 0, -1
	var maxMessages int

	// Parse arguments
	for i := 2; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "--source":
			if i+1 < len(os.Args) {
				source = os.Args[i+1]
				i++
			}
		case "--target":
			if i+1 < len(os.Args) {
				target = os.Args[i+1]
				i++
			}
		case "--start-offset":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &startOffset)
				i++
			}
		case "--end-offset":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &endOffset)
				i++
			}
		case "--max-messages":
			if i+1 < len(os.Args) {
				fmt.Sscanf(os.Args[i+1], "%d", &maxMessages)
				i++
			}
		}
	}

	if source == "" {
		fmt.Println("Error: --source is required")
		os.Exit(1)
	}

	if target == "" {
		target = source
	}

	// Load config
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create producer
	producer, err := events.NewEnhancedProducer(cfg)
	if err != nil {
		fmt.Printf("Error creating producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	// Create replay manager
	replayMgr := replay.NewReplayManager(producer.GetClient())

	// Create replay request
	req := &replay.ReplayRequest{
		Topic:       source,
		StartOffset: startOffset,
		EndOffset:   endOffset,
		TargetTopic: target,
		MaxMessages: maxMessages,
	}

	// Execute replay
	ctx := context.Background()
	fmt.Printf("Starting replay from %s to %s...\n", source, target)

	result, err := replayMgr.ReplayByOffset(ctx, req)
	if err != nil {
		fmt.Printf("Error during replay: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nReplay completed:\n")
	fmt.Printf("  Messages replayed: %d\n", result.MessagesReplayed)
	fmt.Printf("  Messages filtered: %d\n", result.MessagesFiltered)
	fmt.Printf("  Duration: %s\n", result.EndTime.Sub(result.StartTime))

	if len(result.Errors) > 0 {
		fmt.Printf("\nErrors encountered:\n")
		for _, err := range result.Errors {
			fmt.Printf("  - %v\n", err)
		}
	}
}

func handleHealth() {
	serverURL := os.Getenv("EVENTSHARK_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8083"
	}

	// For now, just load config and check connectivity
	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		os.Exit(1)
	}

	producer, err := events.NewEnhancedProducer(cfg)
	if err != nil {
		fmt.Printf("❌ Health check failed: Cannot connect to Kafka: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	// Try to get metadata
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := producer.Flush(ctx); err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ EventShark is healthy")
	fmt.Printf("   Brokers: %s\n", cfg.Kafka.Brokers)
	fmt.Printf("   Schema Registry: %s\n", cfg.SchemaRegistry)
}
