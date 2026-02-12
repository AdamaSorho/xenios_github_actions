package worker

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// mockJobQueue implements repository.JobQueue for testing.
type mockJobQueue struct {
	mu          sync.Mutex
	enqueueFunc func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error)
	dequeueFunc func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error)
	completeFunc func(ctx context.Context, jobID string) error
	failFunc     func(ctx context.Context, jobID string, errMsg string) error
	getStatusFunc func(ctx context.Context) (*entities.QueueStatus, error)
	getJobFunc    func(ctx context.Context, jobID string) (*entities.Job, error)
}

func (m *mockJobQueue) Enqueue(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.enqueueFunc != nil {
		return m.enqueueFunc(ctx, jobType, payload)
	}
	return nil, nil
}

func (m *mockJobQueue) Dequeue(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dequeueFunc != nil {
		return m.dequeueFunc(ctx, jobTypes)
	}
	return nil, nil
}

func (m *mockJobQueue) Complete(ctx context.Context, jobID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.completeFunc != nil {
		return m.completeFunc(ctx, jobID)
	}
	return nil
}

func (m *mockJobQueue) Fail(ctx context.Context, jobID string, errMsg string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failFunc != nil {
		return m.failFunc(ctx, jobID, errMsg)
	}
	return nil
}

func (m *mockJobQueue) GetStatus(ctx context.Context) (*entities.QueueStatus, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getStatusFunc != nil {
		return m.getStatusFunc(ctx)
	}
	return nil, nil
}

func (m *mockJobQueue) GetJob(ctx context.Context, jobID string) (*entities.Job, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.getJobFunc != nil {
		return m.getJobFunc(ctx, jobID)
	}
	return nil, nil
}

var _ repository.JobQueue = &mockJobQueue{}

func TestNewWorker_DefaultValues(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, 0, 0)

	if w == nil {
		t.Fatal("expected non-nil worker")
	}
	if w.pollInterval != 5*time.Second {
		t.Errorf("expected default poll interval 5s, got %v", w.pollInterval)
	}
	if w.jobTimeout != 5*time.Minute {
		t.Errorf("expected default job timeout 5m, got %v", w.jobTimeout)
	}
}

func TestNewWorker_CustomValues(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, 10*time.Second, 30*time.Minute)

	if w.pollInterval != 10*time.Second {
		t.Errorf("expected poll interval 10s, got %v", w.pollInterval)
	}
	if w.jobTimeout != 30*time.Minute {
		t.Errorf("expected job timeout 30m, got %v", w.jobTimeout)
	}
}

func TestWorker_RegisterHandler_AddsHandler(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, time.Second, time.Minute)

	handler := func(ctx context.Context, job *entities.Job) error {
		return nil
	}

	w.RegisterHandler(entities.JobTypeTranscription, handler)

	types := w.RegisteredJobTypes()
	if len(types) != 1 {
		t.Fatalf("expected 1 registered type, got %d", len(types))
	}
	if types[0] != entities.JobTypeTranscription {
		t.Errorf("expected type %q, got %q", entities.JobTypeTranscription, types[0])
	}
}

func TestWorker_RegisterHandler_MultipleHandlers(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, time.Second, time.Minute)

	handler := func(ctx context.Context, job *entities.Job) error {
		return nil
	}

	w.RegisterHandler(entities.JobTypeTranscription, handler)
	w.RegisterHandler(entities.JobTypeDocumentExtraction, handler)
	w.RegisterHandler(entities.JobTypeRiskDetection, handler)

	types := w.RegisteredJobTypes()
	if len(types) != 3 {
		t.Errorf("expected 3 registered types, got %d", len(types))
	}
}

func TestWorker_StartStop(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, 50*time.Millisecond, time.Minute)

	ctx := context.Background()
	w.Start(ctx)

	if !w.IsRunning() {
		t.Error("expected worker to be running")
	}

	w.Stop()

	if w.IsRunning() {
		t.Error("expected worker to be stopped")
	}
}

func TestWorker_Start_DoubleStartIsNoop(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, 50*time.Millisecond, time.Minute)

	ctx := context.Background()
	w.Start(ctx)
	w.Start(ctx) // Should be a no-op

	if !w.IsRunning() {
		t.Error("expected worker to still be running after double start")
	}

	w.Stop()
}

