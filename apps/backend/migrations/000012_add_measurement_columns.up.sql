-- Add columns to measurements table for artifact reference, lab flags, and reference ranges.
-- These columns support the client profile API (issue #50).

DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN flag TEXT;
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN reference_low NUMERIC(10,3);
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN reference_high NUMERIC(10,3);
EXCEPTION WHEN duplicate_column THEN NULL;
END $$;

CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);
CREATE INDEX IF NOT EXISTS idx_measurements_client_type ON measurements(client_id, measurement_type);
CREATE INDEX IF NOT EXISTS idx_measurements_client_measured_at ON measurements(client_id, measured_at DESC);
