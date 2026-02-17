-- Add artifact_id reference to measurements table for linking extracted metrics
-- back to their source InBody PDF artifact.
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);

-- Add extract_inbody to the job_type enum
DO $$ BEGIN
    ALTER TYPE job_type ADD VALUE IF NOT EXISTS 'extract_inbody';
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
