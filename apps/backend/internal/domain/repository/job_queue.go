package repository

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
)

// JobQueue defines the interface for job queue operations.
// Implementations handle enqueueing, dequeueing, status tracking, and dead letter management.
type JobQueue interface {
	// Enqueue adds a new job to the queue and returns the created job.
	Enqueue(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error)

	// Dequeue fetches and locks the next available job for processing.
	// Returns nil if no jobs are available.
	Dequeue(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error)

	// Complete marks a job as successfully completed.
	Complete(ctx context.Context, jobID string) error

	// Fail marks a job as failed. If retries remain, schedules a retry with backoff.
	// If max retries exceeded, moves to dead letter queue.
	Fail(ctx context.Context, jobID string, errMsg string) error

	// GetStatus returns aggregate counts of jobs by status.
	GetStatus(ctx context.Context) (*entities.QueueStatus, error)

	// GetJob returns a single job by ID.
	GetJob(ctx context.Context, jobID string) (*entities.Job, error)
}
