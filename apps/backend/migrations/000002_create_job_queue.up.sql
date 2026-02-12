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
    retry_after TIMESTAMPTZ,
    locked_by UUID,
    locked_at TIMESTAMPTZ
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

-- Events audit table for job lifecycle tracking
CREATE TABLE IF NOT EXISTS events_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id UUID NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index for audit trail queries
CREATE INDEX IF NOT EXISTS idx_events_audit_entity ON events_audit (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_events_audit_type ON events_audit (event_type);

-- Enable Row Level Security
ALTER TABLE events_audit ENABLE ROW LEVEL SECURITY;
