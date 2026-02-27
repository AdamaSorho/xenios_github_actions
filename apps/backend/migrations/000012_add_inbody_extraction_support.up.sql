-- Add extract_inbody to job_type enum
DO $$ BEGIN
    ALTER TYPE job_type ADD VALUE 'extract_inbody';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Add artifact_id column to measurements for linking extracted data back to source PDFs
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);

-- Expand artifact status to include 'processed'
-- Drop existing inline CHECK constraint (auto-named by PostgreSQL)
ALTER TABLE artifacts DROP CONSTRAINT IF EXISTS artifacts_status_check;

DO $$ BEGIN
    ALTER TABLE artifacts ADD CONSTRAINT artifacts_status_check
        CHECK (status IN ('pending', 'uploaded', 'failed', 'processed'));
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
