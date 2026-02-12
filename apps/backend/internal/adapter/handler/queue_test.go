package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xenios/backend/internal/domain/entities"
)

// mockEnqueueJobUseCase implements EnqueueJobUseCase for testing.
type mockEnqueueJobUseCase struct {
	executeFunc func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error)
}

func (m *mockEnqueueJobUseCase) Execute(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, jobType, payload)
	}
	return nil, nil
}

// mockGetQueueStatusUseCase implements GetQueueStatusUseCaseInterface for testing.
type mockGetQueueStatusUseCase struct {
	executeFunc func(ctx context.Context) (*entities.QueueStatus, error)
}

func (m *mockGetQueueStatusUseCase) Execute(ctx context.Context) (*entities.QueueStatus, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx)
	}
	return nil, nil
}

func TestQueueHandler_EnqueueJob_Success(t *testing.T) {
	now := time.Now()
	expectedJob := &entities.Job{
		ID:          "job-123",
		Type:        entities.JobTypeTranscription,
		Status:      entities.JobStatusCreated,
		Attempt:     0,
		MaxAttempts: 3,
		CreatedAt:   now,
	}

	enqueueUC := &mockEnqueueJobUseCase{
		executeFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return expectedJob, nil
		},
	}
	statusUC := &mockGetQueueStatusUseCase{}

	h := NewQueueHandler(enqueueUC, statusUC)

	body := `{"type":"transcription","payload":{"file":"audio.mp3"}}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.EnqueueJob(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	var response entities.Job
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.ID != "job-123" {
		t.Errorf("expected job ID %q, got %q", "job-123", response.ID)
	}
}

func TestQueueHandler_EnqueueJob_InvalidJSON(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{}
	h := NewQueueHandler(enqueueUC, statusUC)

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString("not json"))
	w := httptest.NewRecorder()

	h.EnqueueJob(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Error != "invalid request body" {
		t.Errorf("expected error %q, got %q", "invalid request body", response.Error)
	}
}

func TestQueueHandler_EnqueueJob_MissingType(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{}
	h := NewQueueHandler(enqueueUC, statusUC)

	body := `{"payload":{"file":"audio.mp3"}}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.EnqueueJob(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Error != "job type is required" {
		t.Errorf("expected error %q, got %q", "job type is required", response.Error)
	}
}

func TestQueueHandler_EnqueueJob_InvalidJobType(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{
		executeFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return nil, errors.New(`invalid job type: "invalid"`)
		},
	}
	statusUC := &mockGetQueueStatusUseCase{}
	h := NewQueueHandler(enqueueUC, statusUC)

	body := `{"type":"invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.EnqueueJob(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestQueueHandler_EnqueueJob_WithoutPayload(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{
		executeFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return &entities.Job{ID: "job-789", Type: jobType}, nil
		},
	}
	statusUC := &mockGetQueueStatusUseCase{}
	h := NewQueueHandler(enqueueUC, statusUC)

	body := `{"type":"transcription"}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.EnqueueJob(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}
}

func TestQueueHandler_GetQueueStatus_Healthy(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return &entities.QueueStatus{
				Pending:    5,
				Active:     2,
				Completed:  100,
				Failed:     0,
				Expired:    0,
				DeadLetter: 0,
			}, nil
		},
	}
	h := NewQueueHandler(enqueueUC, statusUC)

	req := httptest.NewRequest(http.MethodGet, "/jobs/status", nil)
	w := httptest.NewRecorder()

	h.GetQueueStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response QueueStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "healthy" {
		t.Errorf("expected status %q, got %q", "healthy", response.Status)
	}
	if response.Queue.Pending != 5 {
		t.Errorf("expected pending %d, got %d", 5, response.Queue.Pending)
	}
	if response.Queue.Active != 2 {
		t.Errorf("expected active %d, got %d", 2, response.Queue.Active)
	}
}

func TestQueueHandler_GetQueueStatus_Degraded_WithFailedJobs(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return &entities.QueueStatus{
				Pending:    5,
				Active:     2,
				Completed:  100,
				Failed:     3,
				Expired:    0,
				DeadLetter: 0,
			}, nil
		},
	}
	h := NewQueueHandler(enqueueUC, statusUC)

	req := httptest.NewRequest(http.MethodGet, "/jobs/status", nil)
	w := httptest.NewRecorder()

	h.GetQueueStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response QueueStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "degraded" {
		t.Errorf("expected status %q, got %q", "degraded", response.Status)
	}
}

func TestQueueHandler_GetQueueStatus_Degraded_WithDeadLetterJobs(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return &entities.QueueStatus{
				Pending:    0,
				Active:     0,
				Completed:  50,
				Failed:     0,
				Expired:    0,
				DeadLetter: 2,
			}, nil
		},
	}
	h := NewQueueHandler(enqueueUC, statusUC)

	req := httptest.NewRequest(http.MethodGet, "/jobs/status", nil)
	w := httptest.NewRecorder()

	h.GetQueueStatus(w, req)

	var response QueueStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Status != "degraded" {
		t.Errorf("expected status %q, got %q", "degraded", response.Status)
	}
}

func TestQueueHandler_GetQueueStatus_Error(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return nil, errors.New("database error")
		},
	}
	h := NewQueueHandler(enqueueUC, statusUC)

	req := httptest.NewRequest(http.MethodGet, "/jobs/status", nil)
	w := httptest.NewRecorder()

	h.GetQueueStatus(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestQueueHandler_GetQueueStatus_ResponseContentType(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{
		executeFunc: func(ctx context.Context) (*entities.QueueStatus, error) {
			return &entities.QueueStatus{}, nil
		},
	}
	h := NewQueueHandler(enqueueUC, statusUC)

	req := httptest.NewRequest(http.MethodGet, "/jobs/status", nil)
	w := httptest.NewRecorder()

	h.GetQueueStatus(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}

func TestNewQueueHandler_ReturnsNonNil(t *testing.T) {
	h := NewQueueHandler(&mockEnqueueJobUseCase{}, &mockGetQueueStatusUseCase{})
	if h == nil {
		t.Error("expected non-nil QueueHandler")
	}
}

func TestQueueHandler_EnqueueJob_ContentType(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{
		executeFunc: func(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
			return &entities.Job{ID: "job-test", Type: jobType}, nil
		},
	}
	statusUC := &mockGetQueueStatusUseCase{}
	h := NewQueueHandler(enqueueUC, statusUC)

	body := `{"type":"transcription"}`
	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.EnqueueJob(w, req)

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type %q, got %q", "application/json", ct)
	}
}

func TestQueueHandler_EnqueueJob_EmptyBody(t *testing.T) {
	enqueueUC := &mockEnqueueJobUseCase{}
	statusUC := &mockGetQueueStatusUseCase{}
	h := NewQueueHandler(enqueueUC, statusUC)

	req := httptest.NewRequest(http.MethodPost, "/jobs", bytes.NewBufferString(""))
	w := httptest.NewRecorder()

	h.EnqueueJob(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
