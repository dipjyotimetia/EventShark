// Package async provides asynchronous event publishing with job tracking
package async

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/dipjyotimetia/event-shark/pkg/config"
	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
)

// JobStatus represents the status of an async job
type JobStatus string

const (
	JobStatusPending   JobStatus = "PENDING"
	JobStatusProcessing JobStatus = "PROCESSING"
	JobStatusCompleted  JobStatus = "COMPLETED"
	JobStatusFailed     JobStatus = "FAILED"
)

// Job represents an asynchronous publishing job
type Job struct {
	ID          string
	Status      JobStatus
	Record      *kgo.Record
	Result      interface{}
	Error       error
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

// AsyncPublisher manages asynchronous event publishing
type AsyncPublisher struct {
	mu              sync.RWMutex
	jobs            map[string]*Job
	queue           chan *Job
	workers         int
	producerFunc    func(context.Context, *kgo.Record) error
	cleanupInterval time.Duration
	jobTimeout      time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewAsyncPublisher creates a new async publisher
func NewAsyncPublisher(cfg *config.AsyncConfig, producerFunc func(context.Context, *kgo.Record) error) *AsyncPublisher {
	ctx, cancel := context.WithCancel(context.Background())

	ap := &AsyncPublisher{
		jobs:            make(map[string]*Job),
		queue:           make(chan *Job, cfg.MaxQueueSize),
		workers:         cfg.WorkerCount,
		producerFunc:    producerFunc,
		cleanupInterval: cfg.CleanupInterval,
		jobTimeout:      cfg.JobTimeout,
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start worker pool
	for i := 0; i < ap.workers; i++ {
		go ap.worker(i)
	}

	// Start cleanup routine
	go ap.cleanupLoop()

	return ap
}

// Submit submits a job for asynchronous processing
func (ap *AsyncPublisher) Submit(record *kgo.Record) (string, error) {
	jobID := uuid.New().String()

	job := &Job{
		ID:        jobID,
		Status:    JobStatusPending,
		Record:    record,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ap.mu.Lock()
	ap.jobs[jobID] = job
	ap.mu.Unlock()

	select {
	case ap.queue <- job:
		return jobID, nil
	default:
		ap.updateJobStatus(jobID, JobStatusFailed, fmt.Errorf("queue is full"))
		return "", fmt.Errorf("async queue is full, please try again later")
	}
}

// GetJobStatus returns the status of a job
func (ap *AsyncPublisher) GetJobStatus(jobID string) (*Job, error) {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	job, exists := ap.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found")
	}

	return job, nil
}

// worker processes jobs from the queue
func (ap *AsyncPublisher) worker(id int) {
	for {
		select {
		case <-ap.ctx.Done():
			return
		case job := <-ap.queue:
			ap.processJob(job)
		}
	}
}

// processJob processes a single job
func (ap *AsyncPublisher) processJob(job *Job) {
	ap.updateJobStatus(job.ID, JobStatusProcessing, nil)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ap.ctx, ap.jobTimeout)
	defer cancel()

	// Execute the producer function
	err := ap.producerFunc(ctx, job.Record)

	if err != nil {
		ap.updateJobStatus(job.ID, JobStatusFailed, err)
	} else {
		ap.updateJobStatus(job.ID, JobStatusCompleted, nil)
	}
}

// updateJobStatus updates the status of a job
func (ap *AsyncPublisher) updateJobStatus(jobID string, status JobStatus, err error) {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	job, exists := ap.jobs[jobID]
	if !exists {
		return
	}

	job.Status = status
	job.UpdatedAt = time.Now()
	job.Error = err

	if status == JobStatusCompleted || status == JobStatusFailed {
		now := time.Now()
		job.CompletedAt = &now
	}
}

// cleanupLoop periodically removes old completed jobs
func (ap *AsyncPublisher) cleanupLoop() {
	ticker := time.NewTicker(ap.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ap.ctx.Done():
			return
		case <-ticker.C:
			ap.cleanup()
		}
	}
}

// cleanup removes old completed jobs
func (ap *AsyncPublisher) cleanup() {
	ap.mu.Lock()
	defer ap.mu.Unlock()

	now := time.Now()
	for jobID, job := range ap.jobs {
		if job.CompletedAt != nil && now.Sub(*job.CompletedAt) > ap.cleanupInterval {
			delete(ap.jobs, jobID)
		}
	}
}

// GetStats returns async publisher statistics
func (ap *AsyncPublisher) GetStats() map[string]interface{} {
	ap.mu.RLock()
	defer ap.mu.RUnlock()

	stats := map[string]interface{}{
		"total_jobs":    len(ap.jobs),
		"queue_length":  len(ap.queue),
		"queue_capacity": cap(ap.queue),
		"workers":       ap.workers,
	}

	// Count jobs by status
	statusCounts := make(map[JobStatus]int)
	for _, job := range ap.jobs {
		statusCounts[job.Status]++
	}

	for status, count := range statusCounts {
		stats[fmt.Sprintf("jobs_%s", status)] = count
	}

	return stats
}

// Shutdown gracefully shuts down the async publisher
func (ap *AsyncPublisher) Shutdown(timeout time.Duration) error {
	ap.cancel()

	// Wait for queue to drain or timeout
	deadline := time.Now().Add(timeout)
	for len(ap.queue) > 0 && time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}

	close(ap.queue)
	return nil
}
