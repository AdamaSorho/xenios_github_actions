package repository

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
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

func TestPostgresJobQueue_DequeueNilTypes_ReturnsNil(t *testing.T) {
	q := NewPostgresJobQueue(nil)

	job, err := q.Dequeue(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil job for nil job types")
	}
}

// createTestPool creates a pgxpool.Pool with an unreachable database.
// pgxpool.New succeeds lazily; actual connection errors occur on use.
func createTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, "postgres://test:test@localhost:19999/testdb?connect_timeout=1")
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	return pool
}

func TestPostgresJobQueue_Enqueue_UnreachableDB_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	_, err := q.Enqueue(context.Background(), entities.JobTypeTranscription, []byte(`{}`))
	if err == nil {
		t.Error("expected error when database is unreachable")
	}
}

func TestPostgresJobQueue_Dequeue_UnreachableDB_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	_, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeTranscription})
	if err == nil {
		t.Error("expected error when database is unreachable")
	}
}

func TestPostgresJobQueue_Complete_UnreachableDB_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	err := q.Complete(context.Background(), "job-123")
	if err == nil {
		t.Error("expected error when database is unreachable")
	}
}

func TestPostgresJobQueue_Fail_UnreachableDB_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	err := q.Fail(context.Background(), "job-123", "test error")
	if err == nil {
		t.Error("expected error when database is unreachable")
	}
}

func TestPostgresJobQueue_GetStatus_UnreachableDB_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	_, err := q.GetStatus(context.Background())
	if err == nil {
		t.Error("expected error when database is unreachable")
	}
}

func TestPostgresJobQueue_GetJob_UnreachableDB_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	_, err := q.GetJob(context.Background(), "job-123")
	if err == nil {
		t.Error("expected error when database is unreachable")
	}
}

func TestPostgresJobQueue_Enqueue_CancelledContext_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := q.Enqueue(ctx, entities.JobTypeTranscription, []byte(`{}`))
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

func TestPostgresJobQueue_Dequeue_CancelledContext_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := q.Dequeue(ctx, []entities.JobType{entities.JobTypeTranscription})
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

func TestPostgresJobQueue_Complete_CancelledContext_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := q.Complete(ctx, "job-123")
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

func TestPostgresJobQueue_Fail_CancelledContext_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := q.Fail(ctx, "job-123", "test error")
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

func TestPostgresJobQueue_GetStatus_CancelledContext_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := q.GetStatus(ctx)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

func TestPostgresJobQueue_GetJob_CancelledContext_ReturnsError(t *testing.T) {
	pool := createTestPool(t)
	defer pool.Close()
	q := NewPostgresJobQueue(pool)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := q.GetJob(ctx, "job-123")
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}
