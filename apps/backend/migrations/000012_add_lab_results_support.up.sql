-- Add extract_lab_results to job_type enum
DO $$ BEGIN
    ALTER TYPE job_type ADD VALUE 'extract_lab_results';
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Add artifact_id column to measurements table (links measurement to source artifact)
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN artifact_id UUID REFERENCES artifacts(id) ON DELETE SET NULL;
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

-- Add reference_low column to measurements table
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN reference_low NUMERIC(10,3);
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

-- Add reference_high column to measurements table
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN reference_high NUMERIC(10,3);
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

-- Add flag column to measurements table
DO $$ BEGIN
    ALTER TABLE measurements ADD COLUMN flag TEXT CHECK (flag IN ('normal', 'low', 'high', 'critical_low', 'critical_high'));
EXCEPTION
    WHEN duplicate_column THEN null;
END $$;

-- Index for querying measurements by artifact
CREATE INDEX IF NOT EXISTS idx_measurements_artifact_id ON measurements(artifact_id);

-- Index for querying flagged measurements
CREATE INDEX IF NOT EXISTS idx_measurements_flag ON measurements(flag)
    WHERE flag IS NOT NULL AND flag != 'normal';
