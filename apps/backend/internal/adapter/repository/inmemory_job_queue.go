package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryJobQueue is an in-memory implementation of JobQueue for testing.
type InMemoryJobQueue struct {
	mu   sync.RWMutex
	jobs []*entities.Job
}

// NewInMemoryJobQueue creates a new InMemoryJobQueue.
func NewInMemoryJobQueue() *InMemoryJobQueue {
	return &InMemoryJobQueue{}
}

// Enqueue adds a new job to the queue.
func (q *InMemoryJobQueue) Enqueue(_ context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	job := &entities.Job{
		ID:          fmt.Sprintf("job-%d", len(q.jobs)+1),
		Type:        jobType,
		Payload:     json.RawMessage(payload),
		Status:      entities.JobStatusCreated,
		Attempt:     0,
		MaxAttempts: entities.MaxRetryAttempts,
		CreatedAt:   now,
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
			now := time.Now()
			job.Status = entities.JobStatusActive
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
			now := time.Now()
			job.Status = entities.JobStatusCompleted
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
			now := time.Now()
			job.Status = entities.JobStatusFailed
			job.FailedAt = &now
			job.ErrorMsg = &errMsg
			return nil
		}
	}
	return fmt.Errorf("job not found: %s", jobID)
}

// GetStatus returns aggregate counts.
func (q *InMemoryJobQueue) GetStatus(_ context.Context) (*entities.QueueStatus, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()

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
	q.mu.RLock()
	defer q.mu.RUnlock()

	for _, job := range q.jobs {
		if job.ID == jobID {
			return job, nil
		}
	}
	return nil, fmt.Errorf("job not found: %s", jobID)
}

// GetJobs returns all jobs (for testing).
func (q *InMemoryJobQueue) GetJobs() []*entities.Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	cp := make([]*entities.Job, len(q.jobs))
	copy(cp, q.jobs)
	return cp
}
