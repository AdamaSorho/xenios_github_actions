package entities

import (
	"encoding/json"
	"fmt"
	"time"
)

// ValidationError represents a domain validation error.
// Handlers can use errors.As to distinguish validation errors from internal errors.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// NewValidationError creates a new ValidationError with a formatted message.
func NewValidationError(format string, args ...interface{}) *ValidationError {
	return &ValidationError{Message: fmt.Sprintf(format, args...)}
}

// JobType represents the type of background job.
type JobType string

const (
	JobTypeTranscription        JobType = "transcription"
	JobTypeDocumentExtraction   JobType = "document_extraction"
	JobTypeInsightGeneration    JobType = "insight_generation"
	JobTypeAnalyticsAggregation JobType = "analytics_aggregation"
	JobTypeRiskDetection        JobType = "risk_detection"
	JobTypeAudioCleanup         JobType = "audio_cleanup"
	JobTypeExtractInBody        JobType = "extract_inbody"
	JobTypeExtractLabResults    JobType = "extract_lab_results"
	JobTypeExtractWearable      JobType = "extract_wearable"
	JobTypeExtractNutrition     JobType = "extract_nutrition"
	JobTypeTranscribeAudio      JobType = "transcribe_audio"
	JobTypeClassifyDocument     JobType = "classify_document"
)

// JobStatus represents the current status of a job.
type JobStatus string

const (
	JobStatusCreated   JobStatus = "created"
	JobStatusActive    JobStatus = "active"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusExpired   JobStatus = "expired"
)

// MaxRetryAttempts is the maximum number of retry attempts before a job is moved to the dead letter queue.
const MaxRetryAttempts = 3

// RetryBackoffSeconds defines exponential backoff intervals in seconds for each retry attempt.
var RetryBackoffSeconds = [MaxRetryAttempts]int{60, 120, 240}

// Job represents a background job in the queue.
type Job struct {
	ID          string          `json:"id"`
	Type        JobType         `json:"type"`
	Payload     json.RawMessage `json:"payload"`
	Status      JobStatus       `json:"status"`
	Attempt     int             `json:"attempt"`
	MaxAttempts int             `json:"max_attempts"`
	CreatedAt   time.Time       `json:"created_at"`
	StartedAt   *time.Time      `json:"started_at,omitempty"`
	CompletedAt *time.Time      `json:"completed_at,omitempty"`
	FailedAt    *time.Time      `json:"failed_at,omitempty"`
	RetryAfter  *time.Time      `json:"retry_after,omitempty"`
	ErrorMsg    *string         `json:"error_msg,omitempty"`
}

// QueueStatus represents aggregate status counts for the job queue.
type QueueStatus struct {
	Pending   int `json:"pending"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Expired   int `json:"expired"`
	DeadLetter int `json:"dead_letter"`
}

// IsValidJobType returns true if the given job type is one of the known types.
func IsValidJobType(jt JobType) bool {
	switch jt {
	case JobTypeTranscription,
		JobTypeDocumentExtraction,
		JobTypeInsightGeneration,
		JobTypeAnalyticsAggregation,
		JobTypeRiskDetection,
		JobTypeAudioCleanup,
		JobTypeExtractInBody,
		JobTypeExtractLabResults,
		JobTypeExtractWearable,
		JobTypeExtractNutrition,
		JobTypeTranscribeAudio,
		JobTypeClassifyDocument:
		return true
	}
	return false
}

// IsTerminal returns true if the job status is a terminal state (completed, failed, expired).
func (s JobStatus) IsTerminal() bool {
	switch s {
	case JobStatusCompleted, JobStatusFailed, JobStatusExpired:
		return true
	}
	return false
}

// ShouldRetry returns true if the job should be retried based on its current attempt count.
func (j *Job) ShouldRetry() bool {
	return j.Attempt < j.MaxAttempts
}

// NextRetryDelay returns the backoff duration for the next retry attempt.
// Returns 0 if no more retries are available.
func (j *Job) NextRetryDelay() time.Duration {
	if j.Attempt >= MaxRetryAttempts {
		return 0
	}
	idx := j.Attempt
	if idx >= len(RetryBackoffSeconds) {
		idx = len(RetryBackoffSeconds) - 1
	}
	return time.Duration(RetryBackoffSeconds[idx]) * time.Second
}
