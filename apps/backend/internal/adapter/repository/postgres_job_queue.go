package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xenios/backend/internal/domain/entities"
	"github.com/xenios/backend/internal/domain/repository"
)

// PostgresJobQueue implements the JobQueue interface using PostgreSQL.
// It uses FOR UPDATE SKIP LOCKED for safe job dequeuing and JSONB for payloads.
type PostgresJobQueue struct {
	db *pgxpool.Pool
}

// NewPostgresJobQueue creates a new PostgresJobQueue instance.
func NewPostgresJobQueue(db *pgxpool.Pool) repository.JobQueue {
	return &PostgresJobQueue{db: db}
}

// Enqueue adds a new job to the queue.
func (q *PostgresJobQueue) Enqueue(ctx context.Context, jobType entities.JobType, payload []byte) (*entities.Job, error) {
	var job entities.Job
	var rawPayload []byte

	err := q.db.QueryRow(ctx,
		`INSERT INTO jobs (type, payload, status, attempt, max_attempts, created_at)
		 VALUES ($1, $2, $3, 0, $4, now())
		 RETURNING id, type, payload, status, attempt, max_attempts, created_at`,
		string(jobType), payload, string(entities.JobStatusCreated), entities.MaxRetryAttempts,
	).Scan(
		&job.ID, &job.Type, &rawPayload, &job.Status,
		&job.Attempt, &job.MaxAttempts, &job.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert job: %w", err)
	}
	job.Payload = json.RawMessage(rawPayload)

	return &job, nil
}

