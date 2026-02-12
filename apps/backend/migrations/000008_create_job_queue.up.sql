-- Create job_type enum
DO $$ BEGIN
    CREATE TYPE job_type AS ENUM (
        'transcription',
        'document_extraction',
        'insight_generation',
        'analytics_aggregation',
        'risk_detection',
        'audio_cleanup'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create job_status enum
DO $$ BEGIN
    CREATE TYPE job_status AS ENUM (
        'created',
        'active',
        'completed',
        'failed',
        'expired'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type job_type NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status job_status NOT NULL DEFAULT 'created',
    attempt INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    error_msg TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    failed_at TIMESTAMPTZ,
    retry_after TIMESTAMPTZ
);

-- Index for worker polling: find available jobs efficiently
CREATE INDEX IF NOT EXISTS idx_jobs_status_retry ON jobs (status, retry_after)
    WHERE status IN ('created', 'failed');

-- Index for status aggregation queries
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status);

-- Index for job type filtering
CREATE INDEX IF NOT EXISTS idx_jobs_type ON jobs (type);

-- Enable Row Level Security
ALTER TABLE jobs ENABLE ROW LEVEL SECURITY;

-- RLS policy: restrict job queue access to the service role only.
-- The backend application connects as the service role; other roles (e.g., anon,
-- authenticated) should not directly access the jobs table.
DROP POLICY IF EXISTS jobs_service_all ON jobs;
CREATE POLICY jobs_service_all ON jobs
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);

-- Dead letter queue table for permanently failed jobs
CREATE TABLE IF NOT EXISTS jobs_dead_letter (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    original_job_id UUID NOT NULL,
    type job_type NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    attempts INTEGER NOT NULL,
    last_error TEXT,
    failed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL
);

-- Index for dead letter queue lookups
CREATE INDEX IF NOT EXISTS idx_jobs_dead_letter_type ON jobs_dead_letter (type);
CREATE INDEX IF NOT EXISTS idx_jobs_dead_letter_failed_at ON jobs_dead_letter (failed_at);

-- Enable Row Level Security
ALTER TABLE jobs_dead_letter ENABLE ROW LEVEL SECURITY;

-- RLS policy: restrict dead letter queue access to the service role only.
DROP POLICY IF EXISTS jobs_dead_letter_service_all ON jobs_dead_letter;
CREATE POLICY jobs_dead_letter_service_all ON jobs_dead_letter
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);

-- Job events table for job lifecycle audit trail.
-- Separate from events_audit (migration 000005) which requires actor_id (user context).
-- Job events are system-generated and have no user actor.
CREATE TABLE IF NOT EXISTS job_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL,
    event_type TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_job_events_job_id ON job_events (job_id);
CREATE INDEX IF NOT EXISTS idx_job_events_event_type ON job_events (event_type);
CREATE INDEX IF NOT EXISTS idx_job_events_created_at ON job_events (created_at);

-- Enable Row Level Security
ALTER TABLE job_events ENABLE ROW LEVEL SECURITY;

-- RLS policy: restrict job events access to the service role only.
DROP POLICY IF EXISTS job_events_service_all ON job_events;
CREATE POLICY job_events_service_all ON job_events
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);
