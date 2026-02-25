-- Add columns to measurements table for lab results and artifact references
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS flag TEXT;
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS reference_low NUMERIC(10,3);
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS reference_high NUMERIC(10,3);

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);
CREATE INDEX IF NOT EXISTS idx_measurements_client_type ON measurements(client_id, measurement_type);
CREATE INDEX IF NOT EXISTS idx_measurements_client_measured_at ON measurements(client_id, measured_at DESC);
