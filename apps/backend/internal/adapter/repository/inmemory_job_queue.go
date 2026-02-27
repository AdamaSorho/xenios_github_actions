package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryJobQueue is an in-memory implementation of JobQueue for testing.
type InMemoryJobQueue struct {
	mu   sync.Mutex
	jobs []*entities.Job
}

// NewInMemoryJobQueue creates a new InMemoryJobQueue.
func NewInMemoryJobQueue() *InMemoryJobQueue {
	return &InMemoryJobQueue{}
}

// Enqueue adds a job to the queue.
func (q *InMemoryJobQueue) Enqueue(_ context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("generate id: %w", err)
	}

	job := &entities.Job{
		ID:          hex.EncodeToString(b),
		Type:        jobType,
		Payload:     payload,
		Status:      entities.JobStatusCreated,
		Attempt:     0,
		MaxAttempts: entities.MaxRetryAttempts,
		CreatedAt:   time.Now(),
	}
	q.jobs = append(q.jobs, job)
	return job, nil
}

// Dequeue returns the next available job.
func (q *InMemoryJobQueue) Dequeue(_ context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	typeSet := make(map[entities.JobType]bool, len(jobTypes))
	for _, jt := range jobTypes {
		typeSet[jt] = true
	}

	for _, job := range q.jobs {
		if job.Status == entities.JobStatusCreated && typeSet[job.Type] {
			job.Status = entities.JobStatusActive
			now := time.Now()
			job.StartedAt = &now
			return job, nil
		}
	}
	return nil, nil
}

// Complete marks a job as completed.
func (q *InMemoryJobQueue) Complete(_ context.Context, jobID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, job := range q.jobs {
		if job.ID == jobID {
			job.Status = entities.JobStatusCompleted
			now := time.Now()
			job.CompletedAt = &now
			return nil
		}
	}
	return fmt.Errorf("job not found: %s", jobID)
}

// Fail marks a job as failed.
func (q *InMemoryJobQueue) Fail(_ context.Context, jobID string, errMsg string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, job := range q.jobs {
		if job.ID == jobID {
			job.Status = entities.JobStatusFailed
			now := time.Now()
			job.FailedAt = &now
			job.ErrorMsg = &errMsg
			return nil
		}
	}
	return fmt.Errorf("job not found: %s", jobID)
}

// GetStatus returns queue status counts.
func (q *InMemoryJobQueue) GetStatus(_ context.Context) (*entities.QueueStatus, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	status := &entities.QueueStatus{}
	for _, job := range q.jobs {
		switch job.Status {
		case entities.JobStatusCreated:
			status.Pending++
		case entities.JobStatusActive:
			status.Active++
		case entities.JobStatusCompleted:
			status.Completed++
		case entities.JobStatusFailed:
			status.Failed++
		case entities.JobStatusExpired:
			status.Expired++
		}
	}
	return status, nil
}

// GetJob returns a job by ID.
func (q *InMemoryJobQueue) GetJob(_ context.Context, jobID string) (*entities.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, job := range q.jobs {
		if job.ID == jobID {
			return job, nil
		}
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

// GetJobs returns a copy of all jobs (for testing).
func (q *InMemoryJobQueue) GetJobs() []*entities.Job {
	q.mu.Lock()
	defer q.mu.Unlock()
	cp := make([]*entities.Job, len(q.jobs))
	copy(cp, q.jobs)
	return cp
}
