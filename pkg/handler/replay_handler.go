// Package handler provides replay handler for event replay operations
package handler

import (
	"net/http"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/errors"
	"github.com/dipjyotimetia/event-shark/pkg/events"
	"github.com/dipjyotimetia/event-shark/pkg/replay"
	"github.com/gofiber/fiber/v2"
)

// ReplayHandler handles event replay operations
type ReplayHandler struct {
	replayManager *replay.ReplayManager
}

// NewReplayHandler creates a new replay handler
func NewReplayHandler(producer *events.EnhancedProducer) *ReplayHandler {
	return &ReplayHandler{
		replayManager: replay.NewReplayManager(producer.GetClient()),
	}
}

// ReplayByOffset handles offset-based replay requests
func (rh *ReplayHandler) ReplayByOffset(c *fiber.Ctx) error {
	var req struct {
		Topic       string `json:"topic"`
		StartOffset int64  `json:"start_offset"`
		EndOffset   int64  `json:"end_offset"`
		TargetTopic string `json:"target_topic"`
		MaxMessages int    `json:"max_messages"`
	}

	if err := c.BodyParser(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	if req.Topic == "" {
		appErr := errors.NewValidationError("Topic is required")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	replayReq := &replay.ReplayRequest{
		Topic:       req.Topic,
		StartOffset: req.StartOffset,
		EndOffset:   req.EndOffset,
		TargetTopic: req.TargetTopic,
		MaxMessages: req.MaxMessages,
	}

	result, err := rh.replayManager.ReplayByOffset(c.Context(), replayReq)
	if err != nil {
		appErr := errors.NewInternalError(err.Error())
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":            "completed",
		"messages_replayed": result.MessagesReplayed,
		"messages_filtered": result.MessagesFiltered,
		"duration_ms":       result.EndTime.Sub(result.StartTime).Milliseconds(),
		"errors":            len(result.Errors),
	})
}

// ReplayByTime handles time-based replay requests
func (rh *ReplayHandler) ReplayByTime(c *fiber.Ctx) error {
	var req struct {
		Topic       string `json:"topic"`
		StartTime   string `json:"start_time"`
		EndTime     string `json:"end_time"`
		TargetTopic string `json:"target_topic"`
		MaxMessages int    `json:"max_messages"`
	}

	if err := c.BodyParser(&req); err != nil {
		appErr := errors.NewValidationError("Invalid request body")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	if req.Topic == "" {
		appErr := errors.NewValidationError("Topic is required")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	if req.StartTime == "" {
		appErr := errors.NewValidationError("Start time is required")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	// Parse times
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		appErr := errors.NewValidationError("Invalid start_time format (use RFC3339)")
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	var endTime *time.Time
	if req.EndTime != "" {
		t, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			appErr := errors.NewValidationError("Invalid end_time format (use RFC3339)")
			return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
		}
		endTime = &t
	}

	replayReq := &replay.ReplayRequest{
		Topic:       req.Topic,
		StartTime:   &startTime,
		EndTime:     endTime,
		TargetTopic: req.TargetTopic,
		MaxMessages: req.MaxMessages,
	}

	result, err := rh.replayManager.ReplayByTime(c.Context(), replayReq)
	if err != nil {
		appErr := errors.NewInternalError(err.Error())
		return c.Status(appErr.GetHTTPStatus()).JSON(errors.ErrorResponse{Error: appErr})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":            "completed",
		"messages_replayed": result.MessagesReplayed,
		"messages_filtered": result.MessagesFiltered,
		"duration_ms":       result.EndTime.Sub(result.StartTime).Milliseconds(),
		"errors":            len(result.Errors),
	})
}
