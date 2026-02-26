package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// mockJobQueue implements repository.JobQueue for testing.
type mockJobQueue struct {
	enqueueFunc   func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error)
	dequeueFunc   func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error)
	completeFunc  func(ctx context.Context, jobID string) error
	failFunc      func(ctx context.Context, jobID string, errMsg string) error
	getStatusFunc func(ctx context.Context) (*entities.QueueStatus, error)
	getJobFunc    func(ctx context.Context, jobID string) (*entities.Job, error)
}

func (m *mockJobQueue) Enqueue(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	if m.enqueueFunc != nil {
		return m.enqueueFunc(ctx, jobType, payload)
	}
	return nil, nil
}

func (m *mockJobQueue) Dequeue(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
	if m.dequeueFunc != nil {
		return m.dequeueFunc(ctx, jobTypes)
	}
	return nil, nil
}

func (m *mockJobQueue) Complete(ctx context.Context, jobID string) error {
	if m.completeFunc != nil {
		return m.completeFunc(ctx, jobID)
	}
	return nil
}

func (m *mockJobQueue) Fail(ctx context.Context, jobID string, errMsg string) error {
	if m.failFunc != nil {
		return m.failFunc(ctx, jobID, errMsg)
	}
	return nil
}

func (m *mockJobQueue) GetStatus(ctx context.Context) (*entities.QueueStatus, error) {
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx)
	}
	return nil, nil
}

func (m *mockJobQueue) GetJob(ctx context.Context, jobID string) (*entities.Job, error) {
	if m.getJobFunc != nil {
		return m.getJobFunc(ctx, jobID)
	}
	return nil, nil
}

var _ repository.JobQueue = &mockJobQueue{}

func TestEnqueueJobUseCase_Execute_ValidJobType_ReturnsJob(t *testing.T) {
	now := time.Now()
	expectedJob := &entities.Job{
		ID:          "job-123",
		Type:        entities.JobTypeTranscription,
		Status:      entities.JobStatusCreated,
		Attempt:     0,
		MaxAttempts: entities.MaxRetryAttempts,
		CreatedAt:   now,
	}

	mock := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return expectedJob, nil
		},
	}

	uc := NewEnqueueJobUseCase(mock)
	job, err := uc.Execute(context.Background(), entities.JobTypeTranscription, []byte(`{"file":"test.mp3"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != expectedJob.ID {
		t.Errorf("expected job ID %q, got %q", expectedJob.ID, job.ID)
	}
	if job.Type != entities.JobTypeTranscription {
		t.Errorf("expected job type %q, got %q", entities.JobTypeTranscription, job.Type)
	}
}

func TestEnqueueJobUseCase_Execute_InvalidJobType_ReturnsError(t *testing.T) {
	mock := &mockJobQueue{}
	uc := NewEnqueueJobUseCase(mock)

	_, err := uc.Execute(context.Background(), "invalid_type", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error for invalid job type")
	}

	// Verify it's a ValidationError (typed error, not string matching)
	var validationErr *entities.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("expected ValidationError, got %T: %v", err, err)
	}

	expectedMsg := `invalid job type: "invalid_type"`
	if err.Error() != expectedMsg {
		t.Errorf("expected error %q, got %q", expectedMsg, err.Error())
	}
}

func TestEnqueueJobUseCase_Execute_NilPayload_DefaultsToEmptyJSON(t *testing.T) {
	var receivedPayload []byte

	mock := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			receivedPayload = payload
			return &entities.Job{ID: "job-456", Type: jobType}, nil
		},
	}

	uc := NewEnqueueJobUseCase(mock)
	_, err := uc.Execute(context.Background(), entities.JobTypeTranscription, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(receivedPayload) != "{}" {
		t.Errorf("expected payload %q, got %q", "{}", string(receivedPayload))
	}
}

func TestEnqueueJobUseCase_Execute_QueueError_PropagatesError(t *testing.T) {
	expectedErr := errors.New("database connection failed")

	mock := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return nil, expectedErr
		},
	}

	uc := NewEnqueueJobUseCase(mock)
	_, err := uc.Execute(context.Background(), entities.JobTypeTranscription, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error from queue")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %q, got %q", expectedErr, err)
	}
}

func TestEnqueueJobUseCase_Execute_AllJobTypes_Succeed(t *testing.T) {
	jobTypes := []entities.JobType{
		entities.JobTypeTranscription,
		entities.JobTypeDocumentExtraction,
		entities.JobTypeInsightGeneration,
		entities.JobTypeAnalyticsAggregation,
		entities.JobTypeRiskDetection,
		entities.JobTypeAudioCleanup,
		entities.JobTypeExtractLabResults,
	}

	for _, jt := range jobTypes {
		t.Run(string(jt), func(t *testing.T) {
			mock := &mockJobQueue{
				enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
					return &entities.Job{ID: "job-" + string(jobType), Type: jobType}, nil
				},
			}

			uc := NewEnqueueJobUseCase(mock)
			job, err := uc.Execute(context.Background(), jt, []byte(`{}`))
			if err != nil {
				t.Fatalf("unexpected error for job type %q: %v", jt, err)
			}
			if job.Type != jt {
				t.Errorf("expected job type %q, got %q", jt, job.Type)
			}
		})
	}
}

func TestEnqueueJobUseCase_Execute_ContextCancelled_PropagatesError(t *testing.T) {
	mock := &mockJobQueue{
		enqueueFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return nil, ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	uc := NewEnqueueJobUseCase(mock)
	_, err := uc.Execute(ctx, entities.JobTypeTranscription, []byte(`{}`))
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
