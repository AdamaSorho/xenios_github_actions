-- Add artifact_id reference to measurements table for linking extracted data
-- back to source artifacts (e.g., InBody PDF scans).

DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);

-- Add extraction_status to track partial vs full extraction results.
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN extraction_status TEXT NOT NULL DEFAULT 'complete'
        CHECK (extraction_status IN ('complete', 'partial'));
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;
