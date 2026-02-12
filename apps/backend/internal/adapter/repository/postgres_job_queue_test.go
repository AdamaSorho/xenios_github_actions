package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

func TestPostgresJobQueue_ImplementsJobQueueInterface(t *testing.T) {
	// Verify that PostgresJobQueue implements the repository.JobQueue interface.
	// We use a nil pool since we're only testing interface compliance, not actual DB operations.
	var _ repository.JobQueue = NewPostgresJobQueue(nil)
}

func TestNewPostgresJobQueue_ReturnsNonNil(t *testing.T) {
	q := NewPostgresJobQueue(nil)
	if q == nil {
		t.Error("expected non-nil PostgresJobQueue")
	}
}

func TestPostgresJobQueue_DequeueEmptyTypes_ReturnsNil(t *testing.T) {
	q := NewPostgresJobQueue(nil)

	// Calling Dequeue with empty job types should return nil without hitting the database
	job, err := q.Dequeue(context.Background(), []entities.JobType{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil job for empty job types")
	}
}
