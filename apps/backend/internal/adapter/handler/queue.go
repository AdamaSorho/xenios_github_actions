package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/xenios/backend/internal/domain/entities"
)

// maxRequestBodySize limits the maximum size of request bodies to 1MB.
const maxRequestBodySize = 1 << 20 // 1MB

// EnqueueJobUseCase defines the interface for enqueuing jobs.
type EnqueueJobUseCase interface {
	Execute(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error)
}

// GetQueueStatusUseCase defines the interface for getting queue status.
type GetQueueStatusUseCaseInterface interface {
	Execute(ctx context.Context) (*entities.QueueStatus, error)
}

// QueueHandler handles HTTP requests for job queue operations.
//
// NOTE: Authentication/authorization middleware is not yet implemented.
// These endpoints are currently unauthenticated. A follow-up issue should
// add auth middleware before production deployment.
type QueueHandler struct {
	enqueueUseCase   EnqueueJobUseCase
	getStatusUseCase GetQueueStatusUseCaseInterface
}

// NewQueueHandler creates a new QueueHandler.
func NewQueueHandler(enqueueUseCase EnqueueJobUseCase, getStatusUseCase GetQueueStatusUseCaseInterface) *QueueHandler {
	return &QueueHandler{
		enqueueUseCase:   enqueueUseCase,
		getStatusUseCase: getStatusUseCase,
	}
}

// EnqueueRequest is the JSON request body for enqueuing a job.
type EnqueueRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// EnqueueJob handles POST /jobs - enqueues a new job.
func (h *QueueHandler) EnqueueJob(w http.ResponseWriter, r *http.Request) {
	// Limit request body size to prevent memory exhaustion
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)

	var req EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Type == "" {
		respondError(w, http.StatusBadRequest, "job type is required")
		return
	}

	job, err := h.enqueueUseCase.Execute(r.Context(), entities.JobType(req.Type), req.Payload)
	if err != nil {
		// Distinguish validation errors from infrastructure errors.
		// Validation errors (e.g., invalid job type) are safe to expose.
		// Infrastructure errors (e.g., database failures) should be hidden.
		if strings.HasPrefix(err.Error(), "invalid job type") {
			respondError(w, http.StatusBadRequest, err.Error())
		} else {
			respondError(w, http.StatusInternalServerError, "failed to enqueue job")
		}
		return
	}

	_ = respondJSON(w, http.StatusAccepted, job)
}

// QueueStatusResponse wraps queue status with a health indicator.
type QueueStatusResponse struct {
	Status string                `json:"status"`
	Queue  *entities.QueueStatus `json:"queue"`
}

// GetQueueStatus handles GET /jobs/status - returns queue health and counts.
func (h *QueueHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.getStatusUseCase.Execute(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get queue status")
		return
	}

	// Determine overall queue health
	healthStatus := "healthy"
	if status.Failed > 0 || status.DeadLetter > 0 {
		healthStatus = "degraded"
	}

	_ = respondJSON(w, http.StatusOK, QueueStatusResponse{
		Status: healthStatus,
		Queue:  status,
	})
}
