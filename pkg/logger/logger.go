// Package logger provides structured logging utilities for the Event Shark application.
package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with application-specific methods.
type Logger struct {
	*slog.Logger
}

// New creates a new Logger instance with structured logging.
func New() *Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithContext adds context values to the logger.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", getRequestID(ctx)),
	}
}

// LogError logs an error with additional context.
func (l *Logger) LogError(ctx context.Context, err error, msg string, args ...interface{}) {
	l.WithContext(ctx).Error(msg, "error", err, "details", args)
}

// LogInfo logs an info message with context.
func (l *Logger) LogInfo(ctx context.Context, msg string, args ...interface{}) {
	l.WithContext(ctx).Info(msg, "details", args)
}

// LogWarn logs a warning message with context.
func (l *Logger) LogWarn(ctx context.Context, msg string, args ...interface{}) {
	l.WithContext(ctx).Warn(msg, "details", args)
}

// LogKafkaEvent logs Kafka-related events.
func (l *Logger) LogKafkaEvent(ctx context.Context, topic string, partition int32, offset int64, msg string) {
	l.WithContext(ctx).Info("kafka_event",
		"message", msg,
		"topic", topic,
		"partition", partition,
		"offset", offset,
	)
}

// getRequestID extracts request ID from context or returns empty string.
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}
