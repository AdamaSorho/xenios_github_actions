package entities

import (
	"encoding/json"
	"testing"
	"time"
)

func TestIsValidJobType_ValidTypes_ReturnsTrue(t *testing.T) {
	validTypes := []JobType{
		JobTypeTranscription,
		JobTypeDocumentExtraction,
		JobTypeInsightGeneration,
		JobTypeAnalyticsAggregation,
		JobTypeRiskDetection,
		JobTypeAudioCleanup,
	}

	for _, jt := range validTypes {
		if !IsValidJobType(jt) {
			t.Errorf("expected %q to be a valid job type", jt)
		}
	}
}

func TestIsValidJobType_InvalidType_ReturnsFalse(t *testing.T) {
	invalidTypes := []JobType{
		"unknown",
		"",
		"TRANSCRIPTION",
		"something_random",
	}

	for _, jt := range invalidTypes {
		if IsValidJobType(jt) {
			t.Errorf("expected %q to be an invalid job type", jt)
		}
	}
}

func TestJobStatus_IsTerminal_TerminalStatuses_ReturnsTrue(t *testing.T) {
	terminalStatuses := []JobStatus{
		JobStatusCompleted,
		JobStatusFailed,
		JobStatusExpired,
	}

	for _, s := range terminalStatuses {
		if !s.IsTerminal() {
			t.Errorf("expected %q to be a terminal status", s)
		}
	}
}

func TestJobStatus_IsTerminal_NonTerminalStatuses_ReturnsFalse(t *testing.T) {
	nonTerminalStatuses := []JobStatus{
		JobStatusCreated,
		JobStatusActive,
	}

	for _, s := range nonTerminalStatuses {
		if s.IsTerminal() {
			t.Errorf("expected %q to not be a terminal status", s)
		}
	}
}

func TestJob_ShouldRetry_UnderMaxAttempts_ReturnsTrue(t *testing.T) {
	job := &Job{
		Attempt:     0,
		MaxAttempts: MaxRetryAttempts,
	}

	if !job.ShouldRetry() {
		t.Error("expected job with 0 attempts to be retryable")
	}

	job.Attempt = 2
	if !job.ShouldRetry() {
		t.Error("expected job with 2 attempts (max 3) to be retryable")
	}
}

func TestJob_ShouldRetry_AtMaxAttempts_ReturnsFalse(t *testing.T) {
	job := &Job{
		Attempt:     MaxRetryAttempts,
		MaxAttempts: MaxRetryAttempts,
	}

	if job.ShouldRetry() {
		t.Error("expected job at max attempts to not be retryable")
	}
}

func TestJob_ShouldRetry_OverMaxAttempts_ReturnsFalse(t *testing.T) {
	job := &Job{
		Attempt:     MaxRetryAttempts + 1,
		MaxAttempts: MaxRetryAttempts,
	}

	if job.ShouldRetry() {
		t.Error("expected job over max attempts to not be retryable")
	}
}

func TestJob_NextRetryDelay_ExponentialBackoff(t *testing.T) {
	tests := []struct {
		name     string
		attempt  int
		expected time.Duration
	}{
		{
			name:     "first retry after 60s",
			attempt:  0,
			expected: 60 * time.Second,
		},
		{
			name:     "second retry after 120s",
			attempt:  1,
			expected: 120 * time.Second,
		},
		{
			name:     "third retry after 240s",
			attempt:  2,
			expected: 240 * time.Second,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			job := &Job{Attempt: tc.attempt, MaxAttempts: MaxRetryAttempts}
			got := job.NextRetryDelay()
			if got != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestJob_NextRetryDelay_AtMaxAttempts_ReturnsZero(t *testing.T) {
	job := &Job{Attempt: MaxRetryAttempts, MaxAttempts: MaxRetryAttempts}
	got := job.NextRetryDelay()
	if got != 0 {
		t.Errorf("expected 0, got %v", got)
	}
}

func TestJob_JSONSerialization(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	job := &Job{
		ID:          "test-id-123",
		Type:        JobTypeTranscription,
		Payload:     json.RawMessage(`{"file":"audio.mp3"}`),
		Status:      JobStatusCreated,
		Attempt:     0,
		MaxAttempts: 3,
		CreatedAt:   now,
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal job: %v", err)
	}

	var decoded Job
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal job: %v", err)
	}

	if decoded.ID != job.ID {
		t.Errorf("expected ID %q, got %q", job.ID, decoded.ID)
	}
	if decoded.Type != job.Type {
		t.Errorf("expected Type %q, got %q", job.Type, decoded.Type)
	}
	if decoded.Status != job.Status {
		t.Errorf("expected Status %q, got %q", job.Status, decoded.Status)
	}
	if decoded.Attempt != job.Attempt {
		t.Errorf("expected Attempt %d, got %d", job.Attempt, decoded.Attempt)
	}
	if decoded.MaxAttempts != job.MaxAttempts {
		t.Errorf("expected MaxAttempts %d, got %d", job.MaxAttempts, decoded.MaxAttempts)
	}
	if string(decoded.Payload) != string(job.Payload) {
		t.Errorf("expected Payload %s, got %s", job.Payload, decoded.Payload)
	}
}

func TestQueueStatus_JSONSerialization(t *testing.T) {
	status := QueueStatus{
		Pending:    5,
		Active:     3,
		Completed:  100,
		Failed:     2,
		Expired:    1,
		DeadLetter: 0,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("failed to marshal queue status: %v", err)
	}

	var decoded QueueStatus
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal queue status: %v", err)
	}

	if decoded != status {
		t.Errorf("expected %+v, got %+v", status, decoded)
	}
}

func TestRetryBackoffSeconds_Values(t *testing.T) {
	expected := [3]int{60, 120, 240}
	if RetryBackoffSeconds != expected {
		t.Errorf("expected retry backoff %v, got %v", expected, RetryBackoffSeconds)
	}
}

func TestMaxRetryAttempts_Value(t *testing.T) {
	if MaxRetryAttempts != 3 {
		t.Errorf("expected MaxRetryAttempts to be 3, got %d", MaxRetryAttempts)
	}
}

func TestJobTypeConstants_Values(t *testing.T) {
	tests := []struct {
		jobType  JobType
		expected string
	}{
		{JobTypeTranscription, "transcription"},
		{JobTypeDocumentExtraction, "document_extraction"},
		{JobTypeInsightGeneration, "insight_generation"},
		{JobTypeAnalyticsAggregation, "analytics_aggregation"},
		{JobTypeRiskDetection, "risk_detection"},
		{JobTypeAudioCleanup, "audio_cleanup"},
	}

	for _, tc := range tests {
		if string(tc.jobType) != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, string(tc.jobType))
		}
	}
}

func TestJobStatusConstants_Values(t *testing.T) {
	tests := []struct {
		status   JobStatus
		expected string
	}{
		{JobStatusCreated, "created"},
		{JobStatusActive, "active"},
		{JobStatusCompleted, "completed"},
		{JobStatusFailed, "failed"},
		{JobStatusExpired, "expired"},
	}

	for _, tc := range tests {
		if string(tc.status) != tc.expected {
			t.Errorf("expected %q, got %q", tc.expected, string(tc.status))
		}
	}
}
