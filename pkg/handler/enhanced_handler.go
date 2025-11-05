// Package handler provides enhanced HTTP handlers with advanced features
package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dipjyotimetia/event-shark/gen"
	"github.com/dipjyotimetia/event-shark/pkg/async"
	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/dipjyotimetia/event-shark/pkg/dlq"
	"github.com/dipjyotimetia/event-shark/pkg/errors"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/dipjyotimetia/event-shark/pkg/idempotency"
	"github.com/dipjyotimetia/event-shark/pkg/resilience"
	"github.com/dipjyotimetia/event-shark/pkg/serialization"
	"github.com/gofiber/fiber/v2"
	"github.com/hamba/avro/v2"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sr"
)

// EnhancedHandler provides advanced event handling capabilities
type EnhancedHandler struct {
	producer        *events.EnhancedProducer
	asyncPublisher  *async.AsyncPublisher
	idempotency     *idempotency.IdempotencyManager
	circuitBreaker  *resilience.CircuitBreaker
	dlqManager      *dlq.DLQManager
	serializer      *serialization.SerializerFactory
	config          *config.Config
	schemaRegistry  *sr.Client
}

// NewEnhancedHandler creates a new enhanced handler
func NewEnhancedHandler(cfg *config.Config) (*EnhancedHandler, error) {
	// Create enhanced producer
	producer, err := events.NewEnhancedProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Create schema registry client
	schemaRegistry, err := sr.NewClient(sr.URLs(cfg.SchemaRegistry))
	if err != nil {
		return nil, fmt.Errorf("failed to create schema registry client: %w", err)
	}

	// Create async publisher
	producerFunc := func(ctx context.Context, record *kgo.Record) error {
		return producer.Produce(ctx, record)
	}
	asyncPublisher := async.NewAsyncPublisher(&cfg.Async, producerFunc)

	// Create other components
	idempotencyMgr := idempotency.NewIdempotencyManager(&cfg.Idempotency)
	circuitBreaker := resilience.NewCircuitBreaker(&cfg.CircuitBreaker)
	dlqMgr := dlq.NewDLQManager(cfg, producer.GetClient())
	serializerFactory := serialization.NewSerializerFactory(schemaRegistry)

	return &EnhancedHandler{
		producer:        producer,
		asyncPublisher:  asyncPublisher,
		idempotency:     idempotencyMgr,
		circuitBreaker:  circuitBreaker,
		dlqManager:      dlqMgr,
		serializer:      serializerFactory,
		config:          cfg,
		schemaRegistry:  schemaRegistry,
	}, nil
}

// HandleExpense handles expense event publishing with all features
func (h *EnhancedHandler) HandleExpense(c *fiber.Ctx) error {
	return h.handleEvent(c, "expense-topic", gen.Expense{})
}

// HandlePayment handles payment event publishing with all features
func (h *EnhancedHandler) HandlePayment(c *fiber.Ctx) error {
	return h.handleEvent(c, "payment-topic", gen.Payment{})
}

