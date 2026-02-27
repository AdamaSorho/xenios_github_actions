package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryJobQueue_Enqueue_ReturnsJob(t *testing.T) {
	q := NewInMemoryJobQueue()
	job, err := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{"artifact_id":"a1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID == "" {
		t.Error("expected non-empty job ID")
	}
	if job.Type != entities.JobTypeExtractInBody {
		t.Errorf("expected type %s, got %s", entities.JobTypeExtractInBody, job.Type)
	}
	if job.Status != entities.JobStatusCreated {
		t.Errorf("expected status %s, got %s", entities.JobStatusCreated, job.Status)
	}
}

func TestInMemoryJobQueue_Dequeue_ReturnsEnqueuedJob(t *testing.T) {
	q := NewInMemoryJobQueue()
	enqueued, _ := q.Enqueue(context.Background(), entities.JobTypeClassifyDocument, []byte(`{}`))

	dequeued, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeClassifyDocument})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dequeued == nil {
		t.Fatal("expected a job")
	}
	if dequeued.ID != enqueued.ID {
		t.Errorf("expected job ID %s, got %s", enqueued.ID, dequeued.ID)
	}
	if dequeued.Status != entities.JobStatusActive {
		t.Errorf("expected status %s, got %s", entities.JobStatusActive, dequeued.Status)
	}
}

func TestInMemoryJobQueue_Dequeue_EmptyQueue_ReturnsNil(t *testing.T) {
	q := NewInMemoryJobQueue()
	job, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeClassifyDocument})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil job from empty queue")
	}
}

func TestInMemoryJobQueue_Complete_UpdatesStatus(t *testing.T) {
	q := NewInMemoryJobQueue()
	job, _ := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))

	err := q.Complete(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := q.GetJob(context.Background(), job.ID)
	if found.Status != entities.JobStatusCompleted {
		t.Errorf("expected status %s, got %s", entities.JobStatusCompleted, found.Status)
	}
}

func TestInMemoryJobQueue_Complete_NotFound_ReturnsError(t *testing.T) {
	q := NewInMemoryJobQueue()
	err := q.Complete(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent job")
	}
}

func TestInMemoryJobQueue_Fail_UpdatesStatus(t *testing.T) {
	q := NewInMemoryJobQueue()
	job, _ := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))

	err := q.Fail(context.Background(), job.ID, "something went wrong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := q.GetJob(context.Background(), job.ID)
	if found.Status != entities.JobStatusFailed {
		t.Errorf("expected status %s, got %s", entities.JobStatusFailed, found.Status)
	}
	if *found.ErrorMsg != "something went wrong" {
		t.Errorf("expected error msg 'something went wrong', got '%s'", *found.ErrorMsg)
	}
}

func TestInMemoryJobQueue_Fail_NotFound_ReturnsError(t *testing.T) {
	q := NewInMemoryJobQueue()
	err := q.Fail(context.Background(), "nonexistent", "err")
	if err == nil {
		t.Error("expected error for nonexistent job")
	}
}

func TestInMemoryJobQueue_GetStatus_ReturnsCounts(t *testing.T) {
	q := NewInMemoryJobQueue()
	q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	q.Enqueue(context.Background(), entities.JobTypeClassifyDocument, []byte(`{}`))

	status, err := q.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Pending != 2 {
		t.Errorf("expected 2 pending, got %d", status.Pending)
	}
}

func TestInMemoryJobQueue_GetJob_NotFound_ReturnsError(t *testing.T) {
	q := NewInMemoryJobQueue()
	_, err := q.GetJob(context.Background(), "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent job")
	}
}

func TestInMemoryJobQueue_GetJobs_ReturnsAllJobs(t *testing.T) {
	q := NewInMemoryJobQueue()
	q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	q.Enqueue(context.Background(), entities.JobTypeClassifyDocument, []byte(`{}`))

	jobs := q.GetJobs()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}
