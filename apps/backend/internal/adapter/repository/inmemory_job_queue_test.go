package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

func TestInMemoryJobQueue_Enqueue_CreatesJob(t *testing.T) {
	q := NewInMemoryJobQueue()

	job, err := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{"artifact_id":"a1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID == "" {
		t.Fatal("expected non-empty job ID")
	}
	if job.Type != entities.JobTypeExtractInBody {
		t.Errorf("expected job type %s, got %s", entities.JobTypeExtractInBody, job.Type)
	}
	if job.Status != entities.JobStatusCreated {
		t.Errorf("expected status created, got %s", job.Status)
	}
}

func TestInMemoryJobQueue_Dequeue_ReturnsJob(t *testing.T) {
	q := NewInMemoryJobQueue()

	_, err := q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	job, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeExtractInBody})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected to dequeue a job")
	}
	if job.Status != entities.JobStatusActive {
		t.Errorf("expected status active, got %s", job.Status)
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

func TestInMemoryJobQueue_Complete_MarksCompleted(t *testing.T) {
	q := NewInMemoryJobQueue()

	job, _ := q.Enqueue(context.Background(), entities.JobTypeClassifyDocument, []byte(`{}`))

	err := q.Complete(context.Background(), job.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := q.GetJob(context.Background(), job.ID)
	if got.Status != entities.JobStatusCompleted {
		t.Errorf("expected status completed, got %s", got.Status)
	}
}

func TestInMemoryJobQueue_Fail_MarksFailed(t *testing.T) {
	q := NewInMemoryJobQueue()

	job, _ := q.Enqueue(context.Background(), entities.JobTypeClassifyDocument, []byte(`{}`))

	err := q.Fail(context.Background(), job.ID, "test error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := q.GetJob(context.Background(), job.ID)
	if got.Status != entities.JobStatusFailed {
		t.Errorf("expected status failed, got %s", got.Status)
	}
	if got.ErrorMsg == nil || *got.ErrorMsg != "test error" {
		t.Error("expected error message 'test error'")
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
		t.Fatal("expected error for nonexistent job")
	}
}

func TestInMemoryJobQueue_GetJobs_ReturnsAll(t *testing.T) {
	q := NewInMemoryJobQueue()

	q.Enqueue(context.Background(), entities.JobTypeExtractInBody, []byte(`{}`))
	q.Enqueue(context.Background(), entities.JobTypeTranscribeAudio, []byte(`{}`))

	jobs := q.GetJobs()
	if len(jobs) != 2 {
		t.Errorf("expected 2 jobs, got %d", len(jobs))
	}
}
