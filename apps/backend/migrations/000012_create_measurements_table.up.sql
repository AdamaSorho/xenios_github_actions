-- Create wearable_source enum type
DO $$ BEGIN
    CREATE TYPE wearable_source AS ENUM ('whoop', 'garmin', 'apple_health', 'oura', 'fitbit');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Create measurement_type enum type
DO $$ BEGIN
    CREATE TYPE measurement_type AS ENUM (
        'hrv_ms',
        'sleep_duration_hrs',
        'recovery_score',
        'strain_score',
        'resting_hr_bpm',
        'steps_count',
        'sleep_quality_score'
    );
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Create measurements table
CREATE TABLE IF NOT EXISTS measurements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID NOT NULL,
    source wearable_source NOT NULL,
    measurement_type measurement_type NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    measured_at DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- Prevent duplicate entries for the same client/source/type/date
    UNIQUE (client_id, source, measurement_type, measured_at)
);

-- Indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_measurements_client_source
    ON measurements (client_id, source);

CREATE INDEX IF NOT EXISTS idx_measurements_client_source_type_date
    ON measurements (client_id, source, measurement_type, measured_at DESC);

CREATE INDEX IF NOT EXISTS idx_measurements_measured_at
    ON measurements (measured_at);

-- Enable Row Level Security
ALTER TABLE measurements ENABLE ROW LEVEL SECURITY;

-- Service role bypass for backend operations
DROP POLICY IF EXISTS measurements_service_role ON measurements;
CREATE POLICY measurements_service_role ON measurements
    FOR ALL
    TO service_role
    USING (true)
    WITH CHECK (true);