// Dequeue fetches and locks the next available job using FOR UPDATE SKIP LOCKED.
// This provides safe concurrent access without advisory locks overhead.
// Returns nil if no jobs are available.
func (q *PostgresJobQueue) Dequeue(ctx context.Context, jobTypes []entities.JobType) (*entities.Job, error) {
	if len(jobTypes) == 0 {
		return nil, nil
	}

	// Convert job types to strings for the query
	types := make([]string, len(jobTypes))
	for i, jt := range jobTypes {
		types[i] = string(jt)
	}

	tx, err := q.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var job entities.Job
	var rawPayload []byte

	// Use FOR UPDATE SKIP LOCKED to safely dequeue without blocking other workers.
	// Select jobs that are either newly created or failed jobs past their retry_after time.
	err = tx.QueryRow(ctx,
		`SELECT id, type, payload, status, attempt, max_attempts, created_at,
		        started_at, completed_at, failed_at, retry_after, error_msg
		 FROM jobs
		 WHERE type = ANY($1::job_type[])
		   AND (
		       (status = 'created')
		       OR (status = 'failed' AND attempt < max_attempts AND retry_after <= now())
		   )
		 ORDER BY created_at ASC
		 LIMIT 1
		 FOR UPDATE SKIP LOCKED`,
		types,
	).Scan(
		&job.ID, &job.Type, &rawPayload, &job.Status,
		&job.Attempt, &job.MaxAttempts, &job.CreatedAt,
		&job.StartedAt, &job.CompletedAt, &job.FailedAt,
		&job.RetryAfter, &job.ErrorMsg,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select job: %w", err)
	}
	job.Payload = json.RawMessage(rawPayload)

	// Mark job as active
	now := time.Now()
	job.Status = entities.JobStatusActive
	job.StartedAt = &now
	job.Attempt++

	_, err = tx.Exec(ctx,
		`UPDATE jobs SET status = $1, started_at = $2, attempt = $3 WHERE id = $4`,
		string(entities.JobStatusActive), now, job.Attempt, job.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("update job status: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return &job, nil
}

// Complete marks a job as successfully completed.
func (q *PostgresJobQueue) Complete(ctx context.Context, jobID string) error {
	now := time.Now()
	result, err := q.db.Exec(ctx,
		`UPDATE jobs SET status = $1, completed_at = $2 WHERE id = $3 AND status = $4`,
		string(entities.JobStatusCompleted), now, jobID, string(entities.JobStatusActive),
	)
	if err != nil {
		return fmt.Errorf("update job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("job %s not found or not in active state", jobID)
	}

	return nil
}

// Fail marks a job as failed. If retries remain, schedules a retry with exponential backoff.
// If max retries exceeded, moves to dead letter queue.
func (q *PostgresJobQueue) Fail(ctx context.Context, jobID string, errMsg string) error {
	tx, err := q.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Fetch current job state
	var job entities.Job
	var rawPayload []byte
	err = tx.QueryRow(ctx,
		`SELECT id, type, payload, status, attempt, max_attempts, created_at
		 FROM jobs WHERE id = $1 FOR UPDATE`,
		jobID,
	).Scan(&job.ID, &job.Type, &rawPayload, &job.Status,
		&job.Attempt, &job.MaxAttempts, &job.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("job %s not found", jobID)
		}
		return fmt.Errorf("select job: %w", err)
	}
	job.Payload = json.RawMessage(rawPayload)

	now := time.Now()

	if job.ShouldRetry() {
		// Schedule retry with exponential backoff
		retryAfter := now.Add(job.NextRetryDelay())
		_, err = tx.Exec(ctx,
			`UPDATE jobs SET status = $1, failed_at = $2, error_msg = $3, retry_after = $4
			 WHERE id = $5`,
			string(entities.JobStatusFailed), now, errMsg, retryAfter, jobID,
		)
		if err != nil {
			return fmt.Errorf("update job for retry: %w", err)
		}
	} else {
		// Move to dead letter queue
		_, err = tx.Exec(ctx,
			`INSERT INTO jobs_dead_letter (original_job_id, type, payload, attempts, last_error, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			job.ID, string(job.Type), rawPayload, job.Attempt, errMsg, job.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("insert dead letter: %w", err)
		}

		// Update job to failed (terminal)
		_, err = tx.Exec(ctx,
			`UPDATE jobs SET status = $1, failed_at = $2, error_msg = $3 WHERE id = $4`,
			string(entities.JobStatusFailed), now, errMsg, jobID,
		)
		if err != nil {
			return fmt.Errorf("update job as permanently failed: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// GetStatus returns aggregate counts of jobs by status.
func (q *PostgresJobQueue) GetStatus(ctx context.Context) (*entities.QueueStatus, error) {
	var status entities.QueueStatus

	err := q.db.QueryRow(ctx,
		`SELECT
			COALESCE(SUM(CASE WHEN status = 'created' THEN 1 ELSE 0 END), 0) AS pending,
			COALESCE(SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END), 0) AS active,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END), 0) AS completed,
			COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0) AS failed,
			COALESCE(SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END), 0) AS expired
		 FROM jobs`,
	).Scan(&status.Pending, &status.Active, &status.Completed, &status.Failed, &status.Expired)
	if err != nil {
		return nil, fmt.Errorf("query job counts: %w", err)
	}

	// Count dead letter entries separately
	err = q.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM jobs_dead_letter`,
	).Scan(&status.DeadLetter)
	if err != nil {
		return nil, fmt.Errorf("query dead letter count: %w", err)
	}

	return &status, nil
}

// GetJob returns a single job by ID.
func (q *PostgresJobQueue) GetJob(ctx context.Context, jobID string) (*entities.Job, error) {
	var job entities.Job
	var rawPayload []byte

	err := q.db.QueryRow(ctx,
		`SELECT id, type, payload, status, attempt, max_attempts, created_at,
		        started_at, completed_at, failed_at, retry_after, error_msg
		 FROM jobs WHERE id = $1`,
		jobID,
	).Scan(
		&job.ID, &job.Type, &rawPayload, &job.Status,
		&job.Attempt, &job.MaxAttempts, &job.CreatedAt,
		&job.StartedAt, &job.CompletedAt, &job.FailedAt,
		&job.RetryAfter, &job.ErrorMsg,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("job %s not found", jobID)
		}
		return nil, fmt.Errorf("select job: %w", err)
	}
	job.Payload = json.RawMessage(rawPayload)

	return &job, nil
}
