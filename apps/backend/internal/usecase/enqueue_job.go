package usecase

import (
	"context"
	"fmt"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// EnqueueJobUseCase handles enqueueing new jobs to the job queue.
type EnqueueJobUseCase struct {
	jobQueue repository.JobQueue
}

// NewEnqueueJobUseCase creates a new EnqueueJobUseCase instance.
func NewEnqueueJobUseCase(jobQueue repository.JobQueue) *EnqueueJobUseCase {
	return &EnqueueJobUseCase{
		jobQueue: jobQueue,
	}
}

// Execute enqueues a new job with the given type and payload.
// Returns the created job or an error if the job type is invalid or enqueueing fails.
func (uc *EnqueueJobUseCase) Execute(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	if !entities.IsValidJobType(jobType) {
		return nil, fmt.Errorf("invalid job type: %q", jobType)
	}

	if payload == nil {
		payload = []byte(`{}`)
	}

	return uc.jobQueue.Enqueue(ctx, jobType, payload)
}