func TestWorker_Stop_DoubleStopIsNoop(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, 50*time.Millisecond, time.Minute)

	ctx := context.Background()
	w.Start(ctx)
	w.Stop()
	w.Stop() // Should be a no-op

	if w.IsRunning() {
		t.Error("expected worker to be stopped")
	}
}

func TestWorker_ProcessesJob_HappyPath(t *testing.T) {
	jobProcessed := make(chan string, 1)
	jobCompleted := make(chan string, 1)

	testJob := &entities.Job{
		ID:          "job-123",
		Type:        entities.JobTypeTranscription,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: 3,
	}

	dequeueCount := 0
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			dequeueCount++
			if dequeueCount == 1 {
				return testJob, nil
			}
			return nil, nil
		},
		completeFunc: func(ctx context.Context, jobID string) error {
			jobCompleted <- jobID
			return nil
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	w.RegisterHandler(entities.JobTypeTranscription, func(ctx context.Context, job *entities.Job) error {
		jobProcessed <- job.ID
		return nil
	})

	ctx := context.Background()
	w.Start(ctx)
	defer w.Stop()

	select {
	case id := <-jobProcessed:
		if id != "job-123" {
			t.Errorf("expected job ID %q, got %q", "job-123", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for job to be processed")
	}

	select {
	case id := <-jobCompleted:
		if id != "job-123" {
			t.Errorf("expected completed job ID %q, got %q", "job-123", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for job to be completed")
	}
}

func TestWorker_ProcessesJob_FailureCallsFail(t *testing.T) {
	jobFailed := make(chan string, 1)

	testJob := &entities.Job{
		ID:          "job-456",
		Type:        entities.JobTypeTranscription,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: 3,
	}

	dequeueCount := 0
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			dequeueCount++
			if dequeueCount == 1 {
				return testJob, nil
			}
			return nil, nil
		},
		failFunc: func(ctx context.Context, jobID string, errMsg string) error {
			jobFailed <- jobID
			return nil
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	w.RegisterHandler(entities.JobTypeTranscription, func(ctx context.Context, job *entities.Job) error {
		return errors.New("processing failed")
	})

	ctx := context.Background()
	w.Start(ctx)
	defer w.Stop()

	select {
	case id := <-jobFailed:
		if id != "job-456" {
			t.Errorf("expected failed job ID %q, got %q", "job-456", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for job to be failed")
	}
}

func TestWorker_NoHandlers_SkipsProcessing(t *testing.T) {
	dequeueCalled := false
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			dequeueCalled = true
			return nil, nil
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	// No handlers registered

	ctx := context.Background()
	w.Start(ctx)
	time.Sleep(150 * time.Millisecond)
	w.Stop()

	if dequeueCalled {
		t.Error("expected dequeue not to be called when no handlers are registered")
	}
}

func TestWorker_ContextCancellation_StopsWorker(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, 50*time.Millisecond, time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	if !w.IsRunning() {
		t.Error("expected worker to be running")
	}

	cancel()
	// Give the worker time to notice the cancellation
	time.Sleep(200 * time.Millisecond)

	// Worker loop exits and running flag is reset to false
	if w.IsRunning() {
		t.Error("expected worker to not be running after context cancellation")
	}
}

func TestWorker_UnregisteredJobType_FailsJob(t *testing.T) {
	jobFailed := make(chan string, 1)

	testJob := &entities.Job{
		ID:          "job-789",
		Type:        entities.JobTypeDocumentExtraction,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: 3,
	}

	dequeueCount := 0
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			dequeueCount++
			if dequeueCount == 1 {
				return testJob, nil
			}
			return nil, nil
		},
		failFunc: func(ctx context.Context, jobID string, errMsg string) error {
			jobFailed <- jobID
			return nil
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	// Register handler for transcription, but dequeue returns document_extraction
	w.RegisterHandler(entities.JobTypeTranscription, func(ctx context.Context, job *entities.Job) error {
		return nil
	})

	ctx := context.Background()
	w.Start(ctx)
	defer w.Stop()

	select {
	case id := <-jobFailed:
		if id != "job-789" {
			t.Errorf("expected failed job ID %q, got %q", "job-789", id)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for unregistered job to be failed")
	}
}

func TestWorker_DequeueError_ContinuesPolling(t *testing.T) {
	errorCount := 0
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			errorCount++
			if errorCount <= 2 {
				return nil, errors.New("temporary error")
			}
			return nil, nil
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	w.RegisterHandler(entities.JobTypeTranscription, func(ctx context.Context, job *entities.Job) error {
		return nil
	})

	ctx := context.Background()
	w.Start(ctx)
	time.Sleep(300 * time.Millisecond)
	w.Stop()

	if errorCount < 2 {
		t.Errorf("expected at least 2 dequeue errors, got %d", errorCount)
	}
}

func TestWorker_ContextCancellation_AllowsRestart(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, 50*time.Millisecond, time.Minute)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	if !w.IsRunning() {
		t.Error("expected worker to be running")
	}

	cancel()
	time.Sleep(200 * time.Millisecond)

	if w.IsRunning() {
		t.Error("expected worker to not be running after context cancellation")
	}

	// Should be able to restart after context cancellation
	ctx2 := context.Background()
	w.Start(ctx2)
	if !w.IsRunning() {
		t.Error("expected worker to be running after restart")
	}
	w.Stop()
}

func TestWorker_IsRunning_InitiallyFalse(t *testing.T) {
	mock := &mockJobQueue{}
	w := NewWorker(mock, time.Second, time.Minute)

	if w.IsRunning() {
		t.Error("expected worker to not be running initially")
	}
}

func TestWorker_CompleteError_IsLogged(t *testing.T) {
	completeCalled := make(chan struct{}, 1)

	testJob := &entities.Job{
		ID:          "job-complete-err",
		Type:        entities.JobTypeTranscription,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: 3,
	}

	dequeueCount := 0
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			dequeueCount++
			if dequeueCount == 1 {
				return testJob, nil
			}
			return nil, nil
		},
		completeFunc: func(ctx context.Context, jobID string) error {
			completeCalled <- struct{}{}
			return errors.New("complete failed: db error")
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	w.RegisterHandler(entities.JobTypeTranscription, func(ctx context.Context, job *entities.Job) error {
		return nil // success
	})

	ctx := context.Background()
	w.Start(ctx)
	defer w.Stop()

	select {
	case <-completeCalled:
		// Complete was called and returned error - the error is logged, not fatal
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for complete to be called")
	}
}

func TestWorker_FailError_AfterHandlerFailure_IsLogged(t *testing.T) {
	failCalled := make(chan struct{}, 1)

	testJob := &entities.Job{
		ID:          "job-fail-err",
		Type:        entities.JobTypeTranscription,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: 3,
	}

	dequeueCount := 0
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			dequeueCount++
			if dequeueCount == 1 {
				return testJob, nil
			}
			return nil, nil
		},
		failFunc: func(ctx context.Context, jobID string, errMsg string) error {
			failCalled <- struct{}{}
			return errors.New("fail failed: db error")
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	w.RegisterHandler(entities.JobTypeTranscription, func(ctx context.Context, job *entities.Job) error {
		return errors.New("handler error")
	})

	ctx := context.Background()
	w.Start(ctx)
	defer w.Stop()

	select {
	case <-failCalled:
		// Fail was called and returned error - the error is logged, not fatal
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for fail to be called")
	}
}

func TestWorker_FailError_NoHandler_IsLogged(t *testing.T) {
	failCalled := make(chan struct{}, 1)

	testJob := &entities.Job{
		ID:          "job-no-handler-err",
		Type:        entities.JobTypeDocumentExtraction,
		Status:      entities.JobStatusActive,
		Attempt:     1,
		MaxAttempts: 3,
	}

	dequeueCount := 0
	mock := &mockJobQueue{
		dequeueFunc: func(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
			dequeueCount++
			if dequeueCount == 1 {
				return testJob, nil
			}
			return nil, nil
		},
		failFunc: func(ctx context.Context, jobID string, errMsg string) error {
			failCalled <- struct{}{}
			return errors.New("fail failed: db error")
		},
	}

	w := NewWorker(mock, 50*time.Millisecond, time.Minute)
	// Only register handler for transcription, but dequeue returns document_extraction
	w.RegisterHandler(entities.JobTypeTranscription, func(ctx context.Context, job *entities.Job) error {
		return nil
	})

	ctx := context.Background()
	w.Start(ctx)
	defer w.Stop()

	select {
	case <-failCalled:
		// Fail was called for unregistered type and returned error
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for fail to be called")
	}
}
