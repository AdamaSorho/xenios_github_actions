package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// InMemoryJobQueue is an in-memory implementation of JobQueue for testing.
type InMemoryJobQueue struct {
	mu   sync.RWMutex
	jobs map[string]*entities.Job
}

// NewInMemoryJobQueue creates a new InMemoryJobQueue.
func NewInMemoryJobQueue() *InMemoryJobQueue {
	return &InMemoryJobQueue{
		jobs: make(map[string]*entities.Job),
	}
}

// Enqueue adds a new job to the queue.
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
		Payload:     json.RawMessage(payload),
		Status:      entities.JobStatusCreated,
		Attempt:     0,
		MaxAttempts: entities.MaxRetryAttempts,
		CreatedAt:   time.Now(),
	}

	q.jobs[job.ID] = job
	result := *job
	return &result, nil
}

// Dequeue fetches the next available job.
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
			job.Attempt++
			result := *job
			return &result, nil
		}
	}
	return nil, nil
}

// Complete marks a job as completed.
func (q *InMemoryJobQueue) Complete(_ context.Context, jobID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}
	now := time.Now()
	job.Status = entities.JobStatusCompleted
	job.CompletedAt = &now
	return nil
}

// Fail marks a job as failed.
func (q *InMemoryJobQueue) Fail(_ context.Context, jobID string, errMsg string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	job, ok := q.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found: %s", jobID)
	}
	now := time.Now()
	job.Status = entities.JobStatusFailed
	job.FailedAt = &now
	job.ErrorMsg = &errMsg
	return nil
}

// GetStatus returns aggregate status counts.
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

	job, ok := q.jobs[jobID]
	if !ok {
		return nil, nil
	}
	result := *job
	return &result, nil
}

// GetJobs returns all jobs (for testing assertions).
func (q *InMemoryJobQueue) GetJobs() []*entities.Job {
	q.mu.RLock()
	defer q.mu.RUnlock()

	jobs := make([]*entities.Job, 0, len(q.jobs))
	for _, job := range q.jobs {
		cp := *job
		jobs = append(jobs, &cp)
	}
	return jobs
}
