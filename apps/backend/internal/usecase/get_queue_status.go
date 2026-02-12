package usecase

import (
	"context"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// GetQueueStatusUseCase retrieves the current status of the job queue.
type GetQueueStatusUseCase struct {
	jobQueue repository.JobQueue
}

// NewGetQueueStatusUseCase creates a new GetQueueStatusUseCase instance.
func NewGetQueueStatusUseCase(jobQueue repository.JobQueue) *GetQueueStatusUseCase {
	return &GetQueueStatusUseCase{
		jobQueue: jobQueue,
	}
}

// Execute returns the current queue status with counts by job state.
func (uc *GetQueueStatusUseCase) Execute(ctx context.Context) (*entities.QueueStatus, error) {
	return uc.jobQueue.GetStatus(ctx)
}
