-- Add lab results reference range and flag columns to measurements table
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS reference_low NUMERIC(10,3);
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS reference_high NUMERIC(10,3);
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS flag TEXT CHECK (flag IN ('normal', 'low', 'high', 'critical_low', 'critical_high'));
ALTER TABLE measurements ADD COLUMN IF NOT EXISTS artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);
CREATE INDEX IF NOT EXISTS idx_measurements_flag ON measurements(flag);
