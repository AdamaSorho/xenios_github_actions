package repository

import (
	"context"
	"testing"

	"github.com/xenios/backend/internal/domain/entities"
)

// Verify that JobQueue interface methods are correctly defined by implementing a mock.
type mockJobQueue struct {
	enqueueFunc   func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error)
	dequeueFunc   func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error)
	completeFunc  func(ctx context.Context, jobID string) error
	failFunc      func(ctx context.Context, jobID string, errMsg string) error
	getStatusFunc func(ctx context.Context) (*entities.QueueStatus, error)
	getJobFunc    func(ctx context.Context, jobID string) (*entities.Job, error)
}

func (m *mockJobQueue) Enqueue(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	return m.enqueueFunc(ctx, jobType, payload)
}

func (m *mockJobQueue) Dequeue(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
	return m.dequeueFunc(ctx, jobTypes)
}

func (m *mockJobQueue) Complete(ctx context.Context, jobID string) error {
	return m.completeFunc(ctx, jobID)
}

func (m *mockJobQueue) Fail(ctx context.Context, jobID string, errMsg string) error {
	return m.failFunc(ctx, jobID, errMsg)
}

func (m *mockJobQueue) GetStatus(ctx context.Context) (*entities.QueueStatus, error) {
	return m.getStatusFunc(ctx)
}

func (m *mockJobQueue) GetJob(ctx context.Context, jobID string) (*entities.Job, error) {
	return m.getJobFunc(ctx, jobID)
}

func TestJobQueue_InterfaceCompliance(t *testing.T) {
	var _ JobQueue = &mockJobQueue{}
}

func TestJobQueue_Enqueue_ReturnsJob(t *testing.T) {
	expectedJob := &entities.Job{
		ID:     "test-123",
		Type:   entities.JobTypeTranscription,
		Status: entities.JobStatusCreated,
	}

	mock := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return expectedJob, nil
		},
	}

	var q JobQueue = mock
	job, err := q.Enqueue(context.Background(), entities.JobTypeTranscription, []byte(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != expectedJob.ID {
		t.Errorf("expected job ID %q, got %q", expectedJob.ID, job.ID)
	}
}

func TestJobQueue_Dequeue_ReturnsNilWhenEmpty(t *testing.T) {
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			return nil, nil
		},
	}

	var q JobQueue = mock
	job, err := q.Dequeue(context.Background(), []entities.JobType{entities.JobTypeTranscription})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job != nil {
		t.Error("expected nil job when queue is empty")
	}
}

func TestJobQueue_GetStatus_ReturnsQueueStatus(t *testing.T) {
	expected := &entities.QueueStatus{
		Pending: 5,
		Active:  2,
	}

	mock := &mockJobQueue{
		getStatusFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return expected, nil
		},
	}

	var q JobQueue = mock
	status, err := q.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Pending != expected.Pending {
		t.Errorf("expected pending %d, got %d", expected.Pending, status.Pending)
	}
}
