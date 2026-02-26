package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryJobQueue_Enqueue_CreatesJob(t *testing.T) {
	q := NewInMemoryJobQueue()

	job, err := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{"artifact_id":"123"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID == "" {
		t.Error("expected non-empty job ID")
	}
	if job.Type != entities.JobTypeExtractInBody {
		t.Errorf("expected type %q, got %q", entities.JobTypeExtractInBody, job.Type)
	}
	if job.Status != entities.JobStatusCreated {
		t.Errorf("expected status %q, got %q", entities.JobStatusCreated, job.Status)
	}
}

func TestInMemoryJobQueue_Dequeue_ReturnsJob(t *testing.T) {
	q := NewInMemoryJobQueue()

	enqueued, _ := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))

	job, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeExtractInBody})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected a job")
	}
	if job.ID != enqueued.ID {
		t.Errorf("expected job ID %q, got %q", enqueued.ID, job.ID)
	}
	if job.Status != entities.JobStatusActive {
		t.Errorf("expected status %q, got %q", entities.JobStatusActive, job.Status)
	}
}

func TestInMemoryJobQueue_Dequeue_NoMatchingType_ReturnsNil(t *testing.T) {
	q := NewInMemoryJobQueue()

	q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))

	job, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeTranscription})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil for non-matching job type")
	}
}

func TestInMemoryJobQueue_Complete_MarksJobCompleted(t *testing.T) {
	q := NewInMemoryJobQueue()

	enqueued, _ := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeExtractInBody})

	err := q.Complete(context.Background(), enqueued.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	job, _ := q.GetJob(context.Background(), enqueued.ID)
	if job.Status != entities.JobStatusCompleted {
		t.Errorf("expected status %q, got %q", entities.JobStatusCompleted, job.Status)
	}
}

func TestInMemoryJobQueue_Complete_NotFound_ReturnsError(t *testing.T) {
	q := NewInMemoryJobQueue()

	err := q.Complete(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInMemoryJobQueue_Fail_MarksJobFailed(t *testing.T) {
	q := NewInMemoryJobQueue()

	enqueued, _ := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeExtractInBody})

	err := q.Fail(context.Background(), enqueued.ID, "something went wrong")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	job, _ := q.GetJob(context.Background(), enqueued.ID)
	if job.Status != entities.JobStatusFailed {
		t.Errorf("expected status %q, got %q", entities.JobStatusFailed, job.Status)
	}
}

func TestInMemoryJobQueue_Fail_NotFound_ReturnsError(t *testing.T) {
	q := NewInMemoryJobQueue()

	err := q.Fail(context.Background(), "nonexistent", "error")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInMemoryJobQueue_GetStatus_CountsCorrectly(t *testing.T) {
	q := NewInMemoryJobQueue()

	q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	q.Enqueue(context.Background(), entities.JobTypeTranscribeAudio, []byte(`{}`))

	status, err := q.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Pending != 2 {
		t.Errorf("expected 2 pending, got %d", status.Pending)
	}
}

func TestInMemoryJobQueue_GetJob_ReturnsNilForMissing(t *testing.T) {
	q := NewInMemoryJobQueue()

	job, err := q.GetJob(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil for missing job")
	}
}

func TestInMemoryJobQueue_GetJobs_ReturnsAllJobs(t *testing.T) {
	q := NewInMemoryJobQueue()

	q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	q.Enqueue(context.Background(), entities.JobTypeTranscribeAudio, []byte(`{}`))

	jobs := q.GetJobs()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}

func TestInMemoryJobQueue_Dequeue_EmptyQueue_ReturnsNil(t *testing.T) {
	q := NewInMemoryJobQueue()

	job, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeExtractInBody})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil from empty queue")
	}
}
