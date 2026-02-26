-- Add artifact_id reference to measurements table for linking extracted data to source PDFs
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);
