-- Add artifact_id reference to measurements table for linking extracted
-- measurements back to their source document (e.g., InBody PDF scans).
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);

-- Add partial_extraction flag to indicate when not all expected fields
-- could be extracted from the source document.
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS partial_extraction BOOLEAN NOT NULL DEFAULT FALSE;
