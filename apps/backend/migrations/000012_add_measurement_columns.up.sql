-- Add columns to measurements table for artifact linking and lab reference ranges
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS flag TEXT;
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS reference_low NUMERIC(10,3);
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS reference_high NUMERIC(10,3);

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);
CREATE INDEX IF NOT EXISTS idx_measurements_client_type_date ON measurements(client_id, measurement_type, measured_at DESC);