// handleEvent is a generic handler for event publishing
func (h *EnhancedHandler) handleEvent(c *fiber.Ctx, topic string, eventType interface{}) error {
	ctx := c.Context()

	// Check idempotency key
	idempotencyKey := c.Get("X-Idempotency-Key")
	if h.config.Idempotency.Enabled && idempotencyKey != "" {
		if err := h.idempotency.ValidateIdempotencyKey(idempotencyKey); err != nil {
			appErr := err.(*errors.AppError)
			return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
		}
	}

	// Parse request body based on event type
	var event interface{}
	switch eventType.(type) {
	case gen.Expense:
		var expense gen.Expense
		if err := c.BodyParser(&expense); err != nil {
			appErr := errors.NewValidationError("Invalid request body")
			return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
		}
		if expense.Timestamp == 0 {
			expense.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}
		event = expense
	case gen.Payment:
		var payment gen.Payment
		if err := c.BodyParser(&payment); err != nil {
			appErr := errors.NewValidationError("Invalid request body")
			return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
		}
		if payment.Timestamp == 0 {
			payment.Timestamp = time.Now().UnixNano() / int64(time.Millisecond)
		}
		event = payment
	default:
		appErr := errors.NewInternalError("Unsupported event type")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	// Detect serialization format
	contentType := c.Get("Content-Type", "application/json")
	format := serialization.DetectFormat(contentType)

	// Create Kafka record
	record, err := h.createRecord(ctx, topic, event, format)
	if err != nil {
		appErr := errors.NewInternalError(fmt.Sprintf("Failed to create record: %v", err))
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	// Check if async mode is requested
	asyncMode := c.Get("X-Async", "false") == "true"

	if h.config.Async.Enabled && asyncMode {
		// Async publishing
		jobID, err := h.asyncPublisher.Submit(record)
		if err != nil {
			appErr := errors.NewInternalError(err.Error())
			return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
		}

		// Store idempotency key if provided
		if idempotencyKey != "" {
			h.idempotency.Store(idempotencyKey, jobID)
		}

		return c.Status(http.StatusAccepted).JSON(fiber.Map{
			"status": "accepted",
			"job_id": jobID,
			"message": "Event queued for publishing",
		})
	}

	// Synchronous publishing with circuit breaker
	var publishErr error
	if h.config.CircuitBreaker.Enabled {
		publishErr = h.circuitBreaker.Execute(func() error {
			return h.producer.Produce(ctx, record)
		})
	} else {
		publishErr = h.producer.Produce(ctx, record)
	}

	if publishErr != nil {
		// Send to DLQ if enabled
		if h.config.DLQ.Enabled {
			if err := h.dlqManager.SendToDLQ(ctx, record, publishErr.Error(), 0); err != nil {
				fmt.Printf("Failed to send to DLQ: %v\n", err)
			}
		}

		appErr, ok := publishErr.(*errors.AppError)
		if !ok {
			appErr = errors.NewKafkaError(publishErr.Error())
		}
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	// Store idempotency key if provided
	if idempotencyKey != "" {
		h.idempotency.Store(idempotencyKey, "success")
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status": "success",
		"message": fmt.Sprintf("%s created successfully", topic),
	})
}

// GetJobStatus returns the status of an async job
func (h *EnhancedHandler) GetJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("job_id")

	job, err := h.asyncPublisher.GetJobStatus(jobID)
	if err != nil {
		appErr := errors.NewAppError(errors.ErrCodeNotFound, "Job not found")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	return c.JSON(fiber.Map{
		"job_id":       job.ID,
		"status":       job.Status,
		"created_at":   job.CreatedAt,
		"updated_at":   job.UpdatedAt,
		"completed_at": job.CompletedAt,
		"error":        job.Error,
	})
}

// GetStats returns system statistics
func (h *EnhancedHandler) GetStats(c *fiber.Ctx) error {
	stats := fiber.Map{
		"async":           h.asyncPublisher.GetStats(),
		"idempotency":     h.idempotency.GetStats(),
		"circuit_breaker": h.circuitBreaker.GetMetrics(),
	}

	return c.JSON(stats)
}

// createRecord creates a Kafka record with serialization
func (h *EnhancedHandler) createRecord(ctx context.Context, topic string, event interface{}, format serialization.Format) (*kgo.Record, error) {
	// Get schema from registry
	schemaSubject, err := h.schemaRegistry.SchemaByVersion(ctx, topic+"-value", -1)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	avroSchema, err := avro.Parse(schemaSubject.Schema.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema: %w", err)
	}

	// Serialize based on format
	serializer, err := h.serializer.GetSerializer(format)
	if err != nil {
		return nil, err
	}

	serialized, err := serializer.Serialize(event, avroSchema)
	if err != nil {
		return nil, err
	}

	return &kgo.Record{
		Topic: topic,
		Value: serialized,
	}, nil
}

// Close closes all resources
func (h *EnhancedHandler) Close() {
	if h.asyncPublisher != nil {
		h.asyncPublisher.Shutdown(30 * time.Second)
	}
	if h.producer != nil {
		h.producer.Close()
	}
}
